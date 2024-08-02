package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/TedMartell/ChirpyServerProject/internal/auth"
)

// Define a new response structure to include the refresh token
type response struct {
	User
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Password         string `json:"password"`
		Email            string `json:"email"`
		ExpiresInSeconds int    `json:"expires_in_seconds"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	user, err := cfg.DB.GetUserByEmail(params.Email) // Adjust according to your database getting functions
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get user")
		return
	}

	err = auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid password")
		return
	}

	defaultExpiration := 60 * 60 // 1 hour for JWT tokens
	if params.ExpiresInSeconds == 0 {
		params.ExpiresInSeconds = defaultExpiration
	} else if params.ExpiresInSeconds > defaultExpiration {
		params.ExpiresInSeconds = defaultExpiration
	}

	// Generate the JWT token
	token, err := auth.MakeJWT(user.ID, cfg.jwtSecret, time.Duration(params.ExpiresInSeconds)*time.Second)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create JWT")
		return
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create refresh token")
		return
	}

	// Set the refresh token expiration time (e.g., 60 days)
	refreshTokenExpiration := time.Now().Add(60 * 24 * time.Hour)

	// Store the refresh token in the database
	err = cfg.DB.StoreRefreshToken(user.ID, refreshToken, refreshTokenExpiration)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't store refresh token")
		return
	}

	// Respond with both the access token (JWT) and the refresh token
	respondWithJSON(w, http.StatusOK, response{
		User: User{
			ID:    user.ID,
			Email: user.Email,
		},
		Token:        token,
		RefreshToken: refreshToken,
	})
}
