package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) error {
	response, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(code)
	w.Write(response)
	return nil
}

func respondWithError(w http.ResponseWriter, code int, msg string) error {
	return respondWithJSON(w, code, map[string]string{"error": msg})
}

func cleanBody(content string) string {
	badWords := []string{"kerfuffle", "sharbert", "fornax"}
	cleanedWords := []string{}
	words := strings.Fields(content)
	for _, word := range words {
		for _, bword := range badWords {
			if strings.ToLower(word) == bword {
				word = "****"
			}
		}
		cleanedWords = append(cleanedWords, word)
	}
	cleanedContent := strings.Join(cleanedWords, " ")
	return cleanedContent

}
