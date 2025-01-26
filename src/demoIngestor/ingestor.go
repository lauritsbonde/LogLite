package demodata

import (
	"fmt"
	"log"
	"time"

	dbhandler "github.com/lauritsbonde/LogLite/src/dbHandler"
	"golang.org/x/exp/rand"
)

func IngestDemoData(db dbhandler.DBHandler, rowsPerSecond int) {
	// Initialize random seed
	rand.Seed(uint64(time.Now().UnixNano()))

	// Define the lists
	levels := []string{"ALL", "ERROR", "WARNING", "DEBUG", "NONE"}
	words := []string{"log", "logs", "debug", "trace", "error", "warning", "info", "audit", "event", "message"}
	httpVerbs := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS", "TRACE", "CONNECT"}

	// Calculate interval duration (rowsPerSecond -> duration in milliseconds)
	interval := time.Second / time.Duration(rowsPerSecond)

	// Create a ticker for the specified interval
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Run ingestion in a loop
	for range ticker.C {
		// Generate a random log entry
		data := map[string]interface{}{
			"level":    levels[rand.Intn(len(levels))],
			"message":  fmt.Sprintf("Random %s %s", words[rand.Intn(len(words))], words[rand.Intn(len(words))]),
			"source":   fmt.Sprintf("%s_service", words[rand.Intn(len(words))]),
			"method":   httpVerbs[rand.Intn(len(httpVerbs))],
			"address":  fmt.Sprintf("192.168.%d.%d", rand.Intn(256), rand.Intn(256)),
			"length":   rand.Intn(1024),
			"metadata": fmt.Sprintf(`{"key": "value%d"}`, rand.Intn(1000)),
			"label":    words[rand.Intn(len(words))],
		}

		// Insert the log entry into the database
		err := db.Put("logs", data)
		if err != nil {
			log.Printf("Error inserting demo log: %v", err)
			continue
		}
	}
}