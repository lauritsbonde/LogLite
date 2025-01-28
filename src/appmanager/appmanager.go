package appmanager

import (
	dbhandler "github.com/lauritsbonde/LogLite/src/dbHandler"
	"github.com/lauritsbonde/LogLite/src/ingestor"
)

type AppManager struct {
	DBHandler dbhandler.DBHandler
	Ingestor  ingestor.Ingestor
}

func NewAppManager() *AppManager {
	return &AppManager{}
}

// Option is a function that modifies an AppManager instance
type Option func(*AppManager)

// BindDBHandler is a self-referential function that injects a DBHandler into AppManager
func BindDBHandler(v dbhandler.DBHandler) Option {
    return func(a *AppManager) {
        a.DBHandler = v
    }
}

// BindIngestor is a self-referential function that injects an Ingestor into AppManager
func BindIngestor(v ingestor.Ingestor) Option {
    return func(a *AppManager) {
        a.Ingestor = v
    }
}