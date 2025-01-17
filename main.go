package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	confighandler "github.com/lauritsbonde/LogLite/src/configHandler"
	dbhandler "github.com/lauritsbonde/LogLite/src/dbHandler"
	demodata "github.com/lauritsbonde/LogLite/src/demoIngestor"
	"github.com/lauritsbonde/LogLite/src/ingestor"
	webapp "github.com/lauritsbonde/LogLite/src/webApp"
)

var loadedConfig confighandler.Config

func init(){
	// Command-line flag for config file
	configPath := flag.String("config", "config.yaml", "Path to the configuration file")
	flag.Parse()

	// Load the configuration
	config, err := confighandler.LoadConfig(*configPath)
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Print loaded configuration (for debugging)
	confighandler.PrintConfigTable(config)

	if err := confighandler.ValidateConfig(config); err != nil {
		fmt.Printf("Invalid configuration: %v\n", err)
		os.Exit(1)
	}

	loadedConfig = config
}

func main() {
	// Set up the ingestor and DBHandler
	ingestorInstance, dbHandler, webappRunner, err := setup(loadedConfig)
	if err != nil {
		log.Fatalf("Setup failed: %v", err)
	}

	// Create channels for error and shutdown handling
	errorChan := make(chan error, 1)
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)

	// Start the ingestor in a goroutine
	go func() {
		fmt.Printf("Starting %s ingestor on port %d...\n", loadedConfig.Ingestor.Protocol, loadedConfig.Ingestor.Port)
		if err := ingestorInstance.Start(); err != nil {
			errorChan <- fmt.Errorf("failed to start ingestor: %w", err)
		}
	}()


	// Start the web app in a goroutine
	go func() {
		if err := webappRunner(); err != nil {
			errorChan <- fmt.Errorf("web app error: %w", err)
		}
	}()

	// this is for demo data ingestion
	demodata.IngestDemoData(dbHandler, 10)

	// Block until an error or shutdown signal is received
	blockUntilShutdown(errorChan, shutdownChan, ingestorInstance, dbHandler)
}

// setup initializes the ingestor, DBHandler, and web app based on the configuration
func setup(config confighandler.Config) (ingestor.Ingestor, dbhandler.DBHandler, func() error, error) {
	// Create the database handler
	var dbHandler dbhandler.DBHandler

	switch config.Database.Type {
	case "SQLite":
		// Create the SQLiteHandler instance
		sqliteHandler, err := dbhandler.NewSQLiteHandler(config.Database.SQLiteFilepath)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("failed to create SQLite handler: %w", err)
		}

		// Assign the concrete SQLiteHandler to the DBHandler interface
		dbHandler = sqliteHandler
	default:
		return nil, nil, nil, fmt.Errorf("unsupported database type: %s", config.Database.Type)
	}

	// Create the ingestor instance
	ingestorInstance, err := ingestor.NewIngestor(config.Ingestor.Protocol, config.Ingestor.Port, dbHandler)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create ingestor: %w", err)
	}

	// Create the web app runner function
	webAppRunner := func() error {
		fmt.Println("Starting web app on port 8080...")
		return webapp.RunWebApp(dbHandler) // Pass dbHandler to the web app
	}

	return ingestorInstance, dbHandler, webAppRunner, nil
}

// blockUntilShutdown handles shutdown signals and ingestor/DB cleanup
func blockUntilShutdown(errorChan <-chan error, shutdownChan <-chan os.Signal, ingestorInstance ingestor.Ingestor, dbHandler dbhandler.DBHandler) {
	for {
		select {
		case err := <-errorChan:
			if err != nil {
				log.Fatalf("Ingestor error: %v", err)
			}
			return // Exit the loop if the error channel closes
		case sig := <-shutdownChan:
			fmt.Printf("\nReceived signal: %v. Shutting down...\n", sig)

			// Stop the ingestor
			if err := ingestorInstance.Stop(); err != nil {
				log.Printf("Error stopping ingestor: %v", err)
			} else {
				log.Println("Ingestor stopped gracefully.")
			}

			// Perform any necessary database cleanup (if applicable)
			if err := dbHandler.Close(); err != nil {
				log.Printf("Error closing database handler: %v", err)
			} else {
				log.Println("Database handler closed gracefully.")
			}

			return // Exit the loop on shutdown signal
		}
	}
}