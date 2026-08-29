package oauth

import "github.com/redis/go-redis/v9"

type StudtrainOAuthConfig struct {
	ServiceId   string
	CallbackUrl string
	OAuthOrigin string
}

type PKCE struct {
	CodeVerifier string `json:"code_verifier"`
	RedirectUrl  string `json:"redirect_url"`
}

type TokenPair struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type ErrorReason struct {
	Reason string `json:"error"`
}

type AccessToken struct {
	AccessToken string `json:"accessToken"`
}

type UserProfileDto struct {
	Login string            `json:"login"`
	Info  map[string]string `json:"info"`
}

type StudtrainOAuth struct {
	rdb    *redis.Client
	config *StudtrainOAuthConfig
}
