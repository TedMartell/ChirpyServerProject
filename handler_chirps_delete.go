package main

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/TedMartell/ChirpyServerProject/internal/auth"
	"github.com/TedMartell/ChirpyServerProject/internal/database"
)

func (cfg *apiConfig) handlerChirpsDelete(w http.ResponseWriter, r *http.Request) {
	// Manually parsing the URL to get the chirpID
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 || pathParts[2] != "chirps" {
		respondWithError(w, http.StatusBadRequest, "Missing or invalid chirp ID")
		return
	}
	chirpIDStr := pathParts[3]
	chirpID, err := strconv.Atoi(chirpIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid chirp ID")
		return
	}

	// Extract the token from the request headers
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Missing or malformed token")
		return
	}

	// Validate the token and extract the user ID
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

	// Fetch the chirp from the database
	chirp, err := cfg.DB.GetChirp(chirpID)
	if err != nil {
		if err == database.ErrNotExist {
			respondWithError(w, http.StatusNotFound, "Chirp not found")
		} else {
			respondWithError(w, http.StatusInternalServerError, "Database error")
		}
		return
	}

	// Check if the user is the author of the chirp
	if chirp.AuthorID != userID {
		respondWithError(w, http.StatusForbidden, "You are not the author of this chirp")
		return
	}

	err = cfg.DB.DeleteChirp(chirpID)
	if err != nil {
		if err == database.ErrNotExist {
			respondWithError(w, http.StatusNotFound, "Chirp not found")
		} else {
			respondWithError(w, http.StatusInternalServerError, "Database error")
		}
		return
	}

	// Respond with a 204 No Content status
	w.WriteHeader(http.StatusNoContent)
}
