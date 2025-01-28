package ingestor

import (
	"fmt"
	"log"
	"net"
	"sync"

	dbhandler "github.com/lauritsbonde/LogLite/src/dbHandler"
)

type UDPIngestor struct {
	Port     int
	listener net.PacketConn // Listener for the UDP server
	stopChan chan struct{}  // Channel to signal stop
	wg       sync.WaitGroup // WaitGroup to ensure goroutines complete on Stop
	dbHandler dbhandler.DBHandler // Database handler to save data
}

// Start begins the UDP server
func (u *UDPIngestor) Start() error {
	println("Starting UDP Ingestor")
	// Start listening on the specified UDP port
	address := fmt.Sprintf(":%d", u.Port)
	pc, err := net.ListenPacket("udp", address)
	if err != nil {
		return fmt.Errorf("failed to start UDP server: %w", err)
	}
	u.listener = pc // Save the listener for future reference
	u.stopChan = make(chan struct{}) // Initialize the stop signal channel

	log.Printf("UDP server is running on port %d\n", u.Port)

	// Use a sync.Pool to reuse buffers, storing pointers to slices
	var bufferPool = sync.Pool{
		New: func() interface{} {
			buf := make([]byte, 1024)
			return &buf // Return a pointer to the slice
		},
	}

	// Start listening for incoming UDP packets in a goroutine
	u.wg.Add(1)
	go func() {
		defer u.wg.Done()

		for {
			select {
			case <-u.stopChan:
				// Stop the server when a stop signal is received
				log.Println("Stopping UDP server...")
				return
			default:
				// Get a buffer from the pool
				bufPtr := bufferPool.Get().(*[]byte)

				// Read incoming data
				n, addr, err := pc.ReadFrom(*bufPtr)
				if err != nil {
					if isNetClosedError(err) {
						// Graceful shutdown, exit the loop
						return
					}
					log.Printf("Error reading from UDP connection: %v", err)
					bufferPool.Put(bufPtr) // Return the buffer pointer to the pool
					continue
				}

				// Process the request in a goroutine
				u.wg.Add(1)
				go func(bufPtr *[]byte, n int, addr net.Addr) {
					defer u.wg.Done()   // Mark goroutine as done
					defer bufferPool.Put(bufPtr) // Return buffer to pool

					// Handle the UDP message
					u.handleRequest(pc, addr, (*bufPtr)[:n])
				}(bufPtr, n, addr)
			}
		}
	}()

	return nil
}

// handleRequest processes a single UDP request
func (u *UDPIngestor) handleRequest(pc net.PacketConn, addr net.Addr, buf []byte) {
	message := string(buf)
	log.Printf("Received %d bytes from %s: %s\n", len(buf), addr.String(), message)

	// Prepare the log data for insertion
	logData := map[string]interface{}{
		"level":    "INFO",          // Example: You can set a default level
		"message":  message,         // The content of the UDP message
		"source":   "udp-ingestor",  // Identify the source of the log
		"method":   nil,             // UDP doesn't have methods, so set this to NULL
		"address":  addr.String(),   // The client's address
		"length":   len(buf),        // The length of the UDP message
		"metadata": nil,             // Add any extra metadata if needed, or leave it NULL
	}

	// Insert the log into the database
	if u.dbHandler != nil {
		err := u.dbHandler.Put("logs", logData)
		if err != nil {
			log.Printf("Error saving log to database: %v", err)
		}
	}

	// Send a response back to the client
	response := fmt.Sprintf("Echo: %s", message)
	_, err := pc.WriteTo([]byte(response), addr)
	if err != nil {
		log.Printf("Error writing to UDP client %s: %v", addr.String(), err)
	}
}

// Stop gracefully stops the UDP server
func (u *UDPIngestor) Stop() error {
	if u.listener != nil {
		// Signal to stop the server
		close(u.stopChan)

		// Close the listener to unblock ReadFrom
		if err := u.listener.Close(); err != nil {
			return fmt.Errorf("failed to close UDP server: %w", err)
		}
		log.Println("UDP server stopped")
	}

	// Wait for all goroutines to finish
	u.wg.Wait()
	return nil
}

// Helper function to check if an error is due to the listener being closed
func isNetClosedError(err error) bool {
	// This is a common way to detect "use of closed network connection"
	return err != nil && (err.Error() == "use of closed network connection")
}

func (u *UDPIngestor) SetDBHandler(dbHandler dbhandler.DBHandler) {
	u.dbHandler = dbHandler
}