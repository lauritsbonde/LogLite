package ingestor

import (
	"fmt"

	dbhandler "github.com/lauritsbonde/LogLite/src/dbHandler"
)

type Ingestor interface {
	Start() error
	Stop() error

	// DB handler to save the data
	SetDBHandler(dbhandler.DBHandler)
}

func NewIngestor(protocol string, port int, dbHandler dbhandler.DBHandler) (Ingestor, error) {
	switch protocol {
	case "HTTP":
		ingestor := &HTTPIngestor{Port: port}
		ingestor.SetDBHandler(dbHandler) // Inject the DBHandler
		return ingestor, nil
	case "UDP":
		ingestor := &UDPIngestor{Port: port}
		ingestor.SetDBHandler(dbHandler) // Inject the DBHandler
		return ingestor, nil
	default:
		return nil, fmt.Errorf("unsupported protocol: %s (must be 'HTTP' or 'UDP')", protocol)
	}
}