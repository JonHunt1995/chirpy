package main

import (
	"net/http"
	"os"

	"chirpy.com/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) deleteChirpHandler(w http.ResponseWriter, r *http.Request) {
	// Get token to extract User ID
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		cfg.respondWithError(w, 401, "token not found")
		return
	}
	userID, err := auth.ValidateJWT(token, os.Getenv("SECRET"))
	if err != nil {
		cfg.respondWithError(w, 401, "token not valid")
		return
	}
	// Get Chirp ID from URL
	chirpIDFromURL := r.PathValue("chirpID")
	if chirpIDFromURL == "" {
		cfg.respondWithError(w, 404, "Chirp ID is blank")
		return
	}
	chirpID, err := uuid.Parse(chirpIDFromURL)
	if err != nil {
		cfg.respondWithError(w, 404, "chirp not found")
		return
	}

	// Check to see if chirp author is same as user ID from header
	chirpAuthor, err := cfg.queries.GetChirpAuthor(r.Context(), chirpID)
	if err != nil {
		cfg.respondWithError(w, 404, "chirp not found")
		return
	}
	if userID != chirpAuthor.UUID {
		cfg.respondWithError(w, 403, "Action not allowed")
		return
	}
	// Delete Chirp
	err = cfg.queries.DeleteChirp(r.Context(), chirpID)
	if err != nil {
		cfg.respondWithError(w, 404, "chirp not found")
		return
	}
	w.WriteHeader(204)
}
