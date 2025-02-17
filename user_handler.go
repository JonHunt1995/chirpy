package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"chirpy.com/internal/auth"
	"chirpy.com/internal/database"
)

func (cfg *apiConfig) userHandler(w http.ResponseWriter, r *http.Request) {
	type Parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := Parameters{}
	err := decoder.Decode(&params)
	// Respond with Error if problems marshalling JSON
	if err != nil {
		msg := fmt.Sprintf("Error marshalling JSON: %s", err)
		cfg.respondWithError(w, 500, msg)
		return
	}
	hash, err := auth.HashPassword(params.Password)
	if err != nil {
		cfg.respondWithError(w, 500, "Failed to hash password")
		return
	}
	dbUser, err := cfg.queries.CreateUser(r.Context(), database.CreateUserParams{
		Email:          params.Email,
		HashedPassword: hash,
	})

	if err != nil {
		cfg.respondWithError(w, 500, "Failed to create user")
		return
	}

	user := User{
		ID:        dbUser.ID,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		Email:     dbUser.Email,
	}
	cfg.respondWithJSON(w, 201, user)
	return
}
