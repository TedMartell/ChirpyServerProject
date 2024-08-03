package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/TedMartell/ChirpyServerProject/internal/auth"
	"github.com/TedMartell/ChirpyServerProject/internal/database"
)

func (cfg *apiConfig) handlerUsersUpdate(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	type response struct {
		database.User
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT")
		return
	}

	subject, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT")
		return
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't hash password")
		return
	}

	userIDInt, err := strconv.Atoi(subject)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't parse user ID")
		return
	}

	// Fetch the existing user to get their current IsChirpyRed status
	existingUser, err := cfg.DB.GetUser(userIDInt)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't fetch user")
		return
	}

	// Update the user with the new email and password while keeping IsChirpyRed status
	user, err := cfg.DB.UpdateUser(userIDInt, params.Email, hashedPassword, existingUser.IsChirpyRed)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't update user")
		return
	}

	respondWithJSON(w, http.StatusOK, response{
		User: database.User{
			ID:          user.ID,
			Email:       user.Email,
			IsChirpyRed: user.IsChirpyRed, // Ensure to include the ChirpyRed status
		},
	})
}
