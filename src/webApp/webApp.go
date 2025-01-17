package webapp

import (
	"fmt"
	"net/http"

	"github.com/a-h/templ"
	dbhandler "github.com/lauritsbonde/LogLite/src/dbHandler"
	"github.com/lauritsbonde/LogLite/src/webApp/handlers"
	"github.com/lauritsbonde/LogLite/src/webApp/views"
)

func IndexHandler(dbHandler dbhandler.DBHandler) http.Handler {
	// Attempt to get logs
	logs, err := handlers.GetLogs(dbHandler)
	if err != nil {
		fmt.Printf("Error fetching logs: %v", err)

		// Return a handler for an error page
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		})
	}

	// Return the templ.Handler if no error occurred
	return templ.Handler(views.Index(logs))
}

func RunWebApp(dbHandler dbhandler.DBHandler) error {
	http.Handle("GET /", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		IndexHandler(dbHandler).ServeHTTP(w, r)
	}))

	http.HandleFunc("GET /api/paginatedLogs", http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
		handlers.GetPaginatedLogs(w, r, dbHandler)
	}))

	// Start the HTTP server
	return http.ListenAndServe(":8080", nil)
}