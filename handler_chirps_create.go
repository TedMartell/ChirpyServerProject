package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/TedMartell/ChirpyServerProject/internal/auth"
)

type Chirp struct {
	ID   int    `json:"id"`
	Body string `json:"body"`
}

func (cfg *apiConfig) handlerChirpsCreate(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	// Extract the token from the request headers using your auth package.
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Missing or malformed token")
		return
	}

	// Validate the token and extract the user ID using your auth package.
	userIDStr, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid token")
		return
	}

	// Convert userID from string to integer
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Invalid user ID")
		return
	}

	// Decode the request body.
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	cleaned, err := validateChirp(params.Body)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Create the chirp with the author_id.
	chirp, err := cfg.DB.CreateChirp(cleaned, userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create chirp")
		return
	}

	// Include the author_id in the response using the database `Chirp` struct.
	respondWithJSON(w, http.StatusCreated, chirp)
}

func validateChirp(body string) (string, error) {
	const maxChirpLength = 140
	if len(body) > maxChirpLength {
		return "", errors.New("Chirp is too long")
	}

	badWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}
	cleaned := getCleanedBody(body, badWords)
	return cleaned, nil
}

func getCleanedBody(body string, badWords map[string]struct{}) string {
	words := strings.Split(body, " ")
	for i, word := range words {
		loweredWord := strings.ToLower(word)
		if _, ok := badWords[loweredWord]; ok {
			words[i] = "****"
		}
	}
	cleaned := strings.Join(words, " ")
	return cleaned
}
