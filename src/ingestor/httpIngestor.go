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
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
			fmt.Fprintf(w, "Hello from HTTP Ingestor!")
	})

	server := &http.Server{
			Addr:    fmt.Sprintf(":%d", h.Port),
			Handler: mux,
	}

	log.Printf("HTTP server is running on port %d\n", h.Port)

	go func() {
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					log.Fatalf("HTTP server error: %v", err)
			}
	}()

	return nil // This allows the Start function to return immediately
}

func (h *HTTPIngestor) Stop() error {
	// Optional: Add logic to gracefully stop the HTTP server if needed
	return nil
}

func (h *HTTPIngestor) SetDBHandler(dbHandler dbhandler.DBHandler) {
	h.dbHandler = dbHandler
}