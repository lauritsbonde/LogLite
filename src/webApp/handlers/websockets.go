package handlers

import (
	"bytes"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	dbhandler "github.com/lauritsbonde/LogLite/src/dbHandler"
	"github.com/lauritsbonde/LogLite/src/webApp/components"
)

var upgrader = websocket.Upgrader{}

func LiveLogs(w http.ResponseWriter, r *http.Request, db dbhandler.DBHandler) {
	println("LiveLogs")
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("websocket upgrader: ", err)
		return
	}
	defer c.Close()

	// Continuously fetch logs and send to the client
	for {
		println("getting logs")
		logs, err := GetLogs(db)
		if err != nil {
			log.Println("Error getting logs from DB:", err)
			time.Sleep(1 * time.Second) // Avoid tight looping on error
			continue
		}

		// Render log entries into HTML
		var buffer bytes.Buffer
		for _, logEntry := range logs {
			component := components.LiveLogEntry(logEntry)
			err := component.Render(r.Context(), &buffer)
			if err != nil {
				log.Println("Error rendering log entry:", err)
				continue // Skip this entry and continue with others
			}
		}

		// Send rendered logs to the WebSocket client
		err = c.WriteMessage(websocket.TextMessage, buffer.Bytes())
		if err != nil {
			log.Println("WebSocket write error:", err)
			break
		}

		// Sleep for some time before fetching the next batch
		time.Sleep(1 * time.Second)
	}
}