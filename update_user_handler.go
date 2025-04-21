package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"chirpy.com/internal/auth"
	"chirpy.com/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) updateUserHandler(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		cfg.respondWithError(w, 401, "token not found")
		return
	}
	type Parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	// Getting UserID
	userID, err := auth.ValidateJWT(token, os.Getenv("SECRET"))
	if err != nil {
		cfg.respondWithError(w, 401, "token not valid")
		return
	}

	// Getting Email and Password
	decoder := json.NewDecoder(r.Body)
	params := Parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		msg := fmt.Sprintf("Error marshalling JSON: %s", err)
		cfg.respondWithError(w, 500, msg)
		return
	}
	// Getting Hashed Password
	hash, err := auth.HashPassword(params.Password)
	if err != nil {
		cfg.respondWithError(w, 500, "Failed to hash password")
		return
	}
	// Update User Email And Password
	user, err := cfg.queries.UpdateUser(r.Context(), database.UpdateUserParams{
		ID:             userID,
		Email:          params.Email,
		HashedPassword: hash,
	})
	if err != nil {
		cfg.respondWithError(w, 500, "Failed to update user")
		return
	}
	// Send user data but omit password
	cfg.respondWithJSON(w, 200, struct {
		ID    uuid.UUID `json:"id"`
		Email string    `json:"email"`
	}{
		ID:    user.ID,
		Email: user.Email,
	})
}
