package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"

	"github.com/TedMartell/ChirpyServerProject/internal/auth"
	"github.com/TedMartell/ChirpyServerProject/internal/database"
)

func (cfg *apiConfig) handlerPolkaWebhooks(w http.ResponseWriter, r *http.Request) {

	// Validate the API key
	err1 := auth.ValidateAPI(r, os.Getenv("POLKA_KEY"))
	if err1 != nil {
		respondWithError(w, http.StatusUnauthorized, err1.Error())
		return
	}
	// Define the webhook struct
	type webhook struct {
		Event string `json:"event"`
		Data  struct {
			UserID int `json:"user_id"`
		} `json:"data"`
	}

	// Decode the JSON request body
	var wh webhook
	err := json.NewDecoder(r.Body).Decode(&wh)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Check if the event is "user.upgraded"
	if wh.Event != "user.upgraded" {
		respondWithError(w, http.StatusNoContent, "Invalid event")
		return
	}

	// Fetch the user from the database
	user, err := cfg.DB.GetUser(wh.Data.UserID)
	if err != nil {
		if errors.Is(err, database.ErrNotExist) {
			respondWithError(w, http.StatusNotFound, "User not found")
		} else {
			respondWithError(w, http.StatusInternalServerError, "Database error")
		}
		return
	}

	// Update the user to be a Chirpy Red member
	user.IsChirpyRed = true

	// Update the user in the database
	_, err = cfg.DB.UpdateUser(user.ID, user.Email, user.HashedPassword, user.IsChirpyRed)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to update user")
		return
	}

	// Respond with a 204 status code for a successful update
	w.WriteHeader(http.StatusNoContent)
}
