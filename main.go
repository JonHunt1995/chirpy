package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) formatHitsCount() string {
	return fmt.Sprintf(
		`<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`, cfg.fileserverHits.Load())
}

func (cfg *apiConfig) metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte(cfg.formatHitsCount()))
}

func (cfg *apiConfig) resetHandler(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)
	w.WriteHeader(200)
}

func (cfg *apiConfig) respondWithError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	respBody := ErrorResponse{
		Error: msg,
	}
	data, err := json.Marshal(respBody)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.Write(data)
}

func (cfg *apiConfig) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.Write(data)
}

func (cfg *apiConfig) chirpValidationHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
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
	respBody := CleanedChirp{
		CleanedBody: cleanedBody,
	}
	cfg.respondWithJSON(w, 200, respBody)
	return
}

func removeProfanity(msg string) string {
	badWords := map[string]string{
		"kerfuffle": "****",
		"sharbert":  "****",
		"fornax":    "****",
	}
	words := strings.Split(msg, " ")
	for i, word := range words {
		loweredWord := strings.ToLower(word)
		if _, ok := badWords[loweredWord]; !ok {
			continue
		}
		words[i] = strings.Replace(word, word, badWords[loweredWord], 1)
	}
	cleanedMsg := strings.Join(words, " ")
	return cleanedMsg
}

func main() {
	cfg := &apiConfig{}
	const filepathRoot = "."
	const port = "8080"
	mux := http.NewServeMux()
	fileServer := http.StripPrefix("/app/", http.FileServer(http.Dir(filepathRoot)))
	mux.Handle("/app/", cfg.middlewareMetricsInc(fileServer))
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	mux.HandleFunc("/api/healthz", healthHandler)

	mux.HandleFunc("/admin/metrics", cfg.metricsHandler)
	mux.HandleFunc("/admin/reset", cfg.resetHandler)
	mux.HandleFunc("/api/validate_chirp", cfg.chirpValidationHandler)
	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())
}
