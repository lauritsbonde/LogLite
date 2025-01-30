package handlers

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/lauritsbonde/LogLite/src/webApp/components"
)

func CollectType(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()

	collectType := queryParams.Get("type")

	switch collectType {
		case "send":
			templ.Handler(components.SendOption()).ServeHTTP(w, r)
		case "scrape":
			templ.Handler(components.ScrapeOption()).ServeHTTP(w, r)
		default:
			http.Error(w, "Invalid or missing 'type' parameter", http.StatusBadRequest)
	}
}

func DBType(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()

	dbType := queryParams.Get("type")

	switch dbType {
		case "sqlite":
			templ.Handler(components.SQLiteOption()).ServeHTTP(w, r)
		default:
			http.Error(w, "Invalid or missing 'type' parameter", http.StatusBadRequest)
	}
}