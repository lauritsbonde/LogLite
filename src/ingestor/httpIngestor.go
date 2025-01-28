package ingestor

import (
	"fmt"
	"log"
	"net/http"

	dbhandler "github.com/lauritsbonde/LogLite/src/dbHandler"
)

type HTTPIngestor struct {
	Port      int
	dbHandler dbhandler.DBHandler // Database handler to save data
}

func (h *HTTPIngestor) Start() error {
	// Use a custom handler for this server - to not use the gloabl http.DefaultServeMux
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "Hello from HTTP Ingestor!")
	})

	// Create a custom HTTP server on port h.Port
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", h.Port),
		Handler: mux, // Use the custom ServeMux for this server
	}

	log.Printf("HTTP server is running on port %d\n", h.Port)
	return server.ListenAndServe()
}

func (h *HTTPIngestor) Stop() error {
	// Optional: Add logic to gracefully stop the HTTP server if needed
	return nil
}

func (h *HTTPIngestor) SetDBHandler(dbHandler dbhandler.DBHandler) {
	h.dbHandler = dbHandler
}