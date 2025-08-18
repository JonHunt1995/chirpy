package main

import (
	"encoding/json"
	"net/http"

	"chirpy.com/internal/database"
)

func (cfg *apiConfig) getAllChirpsHandler(w http.ResponseWriter, r *http.Request) {
	// Extract Author ID Query Param
	var dbChirps []database.Chirp;
	authorId := r.URL.Query().Get("author_id")
	// Querying the database for Chirps
	if len(authorId) == 0 {
		dbChirps, err := cfg.queries.GetAllChirps(r.Context())
		if err != nil {
			cfg.respondWithError(w, http.StatusInternalServerError, "Failed to fetch chirps")
			return
		}
	} else {
		dbChirps, err := cfg.queries.GetAllChirps(r.Context())
		if err != nil {
			cfg.respondWithError(w, http.StatusInternalServerError, "Failed to fetch chirps")
			return
		}
	}
	
	// Explicitly map SQLC chirps to custom Chirp Struct
	chirps := make([]Chirp, len(dbChirps))
	for i, dbChirp := range dbChirps {
		chirps[i] = Chirp{
			ID:        dbChirp.ID,
			CreatedAt: dbChirp.CreatedAt,
			UpdatedAt: dbChirp.UpdatedAt,
			Body:      dbChirp.Body,
			UserID:    dbChirp.UserID.UUID,
		}
	}
	// Set headers for JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	// Encoding chirps to JSON and writes to w
	err = json.NewEncoder(w).Encode(chirps)
	if err != nil {
		cfg.respondWithError(w, http.StatusInternalServerError, "Failed to encode chirps")
		return
	}
}
