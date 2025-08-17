package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"chirpy.com/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) polkaHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Event string `json:"event"`
		Data struct {
			UserID string `json:"user_id"`
		} `json:"data"`
	}

	decoder := json.NewDecoder(r.Body)
	req := parameters{}
	err := decoder.Decode(&req)
	if err != nil {
		msg := fmt.Sprintf("Error marshalling JSON: %s", err)
		cfg.respondWithError(w, 401, msg)
		return
	}

	if req.Event != "user.upgraded" {
		w.WriteHeader(204)
		return
	}

	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		msg := fmt.Sprintf("Error extracting api key from header: %s", err)
		cfg.respondWithError(w, 401, msg)
		return
	}

	if apiKey != os.Getenv("POLKA_KEY") {
		w.WriteHeader(401)
		return 
	}

	userUUID, err := uuid.Parse(req.Data.UserID)
	if err != nil {
		msg := fmt.Sprintf("Error parsing uuid from user_id: %s", err)
		cfg.respondWithError(w, http.StatusBadRequest, msg)
		return
	}

	log.Printf("Attempting to upgrade user with ID: %s", userUUID)

	_, err = cfg.queries.UpgradeUser(r.Context(), userUUID)
	if err != nil {
		log.Printf("Error upgrading user: %s", userUUID)
		cfg.respondWithError(w, http.StatusNotFound, "User ID not found")
		return
	}

	log.Printf("Successfully upgraded user with ID: %s", userUUID)

	w.WriteHeader(204)

}