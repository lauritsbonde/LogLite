package ingestor

import (
	"fmt"

	confighandler "github.com/lauritsbonde/LogLite/src/configHandler"
	dbhandler "github.com/lauritsbonde/LogLite/src/dbHandler"
)

type Ingestor interface {
	Start() error
	Stop() error

	// DB handler to save the data
	SetDBHandler(dbhandler.DBHandler)
}

func NewIngestor(config *confighandler.Config, dbHandler dbhandler.DBHandler) (Ingestor, error) {
	switch config.LogHandler.Mode {
		case "send":
			switch config.LogHandler.Send.Protocol {
				case "HTTP":
					ingestor := &HTTPIngestor{Port: config.LogHandler.Send.Port}
					ingestor.SetDBHandler(dbHandler) // Inject the DBHandler
					return ingestor, nil
				case "UDP":
					ingestor := &UDPIngestor{Port: config.LogHandler.Send.Port}
					ingestor.SetDBHandler(dbHandler) // Inject the DBHandler
					return ingestor, nil
				default:
					return nil, fmt.Errorf("unsupported protocol: %s (must be 'HTTP' or 'UDP')", config.LogHandler.Send.Protocol)
			}
		
		case "scrape": {
			return nil, fmt.Errorf("not implemented yet")
		}
	}

	return nil, fmt.Errorf("something went wrong")
}