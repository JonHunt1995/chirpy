package main

type ValidResponse struct {
	Valid bool `json:"valid"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type CleanedChirp struct {
	CleanedBody string `json:"cleaned_body"`
}
