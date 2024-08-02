package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"golang.org/x/crypto/bcrypt"

	"github.com/gorilla/mux"

	"github.com/TedMartell/ChirpyServerProject/internal/database"
)

type Chirp struct {
	ID   int    `json:"id"`
	Body string `json:"body"`
}

type apiConfig struct {
	fileserverHits int
	db             *database.DB
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits++ // Increment the counter
		next.ServeHTTP(w, r) // Call the next handler in the chain
	})
}

func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	// Check for allowed method
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed) // 405 status code
		return
	}
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	// 1. Ensure only GET requests allowed
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// 2. Set Content-Type to text/html
	w.Header().Add("Content-Type", "text/html; charset=utf-8")

	// 3. HTML Template as a string
	htmlTemplate := `
    <html>
    <body>
        <h1>Welcome, Chirpy Admin</h1>
        <p>Chirpy has been visited %d times!</p>
    </body>
    </html>`

	// 4. Format the HTML string with the hit count
	formattedHTML := fmt.Sprintf(htmlTemplate, cfg.fileserverHits)

	// 5. Write the formatted HTML to the response
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(formattedHTML))
}

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	cfg.fileserverHits = 0
	w.Write([]byte("fileserverHits has been reset to 0"))
}

func (cfg *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	chirps, err := cfg.db.GetChirps()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve chirps")
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(chirps)
}

func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Body string `json:"body"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request: %v", err)
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	if len(req.Body) > 140 {
		log.Printf("Body exceeds 140 characters: %d", len(req.Body))
		respondWithError(w, http.StatusBadRequest, "Body cannot exceed 140 characters")
		return
	}

	// Clean the body content
	cleanedBody := cleanBody(req.Body)
	log.Printf("Cleaned body: %s", cleanedBody)

	chirp, err := cfg.db.CreateChirp(cleanedBody)
	if err != nil {
		log.Printf("Error creating chirp: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to create chirp")
		return
	}

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(chirp); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

func (cfg *apiConfig) handlerGetChirpByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r) // corrected from mutex.Vars to mux.Vars
	chirpIDStr := vars["chirpID"]

	chirpID, err := strconv.Atoi(chirpIDStr)
	if err != nil {
		http.Error(w, "Invalid chirpID", http.StatusBadRequest)
		return
	}

	chirp, err := cfg.db.GetChirpByID(chirpID)
	if err != nil {
		if err.Error() == "chirp not found" {
			http.Error(w, "Chirp not found", http.StatusNotFound)
		} else {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			log.Printf("Error retrieving chirp: %v", err)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(chirp)
	if err != nil {
		http.Error(w, "Failed to marshal chirp into JSON", http.StatusInternalServerError)
		log.Printf("Error encoding response: %v", err)
	}
}

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request: %v", err)
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	// Hashing the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	// Create user with hashed password
	user, err := cfg.db.CreateUser(req.Email, string(hashedPassword))
	if err != nil {
		log.Printf("Error creating user: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	// Create response without the password
	response := struct {
		ID    int    `json:"id"`
		Email string `json:"email"`
	}{
		ID:    user.ID,
		Email: user.Email,
	}

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

func (cfg *apiConfig) handlerAuthorizePassword(w http.ResponseWriter, r *http.Request) {
	// Only allow POST method
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request: %v", err)
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	defer r.Body.Close()

	// Check if user exists
	user, err := cfg.db.GetUserByEmail(req.Email)
	if err != nil {
		log.Printf("Error getting user: %v", err)
		respondWithError(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	// Compare stored hashed password with provided password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		log.Printf("Invalid password for user: %s", req.Email)
		respondWithError(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	// Create response without the password
	response := struct {
		ID    int    `json:"id"`
		Email string `json:"email"`
	}{
		ID:    user.ID,
		Email: user.Email,
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}
