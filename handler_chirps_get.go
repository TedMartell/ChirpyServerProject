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

	dbChirps, err := cfg.DB.GetChirps()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't retrieve chirps")
		return
	}

	chirps := []database.Chirp{}
	if s != "" {
		// Convert author_id to integer
		authorID, err := strconv.Atoi(s)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid author ID")
			return
		}

		// Filter based on author_id
		for _, dbChirp := range dbChirps {
			if dbChirp.AuthorID == authorID {
				chirps = append(chirps, database.Chirp{
					ID:       dbChirp.ID,
					AuthorID: dbChirp.AuthorID,
					Body:     dbChirp.Body,
				})
			}
		}
	} else {
		// Otherwise, include all chirps
		for _, dbChirp := range dbChirps {
			chirps = append(chirps, database.Chirp{
				ID:       dbChirp.ID,
				AuthorID: dbChirp.AuthorID,
				Body:     dbChirp.Body,
			})
		}
	}

	// Sort by ID in ascending order
	sort.Slice(chirps, func(i, j int) bool {
		return chirps[i].ID < chirps[j].ID
	})

	respondWithJSON(w, http.StatusOK, chirps)
}
