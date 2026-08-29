package oauth

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

func CreateOAuthClient(rdb *redis.Client, config StudtrainOAuthConfig) StudtrainOAuth {
	return StudtrainOAuth{
		rdb:    rdb,
		config: &config,
	}
}

func (serv *StudtrainOAuth) OAuthSignInHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		codeVerifier, _ := GenerateCode()
		state, _ := GenerateCode()
		redirectOriginUrl := r.URL.Query().Get("next")

		pkce := PKCE{
			CodeVerifier: codeVerifier,
			RedirectUrl:  redirectOriginUrl,
		}

		pkceRaw, _ := json.Marshal(pkce)
		err := serv.rdb.Set(context.Background(), "pkce:"+state, pkceRaw, time.Minute*15).Err()
		if err != nil {
			SendError(w, err.Error(), 500)
			return
		}

		codeChallengeHash := sha256.Sum256([]byte(codeVerifier))
		codeChallenge := base64.RawURLEncoding.EncodeToString(codeChallengeHash[:])

		redirectUrl := fmt.Sprintf(
			"%s/auth?service=%s&next=%s&code_challenge=%s&state=%s",
			serv.config.OAuthOrigin,
			serv.config.ServiceId,
			serv.config.CallbackUrl,
			codeChallenge,
			state,
		)
		w.Header().Add("Location", redirectUrl)
		w.WriteHeader(302)
	})
}

func (serv *StudtrainOAuth) OAuthCallbackHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		state := r.URL.Query().Get("state")
		authorizationCode := r.URL.Query().Get("authorization_code")

		pkceRaw, err := serv.rdb.Get(context.Background(), "pkce:"+state).Result()
		if err != nil {
			SendError(w, "Authorization error: Invalid PKCE (NO)", 500)
			return
		}

		pkce := PKCE{}
		if err := json.Unmarshal([]byte(pkceRaw), &pkce); err != nil {
			SendError(w, "Authorization error: Invalid PKCE (FORMAT)", 500)
			return
		}

		_, err = serv.rdb.Del(context.Background(), "pkce:"+state).Result()
		if err != nil {
			SendError(w, "Authorization error: Invalid PKCE (REDIS)", 500)
			return
		}

		tokenGetUrl := fmt.Sprintf(
			"%s/api/v1/auth/token?authorization_code=%s&code_verifier=%s",
			serv.config.OAuthOrigin,
			authorizationCode,
			pkce.CodeVerifier,
		)

		response, err := http.Get(tokenGetUrl)
		if err != nil {
			SendError(w, "Authorization error: Unknown error (HTTP)", 500)
			return
		}
		defer response.Body.Close()

		tokensRaw, err := io.ReadAll(response.Body)
		for true {
			bs := make([]byte, 1014)
			n, err := response.Body.Read(bs)
			tokensRaw = append(tokensRaw, bs[:n]...)
			if n == 0 || err != nil {
				break
			}
		}

		tokens := TokenPair{}
		if err = json.Unmarshal(tokensRaw, &tokens); err != nil {
			SendError(w, "Authorization error: Tokens error (FORMAT)", 500)
			return
		}

		refreshCookie := fmt.Sprintf("refresh_token=%s; Path=/; HttpOnly; SameSite=Lax; Max-Age=2592000", tokens.RefreshToken)
		accessCookie := fmt.Sprintf("access_token=%s; Path=/api/v1/oauth; HttpOnly; SameSite=Lax; Max-Age=600", tokens.AccessToken)

		w.Header().Add("Set-Cookie", refreshCookie)
		w.Header().Add("Set-Cookie", accessCookie)
		w.Header().Add("Location", pkce.RedirectUrl)
		w.WriteHeader(302)
	})
}

func (serv *StudtrainOAuth) OAuthGetProfileHandler(convert func(result UserProfileDto) any) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accessToken, err := r.Cookie("access_token")
		if err != nil {
			SendError(w, "Authorization error: No access token", 401)
			return
		}

		requestBody, err := json.Marshal(AccessToken{
			AccessToken: accessToken.Value,
		})
		if err != nil {
			SendError(w, "Authorization error: No access token", 401)
			return
		}

		profileGetUrl := fmt.Sprintf("%s/api/v1/auth/profile", serv.config.OAuthOrigin)
		response, err := http.Post(profileGetUrl, "application/json", strings.NewReader(string(requestBody)))
		if err != nil {
			SendError(w, "Authorization error: User fetch error (HTTP)", 401)
			return
		}
		defer response.Body.Close()

		if response.StatusCode != 200 {
			SendError(w, "Authorization error: User fetch error", 401)
			return
		}

		profileRaw, err := io.ReadAll(response.Body)
		for true {
			bs := make([]byte, 1014)
			n, err := response.Body.Read(bs)
			profileRaw = append(profileRaw, bs[:n]...)
			if n == 0 || err != nil {
				break
			}
		}

		profileData := UserProfileDto{}
		if err := json.Unmarshal(profileRaw, &profileData); err != nil {
			SendError(w, "Authorization error: Invalid user profile", 401)
			return
		}

		reasonRaw, err := json.Marshal(convert(profileData))
		if err := json.Unmarshal(profileRaw, &profileData); err != nil {
			SendError(w, "Authorization error: "+err.Error(), 500)
			return
		}

		w.Write(reasonRaw)
	})
}
