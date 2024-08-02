package main

import (
	"net/http"
	"strings"
	"time"

	"github.com/TedMartell/ChirpyServerProject/internal/auth"
	"github.com/TedMartell/ChirpyServerProject/internal/database"
)

func (cfg *apiConfig) handlerRefreshTokens(w http.ResponseWriter, r *http.Request) {
	// Extract the refresh token from the Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		respondWithError(w, http.StatusUnauthorized, "Authorization header is required")
		return
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		respondWithError(w, http.StatusUnauthorized, "Invalid Authorization header format")
		return
	}

	refreshToken := parts[1]

	// Look up the refresh token in the database
	dbStructure, err := cfg.DB.LoadDB()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}

	var user *database.User
	for _, u := range dbStructure.Users {
		if u.RefreshToken.Token == refreshToken {
			user = &u
			break
		}
	}

	if user == nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid refresh token")
		return
	}

	// Check if the refresh token has expired
	if user.RefreshToken.ExpiresAt.Before(time.Now()) {
		respondWithError(w, http.StatusUnauthorized, "Refresh token has expired")
		return
	}

	// Generate a new access token (JWT)
	token, err := auth.MakeJWT(user.ID, cfg.jwtSecret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create JWT")
		return
	}

	// Respond with the new access token
	respondWithJSON(w, http.StatusOK, map[string]string{
		"token": token,
	})
}

func (cfg *apiConfig) handlerRevokeTokens(w http.ResponseWriter, r *http.Request) {
	// Extract the refresh token from the Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		respondWithError(w, http.StatusUnauthorized, "Authorization header is required")
		return
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		respondWithError(w, http.StatusUnauthorized, "Invalid Authorization header format")
		return
	}

	refreshToken := parts[1]

	// Look up the refresh token in the database
	dbStructure, err := cfg.DB.LoadDB()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}

	var userIDToRemove int
	found := false

	for userID, u := range dbStructure.Users {
		if u.RefreshToken.Token == refreshToken {
			userIDToRemove = userID
			found = true
			break
		}
	}

	if !found {
		respondWithError(w, http.StatusUnauthorized, "Invalid refresh token")
		return
	}

	// Revoke the refresh token by removing the user from the database
	delete(dbStructure.Users, userIDToRemove)

	// Save the updated database structure
	err = cfg.DB.SaveDB(dbStructure)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}

	// Respond with a 204 status code to indicate success
	w.WriteHeader(http.StatusNoContent)
}
