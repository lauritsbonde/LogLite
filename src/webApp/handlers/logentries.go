package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	dbhandler "github.com/lauritsbonde/LogLite/src/dbHandler"
	"github.com/lauritsbonde/LogLite/src/webApp/components"
	"github.com/lauritsbonde/LogLite/src/webApp/interfaces"
)

func ConvertToLogEntries(results []map[string]interface{}) ([]interfaces.LogEntry, error) {
	var logEntries []interfaces.LogEntry

	for _, row := range results {
		entry := interfaces.LogEntry{}

		// Parse ID
		if idValue, ok := row["id"]; ok {
			switch id := idValue.(type) {
			case int:
				entry.ID = id
			case int64:
				entry.ID = int(id)
			case float64:
				entry.ID = int(id)
			default:
				return nil, fmt.Errorf("unsupported id type: %T", idValue)
			}
		} else {
			return nil, errors.New("missing or invalid field: id")
		}

		// Parse timestamp
		if timestampValue, ok := row["timestamp"]; ok {
			switch t := timestampValue.(type) {
			case time.Time:
					// If it's already a time.Time object, use it directly
					entry.Timestamp = t
			case string:
					// If it's a string, parse it
					timestamp, err := time.Parse("2006-01-02 15:04:05", t)
					if err != nil {
							return nil, fmt.Errorf("invalid timestamp format: %v", err)
					}
					entry.Timestamp = timestamp
			default:
					return nil, fmt.Errorf("unsupported timestamp type: %T", t)
			}
		} else {
				return nil, errors.New("missing or invalid field: timestamp")
		}

		// Parse level
		if level, ok := row["level"].(string); ok {
				entry.Level = level
		} else {
				return nil, errors.New("missing or invalid field: level")
		}

		// Parse message
		if message, ok := row["message"].(string); ok {
				entry.Message = message
		} else {
				return nil, errors.New("missing or invalid field: message")
		}

		// Parse optional fields
		if source, ok := row["source"].(string); ok {
				entry.Source = &source
		}
		if method, ok := row["method"].(string); ok {
				entry.Method = &method
		}
		if address, ok := row["address"].(string); ok {
				entry.Address = &address
		}
		if length, ok := row["length"].(int); ok {
				entry.Length = &length
		}

		// Parse metadata as JSON
		if metadataStr, ok := row["metadata"].(string); ok {
			var metadataMap map[string]interface{}
			if err := json.Unmarshal([]byte(metadataStr), &metadataMap); err != nil {
					return nil, fmt.Errorf("invalid metadata JSON: %v", err)
			}
			metadataJSON, err := json.Marshal(metadataMap)
			if err != nil {
					return nil, fmt.Errorf("error marshaling metadata: %v", err)
			}
			metadataString := string(metadataJSON)
			entry.Metadata = &metadataString // Assign pointer to string
		}

		// Append the populated LogEntry
		logEntries = append(logEntries, entry)
	}

	return logEntries, nil
}

func GetLogs(dbhandler dbhandler.DBHandler) ([]interfaces.LogEntry, error) {
	// Fetch data from the database
	var m map[string]interface{} = make(map[string]interface{})
	m["limit"] = 10
	m["orderBy"] = "timestamp"
	dbRes, err := dbhandler.Get("logs", m)
	if err != nil {
			return nil, fmt.Errorf("error getting logs from database: %w", err)
	}

	// Convert database results to LogEntry structs
	entries, err := ConvertToLogEntries(dbRes)
	if err != nil {
			return nil, fmt.Errorf("error converting logs to LogEntries: %w", err)
	}

	return entries, nil
}

func GetPaginatedLogs(w http.ResponseWriter, r *http.Request, db dbhandler.DBHandler) {
	println("get paginated logs")
	// Parse query parameters
	pageStr := r.URL.Query().Get("page")
	limit := 10 // Items per page
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	offset := (page - 1) * limit

	// Query conditions
	m := map[string]interface{}{
		"offset":  offset,
		"limit":   limit,
		"orderBy": "timestamp",
	}

	// Fetch data from the database
	dbRes, err := db.Get("logs", m)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting logs: %v", err), http.StatusInternalServerError)
		return
	}

	// Convert database results to LogEntries
	entries, err := ConvertToLogEntries(dbRes)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error converting logs: %v", err), http.StatusInternalServerError)
		return
	}

	// Render the LogTable component
	component := components.LogTable(entries)

	ctx := r.Context()

	// Write the component to the HTTP response
	if err := component.Render(ctx, w); err != nil {
		http.Error(w, fmt.Sprintf("Error rendering logs: %v", err), http.StatusInternalServerError)
	}
}