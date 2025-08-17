package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"chirpy.com/internal/auth"
	"chirpy.com/internal/database"
)

func (cfg *apiConfig) loginHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	decoder := json.NewDecoder(r.Body)
	req := parameters{}
	err := decoder.Decode(&req)
	if err != nil {
		msg := fmt.Sprintf("Error marshalling JSON: %s", err)
		cfg.respondWithError(w, 500, msg)
		return
	}

	dbUser, err := cfg.queries.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		cfg.respondWithError(w, 401, "Incorrect email or password")
		return
	}

	err = auth.CheckPasswordHash(req.Password, dbUser.HashedPassword)
	if err != nil {
		cfg.respondWithError(w, 401, "Incorrect email or password")
		return
	}
	expireTime := 3600

	token, err := auth.MakeJWT(dbUser.ID, os.Getenv("SECRET"), time.Second*time.Duration(expireTime))
	if err != nil {
		cfg.respondWithError(w, 500, "Unable to make Access Token")
		return
	}
	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		cfg.respondWithError(w, 500, "Unable to make Refresh Token")
		return
	}

	user := User{
		ID:           dbUser.ID,
		CreatedAt:    dbUser.CreatedAt,
		UpdatedAt:    dbUser.UpdatedAt,
		Email:        dbUser.Email,
		Token:        token,
		RefreshToken: refreshToken,
		IsChirpyRed: dbUser.IsChirpyRed,
	}

	err = cfg.queries.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:     refreshToken,
		UserID:    user.ID,
		ExpiresAt: time.Now().AddDate(0, 0, 60),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		
	})

	if err != nil {
		cfg.respondWithError(w, 500, "Unable to add refresh token to database")
		return
	}

	cfg.respondWithJSON(w, 200, user)
}
