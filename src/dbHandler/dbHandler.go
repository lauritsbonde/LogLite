package dbhandler

import (
	"fmt"

	confighandler "github.com/lauritsbonde/LogLite/src/configHandler"
)

type DBHandler interface {
	Put(table string, data map[string]interface{}) error
	Get(table string, conditions map[string]interface{}) ([]map[string]interface{}, error)
	Close() error
}

func NewDBHandler(config *confighandler.Config) (DBHandler, error) {
	var err error
	var dbHandler DBHandler
	switch config.Database.Type {
	case "SQLite":
		dbHandler, err = NewSQLiteHandler(config.Database.SQLiteFilepath)
		if err != nil {
			return nil, fmt.Errorf("error initializing SQLite handler: %v", err)
		}
	default:
		return nil, fmt.Errorf("unsupported database type %s", config.Database.Type)
	}
	return dbHandler, nil
}