package ingestor

import (
	"fmt"
	"net/http"

	dbhandler "github.com/lauritsbonde/LogLite/src/dbHandler"
)

type HTTPIngestor struct {
	Port      int
	dbHandler dbhandler.DBHandler // Database handler to save data
}

func (h *HTTPIngestor) Start() error {
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "Hello from HTTP Ingestor!")
	})
	return http.ListenAndServe(fmt.Sprintf(":%d", h.Port), nil)
}

func (h *HTTPIngestor) Stop() error {
	// Optional: Add logic to gracefully stop the HTTP server if needed
	return nil
}

func (h *HTTPIngestor) SetDBHandler(dbHandler dbhandler.DBHandler) {
	h.dbHandler = dbHandler
}