package main

import (
	"flag"
	"fmt"
	"log"
	"sync"

	confighandler "github.com/lauritsbonde/LogLite/src/configHandler"
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

func main() {
	var wg sync.WaitGroup

	// check if we have a config
	confighandler.PrintConfigTable(loadedConfig)

	// var dbHandler dbhandler.DBHandler

	log.Printf("version: %d\n", len(loadedConfig.Version))
	if len(loadedConfig.Version) == 0 {
		// adding the webapp
		wg.Add(1)
		webApp := &webapp.WebApp{
			DBHandler: nil,
			SettingsChan: make(chan string, 1),
			Configuration: &confighandler.Config{},
		}

		go func() {
			defer wg.Done()
			if err := webApp.RunWebApp(); err != nil {
				fmt.Printf("Error starting web server: %v\n", err)
			}
		}()
	}

	// this is for demo data ingestion
	wg.Wait()
	// demodata.IngestDemoData(dbHandler, 10)
}
