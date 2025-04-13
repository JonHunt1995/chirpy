package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"chirpy.com/internal/auth"
	"chirpy.com/internal/database"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

func (cfg *apiConfig) chirpsHandler(w http.ResponseWriter, r *http.Request) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	// Ensure That User Has Valid JWT
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		cfg.respondWithError(w, http.StatusUnauthorized, "Missing or invalid token")
	}
	userID, err := auth.ValidateJWT(token, os.Getenv("SECRET"))
	if err != nil {
		cfg.respondWithError(w, http.StatusUnauthorized, "Token is invalid")
	}

	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	params := CreateChirpRequest{}
	err = decoder.Decode(&params)
	// Respond with Error if problems marshalling JSON
	if err != nil {
		msg := fmt.Sprintf("Error marshalling JSON: %s", err)
		cfg.respondWithError(w, 500, msg)
		return
	}
	// Respond with Error if message is > 140 characters
	if len(params.Body) > 140 {
		statusCode := 400
		msg := "Chirp is too long"
		cfg.respondWithError(w, statusCode, msg)
		return
	}
	// Use removeProfanity to clean the Chirp Body
	cleanedBody := removeProfanity(params.Body)
	// Chirp is valid if past this point
	chirp, err := cfg.queries.CreateChirp(r.Context(), database.CreateChirpParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Body:      cleanedBody,
		UserID: uuid.NullUUID{
			UUID:  userID, // from JWT validation
			Valid: true,
	},
	})
	if err != nil {
		cfg.respondWithError(w, 500, "Failed to create chirp")
		return
	}
	chirpResponse := Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID.UUID,
	}
	cfg.respondWithJSON(w, 201, chirpResponse)
	return
}
