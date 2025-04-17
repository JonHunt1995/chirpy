package main

import (
	"database/sql"
	"net/http"
	"time"

	"chirpy.com/internal/auth"
	"chirpy.com/internal/database"
)

func (cfg *apiConfig) revokeHandler(w http.ResponseWriter, r *http.Request) {
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

	err = cfg.queries.RevokeRefreshToken(r.Context(), database.RevokeRefreshTokenParams{
		RevokedAt: sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		},
		UpdatedAt: time.Now(),
		Token:     refreshToken.Token,
	})

	if err != nil {
		cfg.respondWithError(w, 401, "Unable to revoke token")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
