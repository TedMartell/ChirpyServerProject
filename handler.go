package main

import (
	"fmt"
	"net/http"
)

type apiConfig struct {
	fileserverHits int
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
