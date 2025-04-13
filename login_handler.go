package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"chirpy.com/internal/auth"
)

func (cfg *apiConfig) loginHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email     string `json:"email"`
		Password  string `json:"password"`
		ExpiresAt int    `json:"expires_in_seconds"`
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
	if req.ExpiresAt > 0 && req.ExpiresAt <= 3600 {
		expireTime = req.ExpiresAt
	}
	token, err := auth.MakeJWT(dbUser.ID, os.Getenv("SECRET"), time.Second * time.Duration(expireTime))
	if err != nil {
		cfg.respondWithError(w, 500, "Unable to make JWT")
	}
	user := User{
		ID:        dbUser.ID,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		Email:     dbUser.Email,
		Token: token,
	}
	cfg.respondWithJSON(w, 200, user)
}
