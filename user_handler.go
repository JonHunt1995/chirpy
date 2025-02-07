package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (cfg *apiConfig) userHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	params := Email{}
	err := decoder.Decode(&params)
	// Respond with Error if problems marshalling JSON
	if err != nil {
		msg := fmt.Sprintf("Error marshalling JSON: %s", err)
		cfg.respondWithError(w, 500, msg)
		return
	}
	dbUser, err := cfg.queries.CreateUser(r.Context(), params.Email)
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
