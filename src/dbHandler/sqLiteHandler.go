package dbhandler

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

type SQLiteHandler struct {
	db *sql.DB
}

// NewSQLiteHandler initializes and returns a new SQLiteHandler
func NewSQLiteHandler(dbFile string) (*SQLiteHandler, error) {
	// Ensure the directory for the database file exists
	dir := filepath.Dir(dbFile)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory for SQLite file: %w", err)
		}
	}

	// Open the SQLite database
	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SQLite: %w", err)
	}

	// Create a new SQLiteHandler instance
	handler := &SQLiteHandler{db: db}

	// Initialize the database (create tables and indexes if necessary)
	if err := Initialize(handler); err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	return handler, nil
}

// create the Log Table
func Initialize(h *SQLiteHandler) error{
	// SQL to create the logs table
	createTableQuery := `
	CREATE TABLE IF NOT EXISTS logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		level TEXT NOT NULL,
		message TEXT NOT NULL,
		source TEXT,
		method TEXT,
		address TEXT,
		length INTEGER,
		metadata TEXT,
		label TEXT
	);
	`

	// Execute the table creation query
	_, err := h.db.Exec(createTableQuery)
	if err != nil {
		return fmt.Errorf("failed to create logs table: %w", err)
	}

	// Create indexes for faster querying
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_logs_level ON logs(level);",
		"CREATE INDEX IF NOT EXISTS idx_logs_timestamp ON logs(timestamp);",
		"CREATE INDEX IF NOT EXISTS idx_logs_source ON logs(source);",
	}

	for _, indexQuery := range indexes {
		_, err := h.db.Exec(indexQuery)
		if err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	return nil
}

// Put inserts data into the specified table
func (h *SQLiteHandler) Put(table string, data map[string]interface{}) error {
	// Build the INSERT query
	columns := []string{}
	values := []interface{}{}
	placeholders := []string{}

	for col, val := range data {
		columns = append(columns, col)
		values = append(values, val)
		placeholders = append(placeholders, "?")
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		table,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	_, err := h.db.Exec(query, values...)
	if err != nil {
		return fmt.Errorf("failed to insert data into %s: %w", table, err)
	}

	return nil
}

// Ger retrieves data from the specified table
func (h *SQLiteHandler) Get(table string, conditions map[string]interface{}) ([]map[string]interface{}, error) {
	query := fmt.Sprintf("SELECT * FROM %s", table)
	values := []interface{}{}
	limit := ""
	offset := ""
	orderBy := ""

	if len(conditions) > 0 {
		conditionList := []string{}
		for col, val := range conditions {
			switch strings.ToLower(col) {
			case "limit":
				limit = fmt.Sprintf(" LIMIT %d", val)
			case "offset":
				offset = fmt.Sprintf(" OFFSET %d", val)
			case "orderby":
				// Safely handle the ORDER BY clause
				orderBy = fmt.Sprintf(" ORDER BY %s DESC", val)
			default:
				conditionList = append(conditionList, fmt.Sprintf("%s = ?", col))
				values = append(values, val)
			}
		}

		// Add WHERE clause if conditions exist
		if len(conditionList) > 0 {
			query += " WHERE " + strings.Join(conditionList, " AND ")
		}
	}

	// Append LIMIT and OFFSET clauses
	query += orderBy + limit + offset

	println(query)

	// execute the query
	rows, err := h.db.Query(query, values...)
	if err != nil {
		return nil, fmt.Errorf("failed to query data from %s: %w", table, err)
	}
	defer rows.Close()

	// parse the rows into a slice of maps
	var results []map[string]interface{}
	columns, _ := rows.Columns()

	for rows.Next() {
		rowData := make([]interface{}, len(columns))
		rowPointers := make([]interface{}, len(columns))

		for i := range rowData {
			rowPointers[i] = &rowData[i]
		}

		if err := rows.Scan(rowPointers...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		rowMap := make(map[string]interface{})
		for i, col := range columns {
			rowMap[col] = rowData[i]
		}

		results = append(results, rowMap)
	}

	return results, nil
}


func (h *SQLiteHandler) Close() error {
	if h.db != nil {
		if err := h.db.Close(); err != nil {
			return fmt.Errorf("failed to close SQLite database: %w", err)
		}
	}
	return nil
}