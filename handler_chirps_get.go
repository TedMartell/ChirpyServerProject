package main

import (
	"net/http"
	"sort"
	"strconv"

	"github.com/TedMartell/ChirpyServerProject/internal/database"
)

func (cfg *apiConfig) handlerChirpsGet(w http.ResponseWriter, r *http.Request) {
	chirpIDString := r.PathValue("chirpID")
	chirpID, err := strconv.Atoi(chirpIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid chirp ID")
		return
	}

	dbChirp, err := cfg.DB.GetChirp(chirpID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Couldn't get chirp")
		return
	}

	respondWithJSON(w, http.StatusOK, Chirp{
		ID:   dbChirp.ID,
		Body: dbChirp.Body,
	})
}

func (cfg *apiConfig) handlerChirpsRetrieve(w http.ResponseWriter, r *http.Request) {
	s := r.URL.Query().Get("author_id")
	srt := r.URL.Query().Get("sort")

	dbChirps, err := cfg.DB.GetChirps()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't retrieve chirps")
		return
	}

	chirps := []database.Chirp{}

	// Collect all chirps first
	for _, dbChirp := range dbChirps {
		chirps = append(chirps, database.Chirp{
			ID:       dbChirp.ID,
			AuthorID: dbChirp.AuthorID,
			Body:     dbChirp.Body,
		})
	}

	// Filter based on author_id if provided
	if s != "" {
		authorID, err := strconv.Atoi(s)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid author ID")
			return
		}
		// Filter after collecting all chirps
		filteredChirps := []database.Chirp{}
		for _, chirp := range chirps {
			if chirp.AuthorID == authorID {
				filteredChirps = append(filteredChirps, chirp)
			}
		}
		chirps = filteredChirps
	}

	// Sort based on the sort parameter
	if srt == "desc" {
		sort.Slice(chirps, func(i, j int) bool {
			return chirps[j].ID < chirps[i].ID
		})
	} else {
		// Default to ascending if sort is "asc" or empty
		sort.Slice(chirps, func(i, j int) bool {
			return chirps[i].ID < chirps[j].ID
		})
	}

	// Respond with sorted chirps
	respondWithJSON(w, http.StatusOK, chirps)
}
