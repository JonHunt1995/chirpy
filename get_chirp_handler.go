package main

import (
	"net/http"

	"github.com/google/uuid"
)

func (cfg *apiConfig) getChirpHandler(w http.ResponseWriter, r *http.Request) {
	// Get the chirp id path
	chirpID := r.PathValue("chirpID")
	// Parse string into a unique user id
	id, err := uuid.Parse(chirpID)
	if err != nil {
		cfg.respondWithError(w, http.StatusNotFound, "chirp not found")
		return
	}
	dbChirp, err := cfg.queries.GetChirp(r.Context(), id)
	if err != nil {
		cfg.respondWithError(w, http.StatusInternalServerError, "chirp unable to be fetched")
		return
	}
	// Map DB Query to a Go Struct
	chirp := Chirp{
		ID:        dbChirp.ID,
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,
		Body:      dbChirp.Body,
		UserID:    dbChirp.UserID.UUID,
	}
	// MarshalChirp to JSON response and send out!

	cfg.respondWithJSON(w, http.StatusOK, chirp)
}
