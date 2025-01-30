package confighandler

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config represents the overall configuration structure
type Config struct {
	Version        string     `mapstructure:"version"`         // Always present
	LogLevel       string     `mapstructure:"log_level"`       // Always present
	LogFile        string     `mapstructure:"log_file"`        // Always present
	MaxConnections int        `mapstructure:"max_connections"` // Always present
	LogHandler     LogHandler `mapstructure:"log_handler"`     // Log handling configuration
	Database       Database   `mapstructure:"database"`        // Database configuration
}

type LogHandler struct {
	Mode   string `mapstructure:"mode"`   // "send" or "scrape"
	Send   Send   `mapstructure:"send"`   // Send configuration
	Scrape Scrape `mapstructure:"scrape"` // Scrape configuration
}

type Send struct {
	Protocol string `mapstructure:"protocol"` // "HTTP" or "UDP"
	Port     int    `mapstructure:"port"`     // number
}

type Scrape struct {
	Type string `mapstructure:"type"` // "pure_docker", "docker_swarm", "kubernetes"
}

type Database struct {
	Type           string `mapstructure:"type"`            // Currently only "SQLite"
	SQLiteFilepath string `mapstructure:"sqlite_filepath"` // Required if Type is "SQLite"
}

// LoadConfig loads the configuration from a file and applies defaults
func LoadConfig(configPath string) (Config, error) {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return Config{}, fmt.Errorf("config file does not exist: %s", configPath)
	}

	var config Config

	// Use Viper to read the config file
	viper.SetConfigFile(configPath) // Specify the exact file path
	viper.SetConfigType("yaml")     // Specify file type

	// Set default values for the config
	viper.SetDefault("version", "1.0.0")
	viper.SetDefault("log_level", "DEBUG")
	viper.SetDefault("log_file", "logs/app.log")
	viper.SetDefault("max_connections", 10)

	viper.SetDefault("log_handler.mode", "send") // Default to "send" mode
	viper.SetDefault("log_handler.send.protocol", "UDP")
	viper.SetDefault("log_handler.send.port", 2020)
	viper.SetDefault("log_handler.scrape.type", "pure_docker")

	viper.SetDefault("database.type", "SQLite")
	viper.SetDefault("database.sqlite_filepath", "./myDB.db")

	// Read the config file
	if err := viper.ReadInConfig(); err != nil {
		return config, fmt.Errorf("could not read config file: %w", err)
	}

	// Unmarshal the configuration into the struct
	if err := viper.Unmarshal(&config); err != nil {
		return config, fmt.Errorf("could not decode config: %w", err)
	}

	return config, nil
}

// ValidateConfig validates the loaded configuration
func ValidateConfig(config Config) error {
	// Validate log level
	validLogLevels := map[string]bool{"ALL": true, "ERROR": true, "WARNING": true, "DEBUG": true, "NONE": true}
	if !validLogLevels[config.LogLevel] {
		return fmt.Errorf("invalid log_level: %s (must be one of ALL, ERROR, WARNING, DEBUG, NONE)", config.LogLevel)
	}

	// Validate LogHandler mode
	if config.LogHandler.Mode != "send" && config.LogHandler.Mode != "scrape" {
		return fmt.Errorf("invalid log_handler mode: %s (must be send or scrape)", config.LogHandler.Mode)
	}

	// Validate send protocol
	if config.LogHandler.Mode == "send" {
		if config.LogHandler.Send.Protocol != "UDP" && config.LogHandler.Send.Protocol != "HTTP" {
			return fmt.Errorf("invalid protocol: %s (must be UDP or HTTP)", config.LogHandler.Send.Protocol)
		}
	}

	// Validate scrape type
	if config.LogHandler.Mode == "scrape" {
		validScrapeTypes := map[string]bool{"pure_docker": true, "docker_swarm": true, "kubernetes": true}
		if !validScrapeTypes[config.LogHandler.Scrape.Type] {
			return fmt.Errorf("invalid scrape type: %s (must be pure_docker, docker_swarm, or kubernetes)", config.LogHandler.Scrape.Type)
		}
	}

	// Validate max connections
	if config.MaxConnections <= 0 {
		return fmt.Errorf("max_connections must be greater than 0")
	}

	// Validate database type
	if config.Database.Type != "SQLite" {
		return fmt.Errorf("unsupported database type: %s (only SQLite is supported for now)", config.Database.Type)
	}

	// Validate SQLite filepath
	if config.Database.Type == "SQLite" && config.Database.SQLiteFilepath == "" {
		return fmt.Errorf("sqlite_filepath cannot be empty")
	}

	return nil
}

// PrintConfigTable prints the loaded configuration in a human-readable format
func PrintConfigTable(config Config) {
	fmt.Println("Loaded Configuration:")
	fmt.Printf("  Version          : %s\n", config.Version)
	fmt.Printf("  Log Level        : %s\n", config.LogLevel)
	fmt.Printf("  Log File         : %s\n", config.LogFile)
	fmt.Printf("  Max Connections  : %d\n", config.MaxConnections)

	fmt.Println("  Log Handler:")
	fmt.Printf("    Mode           : %s\n", config.LogHandler.Mode)
	if config.LogHandler.Mode == "send" {
		fmt.Printf("    Protocol       : %s\n", config.LogHandler.Send.Protocol)
		fmt.Printf("    Port       : %d\n", config.LogHandler.Send.Port)
	} else if config.LogHandler.Mode == "scrape" {
		fmt.Printf("    Type           : %s\n", config.LogHandler.Scrape.Type)
	}

	fmt.Println("  Database:")
	fmt.Printf("    Type           : %s\n", config.Database.Type)
	fmt.Printf("    SQLite Filepath: %s\n", config.Database.SQLiteFilepath)
}

func SaveConfig(config Config, filePath string) error {
	// Ensure the directory exists before writing the file
	dir := filepath.Dir(filePath)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return fmt.Errorf("error creating config directory: %v", err)
	}

	// Open (or create) the file
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("error opening config file: %v", err)
	}
	defer file.Close()

	// Set the file path where Viper should write the config
	viper.SetConfigFile(filePath)
	viper.SetConfigType("yaml") // Explicitly set the config type

	// Set the configuration values
	viper.Set("version", config.Version)
	viper.Set("log_level", config.LogLevel)
	viper.Set("log_file", config.LogFile)
	viper.Set("max_connections", config.MaxConnections)

	viper.Set("log_handler.mode", config.LogHandler.Mode)
	viper.Set("log_handler.send.protocol", config.LogHandler.Send.Protocol)
	viper.Set("log_handler.send.port", config.LogHandler.Send.Port)
	viper.Set("log_handler.scrape.type", config.LogHandler.Scrape.Type)

	viper.Set("database.type", config.Database.Type)
	viper.Set("database.sqlite_filepath", config.Database.SQLiteFilepath)

	// Write the config file
	if err := viper.WriteConfigAs(filePath); err != nil {
		return fmt.Errorf("error writing config file: %v", err)
	}

	fmt.Printf("Configuration saved to %s\n", filePath)
	return nil
}