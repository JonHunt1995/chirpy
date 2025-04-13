package auth

import (
	"fmt"
	"net/http"
	"strings"
)

func GetBearerToken(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("no authorization header found")
	}
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", fmt.Errorf("auth header has improper formatting")
	}

	token := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
	if len(token) < 1 {
		return "", fmt.Errorf("no token found")
	}
	return token, nil
}


