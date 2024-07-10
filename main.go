package main

import (
	"log"
	"net/http"
)

func main() {
	const filepathRoot = "."
	const port = "8080"

	// Initialize apiConfig
	apiCfg := &apiConfig{}

	// Set up a file server handler
	fileServer := http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))

	// Set up the serve mux and add handlers for the GET methods only
	mux := http.NewServeMux()
	mux.HandleFunc("/api/healthz", handlerReadiness)
	mux.HandleFunc("/admin/metrics", apiCfg.handlerMetrics)
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(fileServer))
	mux.HandleFunc("/api/reset", apiCfg.handlerReset)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())
}
