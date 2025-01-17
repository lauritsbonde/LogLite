package demodata

import (
	"fmt"
	"log"
	"time"

	dbhandler "github.com/lauritsbonde/LogLite/src/dbHandler"
	"golang.org/x/exp/rand"
)

func IngestDemoData(db dbhandler.DBHandler, rowsPerSecond int) {
	// Initialize random seed
	rand.Seed(uint64(time.Now().UnixNano()))

	// Define the lists
	// minDelay := 10 // ms
	levels := []string{"ALL", "ERROR", "WARNING", "DEBUG", "NONE"}
	words := []string{"log", "logs", "logging", "debug", "trace", "error", "warning", "info", "audit", "event", "message", "alert", "metric", "data", "session", "request", "response", "user", "timestamp", "context", "status", "failure", "success", "exception", "level", "process", "monitor", "entry", "record", "output", "system", "network", "server", "client", "thread", "operation", "details", "transaction", "id", "identifier", "path", "endpoint", "duration", "elapsed", "start", "end", "init", "shutdown", "retry", "attempt", "priority", "severity", "code", "type", "source", "destination", "hostname", "ip", "address", "port", "protocol", "method", "stacktrace", "traceback", "payload", "body", "headers", "metadata", "label", "tag", "category", "action", "target", "result", "validation", "query", "route", "handler", "function", "module", "library", "component", "container", "threadpool", "worker", "instance", "sessionId", "connection", "stream", "buffer", "file", "directory", "access", "permission", "read", "write", "append", "update", "delete", "insert", "fetch", "cache", "memory", "disk", "storage", "quota", "limit", "threshold", "heartbeat", "polling", "retry", "attempt", "eventId", "messageId", "operationId", "debugId", "contextId", "dataId", "region", "zone", "service", "namespace", "cluster", "deployment", "job", "task", "workerId", "node", "machine", "host", "processId", "pid", "tid", "username", "role", "privileges", "schema", "table", "column", "row", "recordId", "api", "endpointId", "latency", "throughput", "rate", "frequency", "interval", "timeout", "expiry", "expiryTime", "retries", "maxRetries", "delay", "startTime", "endTime", "createdAt", "updatedAt", "deletedAt", "messageBody", "responseBody", "requestBody", "errorCode", "errorMessage", "responseCode", "responseMessage", "statusCode", "statusMessage", "config", "settings", "options", "params", "attributes", "fields", "values", "keys", "pairs", "entries", "variables", "constants", "exceptions", "errors", "failures", "successes", "warnings", "information", "alerts", "metrics", "traces", "logs"}
	httpVerbs := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS", "TRACE", "CONNECT"}

	// Generate a random log entry
	data := map[string]interface{}{
		"level":    levels[rand.Intn(len(levels))],
		"message":  fmt.Sprintf("Random %s %s", words[rand.Intn(len(words))], words[rand.Intn(len(words))]),
		"source":   fmt.Sprintf("%s_service", words[rand.Intn(len(words))]),
		"method":   httpVerbs[rand.Intn(len(httpVerbs))],
		"address":  fmt.Sprintf("192.168.%d.%d", rand.Intn(256), rand.Intn(256)),
		"length":   rand.Intn(1024),
		"metadata": fmt.Sprintf(`{"key": "value%d"}`, rand.Intn(1000)),
		"label":    words[rand.Intn(len(words))],
	}

	err := dbhandler.DBHandler.Put(db, "logs", data)
	if err != nil {
		log.Fatal("error inserting demo logs", err)
	}
	println("ingested")

}