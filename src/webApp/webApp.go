package webapp

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/a-h/templ"
	confighandler "github.com/lauritsbonde/LogLite/src/configHandler"
	dbhandler "github.com/lauritsbonde/LogLite/src/dbHandler"
	"github.com/lauritsbonde/LogLite/src/webApp/handlers"
	"github.com/lauritsbonde/LogLite/src/webApp/views"
)

type WebApp struct {
	DBHandler dbhandler.DBHandler
	
	SettingsChan chan ConfigMessage
	Configuration *confighandler.Config
}

type ConfigMessage struct {
	NewConfig *confighandler.Config
	ResponseCh chan string
}


func (app *WebApp) indexHandler(w http.ResponseWriter, r *http.Request) {
	// Render logs with templ.Handler - if ther version is empty, then there is no config
	templ.Handler(views.Index(app.Configuration.Version == "")).ServeHTTP(w, r)
}

func (app *WebApp) settingsHandler(w http.ResponseWriter, r *http.Request) {
	// Render logs with templ.Handler - if ther version is empty, then there is no config
	templ.Handler(views.Settings()).ServeHTTP(w, r)
}

func (app *WebApp) setupHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Parse form values
	collectType := r.FormValue("collect-type") // send || scrape
	var sendConfig confighandler.Send
	var scrapeConfig confighandler.Scrape

	switch collectType {
	case "send":
		protocol := r.FormValue("endpoint-type") // http || udp
		port := r.FormValue("ingest-port")
		portParsed, err := strconv.Atoi(port)
		if err != nil {
			log.Printf("Error parsing port: %v\n", err) // Log the error but continue execution
		} 
		sendConfig = confighandler.Send{
			Protocol: protocol,
			Port: portParsed,
		}
	case "scrape":
		scrapeType := r.FormValue("scrape-type") // pure_docker || docker_swarm || kubernetes
		scrapeConfig = confighandler.Scrape{
			Type: scrapeType,
		}
	default:
		http.Error(w, "Invalid collect-type value. Must be 'send' or 'scrape'.", http.StatusBadRequest)
		return
	}

	db := r.FormValue("database-type") // Required
	file := r.FormValue("sqlite-path")
	if db == "SQLite" && file == "" {
		http.Error(w, "sqlite-path is required when database-type is SQLite", http.StatusBadRequest)
		return
	}

	logLevel := r.FormValue("log-level")
	logFile := r.FormValue("logfile-path")

	// Create a new config object
	newConfig := confighandler.Config{
		Version:        "1.0.0",
		LogLevel:       logLevel,
		LogFile:        logFile,
		MaxConnections: 100, // Example default value
		LogHandler: confighandler.LogHandler{
			Mode:   collectType,
			Send:   sendConfig,
			Scrape: scrapeConfig,
		},
		Database: confighandler.Database{
			Type:           db,
			SQLiteFilepath: file,
		},
	}

	err = confighandler.ValidateConfig(newConfig)
	if err != nil {
		log.Printf("error validating new config %v \n", err)
		return
	}

	// save the config to a file
	err = confighandler.SaveConfig(newConfig, "./etc/config.yaml")
	if err != nil {
		log.Printf("error saving new config %v \n", err)
		return
	}


	// Send newConfig to the main thread
	responseCh := make(chan string)
	app.SettingsChan <- ConfigMessage{
		NewConfig:  &newConfig,
		ResponseCh: responseCh,
	}

	// Wait for the response
	response := <-responseCh
	if strings.HasPrefix(response, "Error") {
		http.Error(w, response, http.StatusInternalServerError)
		return
	}

	// Respond with success
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Configuration setup successfully: " + response))
}

func assetHandler(w http.ResponseWriter, r *http.Request) {
		filepath := strings.TrimPrefix(r.URL.Path, "/asset/")

		assetsDir := "./src/webApp/assets/"

		fullPath := assetsDir + filepath

		http.ServeFile(w, r, fullPath)
}

func middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Example: Log the request details
			log.Printf("Request: %s %s", r.Method, r.URL.Path)

			
			// Call the next handler
			next.ServeHTTP(w, r)
	})
}

// Middleware wrapper for http.HandlerFunc
func middlewareFunc(next http.HandlerFunc) http.Handler {
	return middleware(http.HandlerFunc(next))
}

// Modify RunWebApp to apply middleware
func (app *WebApp) RunWebApp() error {
	// Wrap routes with middleware
	http.Handle("/{$}", middleware(http.HandlerFunc(app.indexHandler)))
	http.Handle("/asset/", middlewareFunc(assetHandler))
	http.Handle("/ingest-options", middlewareFunc(handlers.CollectType))
	http.Handle("/db-options", middlewareFunc(handlers.DBType))
	http.Handle("/setup", middleware(http.HandlerFunc(app.setupHandler)))
	http.Handle("/settings", middlewareFunc(http.HandlerFunc(app.settingsHandler)))

	// Register the "/livelogs" route
	http.HandleFunc("GET /livelogs", func(w http.ResponseWriter, r *http.Request) {
		log.Print("livelogs")
		if app.DBHandler == nil {
			log.Print("No database configured")
			http.Error(w, "No database configured. Please configure the database.", http.StatusServiceUnavailable)
			return
		}

		handlers.LiveLogs(w, r, app.DBHandler)
	})

	// Start the HTTP server
	server := &http.Server{
		Addr:    ":8080",
		Handler: http.DefaultServeMux, // Explicitly set DefaultServeMux as the handler
	}

	return server.ListenAndServe()
}