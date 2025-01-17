package interfaces

import "time"

type LogEntry struct {
	ID        int       `db:"id"`         // Maps to PRIMARY KEY
	Timestamp time.Time `db:"timestamp"`  // Maps to timestamp
	Level     string    `db:"level"`      // Maps to level
	Message   string    `db:"message"`    // Maps to message
	Source    *string   `db:"source"`     // Maps to source (nullable)
	Method    *string   `db:"method"`     // Maps to method (nullable)
	Address   *string   `db:"address"`    // Maps to address (nullable)
	Length    *int      `db:"length"`     // Maps to length (nullable)
	Metadata  *string   `db:"metadata"`   // Maps to metadata (nullable)
}