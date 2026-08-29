package oauth

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
)

func GenerateCode() (string, error) {
	b := make([]byte, 32)

	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}

func SendError(w http.ResponseWriter, errorMessage string, code int) {
	reason := ErrorReason{Reason: errorMessage}
	reasonRaw, _ := json.Marshal(reason)
	w.WriteHeader(code)
	w.Write(reasonRaw)
}
