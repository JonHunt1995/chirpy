package main

import (
	"net/http"
	"time"

	//"strings"

	"chirpy.com/internal/auth"
	//"chirpy.com/internal/database"
)

func (cfg *apiConfig) refreshHandler(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		cfg.respondWithError(w, 401, "unable to make access token")
		return
	}

	refreshToken, err := cfg.queries.GetRefreshToken(r.Context(), token)
	if err != nil {
		cfg.respondWithError(w, 401, "refresh token not found in database")
		return
	}

	if refreshToken.ExpiresAt.Before(time.Now()) {
		cfg.respondWithError(w, 401, "expired refresh token")
		return
	}

	if refreshToken.RevokedAt.Valid {
		cfg.respondWithError(w, 401, "revoked refresh token")
		return
	}

	if !refreshToken.UserID.Valid {
		cfg.respondWithError(w, 401, "invalid user id in refresh token")
		return
	}

	accessToken, err := auth.MakeJWT(refreshToken.UserID.UUID, cfg.jwtSecret, time.Hour)
	if err != nil {
		cfg.respondWithError(w, 401, "Unable to make access token")
		return
	}

	cfg.respondWithJSON(w, 200, struct {
		Token string `json:"token"`
	}{
		Token: accessToken,
	})
}
