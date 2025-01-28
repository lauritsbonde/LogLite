package main

import (
	"flag"
	"fmt"
	"log"
	"sync"

	"github.com/lauritsbonde/LogLite/src/appmanager"
	confighandler "github.com/lauritsbonde/LogLite/src/configHandler"
	dbhandler "github.com/lauritsbonde/LogLite/src/dbHandler"
	"github.com/lauritsbonde/LogLite/src/ingestor"
	webapp "github.com/lauritsbonde/LogLite/src/webApp"
)

var loadedConfig confighandler.Config

func init(){
	// Command-line flag for config file
	configPathFlag := flag.String("config", "", "Path to the configuration file")
	flag.Parse()

	configpath := *configPathFlag

	if len(configpath) == 0 {
		println("No config provided")
		return
	}

	// Load the configuration
	config, err := confighandler.LoadConfig(configpath)
	if err != nil {
		log.Fatalf("Error loading config: %v\n", err)
	}

	// Print loaded configuration (for debugging)
	confighandler.PrintConfigTable(config)

	if err := confighandler.ValidateConfig(config); err != nil {
		log.Fatalf("Invalid configuration: %v\n", err)
	}

	loadedConfig = config
}

func messageHandler(webApp *webapp.WebApp, appManager *appmanager.AppManager, ingestorReady chan ingestor.Ingestor) {
	for msg := range webApp.SettingsChan {
		log.Println("Main thread: Received new configuration")
		var err error

		// Validate the configuration
		err = confighandler.ValidateConfig(*msg.NewConfig)
		if err != nil {
			log.Printf("main thread could not validate the configuration: %v\n", err)
			msg.ResponseCh <- fmt.Sprintf("Validation error: %v", err)
			continue
		}

		// Apply the appropriate DBHandler
		var dbHandler dbhandler.DBHandler
		switch msg.NewConfig.Database.Type {
		case "SQLite":
			dbHandler, err = dbhandler.NewSQLiteHandler(msg.NewConfig.Database.SQLiteFilepath)
			if err != nil {
				log.Printf("Error initializing SQLite handler: %v\n", err)
				msg.ResponseCh <- fmt.Sprintf("Error initializing SQLite handler: %v", err)
				continue
			}
		default:
			msg.ResponseCh <- "Error: Unsupported database type"
			continue
		}

		// Apply the appropriate Ingestor using the NewIngestor function
		var ing ingestor.Ingestor
		switch msg.NewConfig.LogHandler.Mode {
		case "send":
			ing, err = ingestor.NewIngestor(msg.NewConfig, dbHandler)
			if err != nil {
				println("Error initializing Ingestor")
				msg.ResponseCh <- fmt.Sprintf("Error initializing Ingestor: %v", err)
				continue
			}
		case "scrape":
			// Handle scrape-specific cases (if required)
			println("Scrape Ingestor not implemented")
			msg.ResponseCh <- "Scrape Ingestor not implemented"
			continue
		default:
			println("Error: Unsupported collect type")
			msg.ResponseCh <- "Error: Unsupported collect type"
			continue
		}

		// Dynamically bind the DBHandler and Ingestor to the AppManager
		appManager.DBHandler = dbHandler
		appManager.Ingestor = ing

		ingestorReady <- ing

		// Update the WebApp configuration
		webApp.Configuration = msg.NewConfig

		// Respond to the sender
		msg.ResponseCh <- "Configuration applied successfully"
		println("Configuration applied successfully")
	}
}

func main() {
	var wg sync.WaitGroup

	// check if we have a config
	confighandler.PrintConfigTable(loadedConfig)

	appManager := appmanager.NewAppManager()

	// ingestor ready channel
	ingestorReady := make(chan ingestor.Ingestor)

	// var dbHandler dbhandler.DBHandler
	log.Printf("version: %d\n", len(loadedConfig.Version))

	if len(loadedConfig.Version) == 0 {
		// adding the webapp
		wg.Add(1)
		webApp := &webapp.WebApp{
			DBHandler: nil,
			SettingsChan: make(chan webapp.ConfigMessage, 1),
			Configuration: &confighandler.Config{},
		}

		go messageHandler(webApp, appManager, ingestorReady)

		go func() {
			defer wg.Done()
			if err := webApp.RunWebApp(); err != nil {
				fmt.Printf("Error starting web server: %v\n", err)
			}
		}()
	}

	go func() {
		defer wg.Done()

		ing := <- ingestorReady
		println("Starting ingestor")
		if err := ing.Start(); err != nil {
			log.Fatalf("Error starting ingestor: %v", err)
		}
	}()

	wg.Wait()
}
