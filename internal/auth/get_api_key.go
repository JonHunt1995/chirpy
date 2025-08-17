package auth

import (
	"errors"
	"net/http"
	"strings"
)

func GetAPIKey(headers http.Header) (string, error) {
	authHeader := strings.Split(headers.Get("Authorization"), " ")
	if len(authHeader) < 2 {
		return "", errors.New("Issue with Authorization header")
	}
	apiKey := authHeader[1]
	if len(apiKey) == 0 {
		return "", errors.New("apiKey not found")
	}
	return apiKey, nil
}