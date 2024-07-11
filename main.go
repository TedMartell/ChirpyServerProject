package main

import (
	"log"
	"net/http"

	"github.com/TedMartell/ChirpyServerProject/internal/database"
)

// Add a reference to your DB in apiConfig

func main() {
	const filepathRoot = "."
	const port = "8080"

	// Initialize the database
	db, err := database.NewDB("database.json")
	if err != nil {
		log.Fatalf("Failed to initialize the database: %v", err)
	}

	// Initialize apiConfig with the database
	apiCfg := &apiConfig{
		db: db,
	}

	// Set up a file server handler
	fileServer := http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))

	// Set up the serve mux and add handlers for the GET methods only
	mux := http.NewServeMux()
	mux.HandleFunc("/api/healthz", handlerReadiness)
	mux.HandleFunc("/admin/metrics", apiCfg.handlerMetrics)
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(fileServer))
	mux.HandleFunc("/api/reset", apiCfg.handlerReset)
	mux.HandleFunc("/api/chirps", apiCfg.handlerChirps)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())
}
