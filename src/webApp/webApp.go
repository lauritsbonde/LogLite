package webapp

import (
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/a-h/templ"
	confighandler "github.com/lauritsbonde/LogLite/src/configHandler"
	dbhandler "github.com/lauritsbonde/LogLite/src/dbHandler"
	"github.com/lauritsbonde/LogLite/src/webApp/handlers"
	"github.com/lauritsbonde/LogLite/src/webApp/views"
)

type WebApp struct {
	mu        sync.RWMutex // this lock does not handle any db synchronization, it is used for not changing the dbhandler while reads are ongoing
	DBHandler dbhandler.DBHandler
	SettingsChan chan string
	Configuration *confighandler.Config
}

// SetDBHandler updates the dbHandler safely with a write lock.
func (app *WebApp) SetDBHandler(handler dbhandler.DBHandler) {
	app.mu.Lock()
	defer app.mu.Unlock()
	app.DBHandler = handler
}

// GetDBHandler retrieves the dbHandler safely with a read lock.
func (app *WebApp) GetDBHandler() dbhandler.DBHandler {
	app.mu.RLock()
	defer app.mu.RUnlock()
	return app.DBHandler
}

func (app *WebApp) indexHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Safely get the current dbHandler
		dbHandler := app.GetDBHandler()

		if dbHandler == nil {
			log.Print("No db\n")
		}

		// Render logs with templ.Handler - if ther version is empty, then there is no config
		templ.Handler(views.Index(app.Configuration.Version == "")).ServeHTTP(w, r)
	})
}

func assetHandler(w http.ResponseWriter, r *http.Request) {
		filepath := strings.TrimPrefix(r.URL.Path, "/asset/")

		assetsDir := "./src/webApp/assets/"

		fullPath := assetsDir + filepath

		http.ServeFile(w, r, fullPath)
}

func (app *WebApp) RunWebApp() error {
	// Register the index "/" route - The special wildcard {$} matches only the end of the URL. For example, the pattern "/{$}" matches only the path "/", whereas the pattern "/" matches every path.
	http.Handle("GET /{$}", app.indexHandler())

	http.HandleFunc("GET /asset/", assetHandler)

	http.HandleFunc("GET /ingest-options", handlers.CollectType)
	http.HandleFunc("GET /db-options", handlers.DBType)

	http.HandleFunc("POST /setup", func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		collectType := (r.FormValue("collect-type")) // send || scrape
		var subtype string

		switch collectType {
			case "send":
				subtype = r.FormValue("endpoint-type") // http || udp
			case "scrap":
				http.Error(w, "not implemented yet", http.StatusNotImplemented)
		}
		
		db := r.FormValue("database-type")
		file := r.FormValue("sqlite-path")

		log.Printf("%s, %s, %s, %s\n", collectType, subtype, db, file)


	})

	// Register the "/livelogs" route
	http.HandleFunc("GET /livelogs", func(w http.ResponseWriter, r *http.Request) {
		dbHandler := app.GetDBHandler()

		if dbHandler == nil {
			http.Error(w, "No database configured. Please configure the database.", http.StatusServiceUnavailable)
			return
		}

		handlers.LiveLogs(w, r, dbHandler)
	})

	// Start the HTTP server
	return http.ListenAndServe(":8080", nil)
}