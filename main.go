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

	// Set up the serve mux and wrap the file server with the middleware
	mux := http.NewServeMux()
	mux.Handle("/healthz", http.HandlerFunc(handlerReadiness))
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(fileServer))
	mux.HandleFunc("/metrics", apiCfg.handlerMetrics)
	mux.HandleFunc("/reset", apiCfg.handlerReset)

	// Add `mux.Handle` for `/metrics` and `/reset`

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())
}
