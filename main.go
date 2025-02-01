package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"chirpy.com/internal/database"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	queries        *database.Queries
	platform       string
}

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

type Email struct {
	Email string `json:"email"`
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
	if cfg.platform != "dev" {
		w.WriteHeader(403)
		return
	}

	err := cfg.queries.DeleteAllUsers(r.Context())
	if err != nil {
		cfg.respondWithError(w, 500, "Failed to delete users")
		return
	}

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

func (cfg *apiConfig) chirpsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	params := CreateChirpRequest{}
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
	// Chirp is valid if past this point
	chirp, err := cfg.queries.CreateChirp(r.Context(), database.CreateChirpParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Body:      cleanedBody,
		UserID: uuid.NullUUID{
			UUID:  params.UserID,
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

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Connection to database failed with error: %v. Check your DB_URL (%s)", err, dbURL)
	}
	dbQueries := database.New(db)
	if err := db.Ping(); err != nil {
		log.Fatalf("Could not connect to the database: %v", err)
	}
	defer db.Close()
	cfg := &apiConfig{
		queries:  dbQueries,
		platform: os.Getenv("PLATFORM"),
	}
	const filepathRoot = "."
	const port = "8080"
	mux := http.NewServeMux()
	fileServer := http.StripPrefix("/app/", http.FileServer(http.Dir(filepathRoot)))
	mux.Handle("/app/", cfg.middlewareMetricsInc(fileServer))
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	mux.HandleFunc("GET /api/healthz", healthHandler)
	mux.HandleFunc("POST /api/users", cfg.userHandler)
	mux.HandleFunc("GET /admin/metrics", cfg.metricsHandler)
	mux.HandleFunc("POST /admin/reset", cfg.resetHandler)
	mux.HandleFunc("POST /api/chirps", cfg.chirpsHandler)

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())
}
