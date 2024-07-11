package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/TedMartell/ChirpyServerProject/internal/database"
	"github.com/gorilla/mux"
)

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

	// Create a new router
	r := mux.NewRouter()

	// Define routes
	r.HandleFunc("/api/healthz", handlerReadiness).Methods("GET")
	r.HandleFunc("/admin/metrics", apiCfg.handlerMetrics).Methods("GET")
	r.PathPrefix("/app/").Handler(apiCfg.middlewareMetricsInc(fileServer))
	r.HandleFunc("/api/reset", apiCfg.handlerReset).Methods("POST")
	r.HandleFunc("/api/chirps", apiCfg.handlerCreateChirp).Methods("POST")
	r.HandleFunc("/api/chirps", apiCfg.handlerGetChirps).Methods("GET")
	r.HandleFunc("/api/chirps/{chirpID}", apiCfg.handlerGetChirpByID).Methods("GET")
	r.HandleFunc("/api/users", apiCfg.handlerCreateUser).Methods("POST")

	// Start the server
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	dbg := flag.Bool("debug", false, "Enable debug mode")
	flag.Parse()

	if *dbg {
		// Try to delete the file and handle any potential errors
		err := os.Remove("./database.json") // Adjust path for Unix-like systems
		if err != nil {
			fmt.Println("Error deleting database:", err)
		} else {
			fmt.Println("Database deleted successfully.")
		}
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())
}
