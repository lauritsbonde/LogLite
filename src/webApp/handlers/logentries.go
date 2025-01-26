package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	dbhandler "github.com/lauritsbonde/LogLite/src/dbHandler"
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
	m["limit"] = 1
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

func optionalString(strPtr *string) string {
	if strPtr != nil {
		return *strPtr
	}
	return "N/A"
}

func optionalInt(intPtr *int) int {
	if intPtr != nil {
		return *intPtr
	}
	return 0
}

func PrintLgos(logEntries []interfaces.LogEntry) {
	for _, entry := range logEntries {
		fmt.Printf("Timestamp: %s, Level: %s, Message: %s, Source: %s, Method: %s, Address: %s, Length: %d, Metadata: %s\n",
			entry.Timestamp.String(),
			entry.Level,
			entry.Message,
			optionalString(entry.Source),
			optionalString(entry.Method),
			optionalString(entry.Address),
			optionalInt(entry.Length),
			optionalString(entry.Metadata),
		)
	}
}
