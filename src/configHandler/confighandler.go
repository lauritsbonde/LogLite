package confighandler

import (
	"fmt"

	"github.com/spf13/viper"
)

// Config represents the overall configuration structure
type Config struct {
	LogLevel       string `mapstructure:"log_level"`       // Default: "DEBUG"
	LogFile        string `mapstructure:"log_file"`        // Default: "logs/app.log"
	MaxConnections int    `mapstructure:"max_connections"` // Default: 10

	Ingestor struct {
		Protocol string `mapstructure:"protocol"` // Default: "UDP"
		Port     int    `mapstructure:"port"`     // Default: 1053
	} `mapstructure:"ingestor"`

	Database struct {
		Type           string `mapstructure:"type"`            // Default: "SQLite"
		SQLiteFilepath string `mapstructure:"sqlite_filepath"` // Default: "./myDB.db"
	} `mapstructure:"database"`
}

// LoadConfig loads the configuration from a file and applies defaults
func LoadConfig(configPath string) (Config, error) {
	var config Config

	// Use Viper to read the config file
	viper.SetConfigFile(configPath) // Specify the exact file path
	viper.SetConfigType("yaml")     // Specify file type

	// Set default values for the config
	viper.SetDefault("log_level", "DEBUG")
	viper.SetDefault("log_file", "logs/app.log")
	viper.SetDefault("max_connections", 10)

	viper.SetDefault("ingestor.protocol", "UDP")
	viper.SetDefault("ingestor.port", 1053)

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

	// Validate protocol
	if config.Ingestor.Protocol != "UDP" && config.Ingestor.Protocol != "HTTP" {
		return fmt.Errorf("invalid protocol: %s (must be UDP or HTTP)", config.Ingestor.Protocol)
	}

	// Validate port range
	if config.Ingestor.Port < 1 || config.Ingestor.Port > 65535 {
		return fmt.Errorf("invalid port: %d (must be between 1 and 65535)", config.Ingestor.Port)
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
	if config.Database.SQLiteFilepath == "" {
		return fmt.Errorf("sqlite_filepath cannot be empty")
	}

	return nil
}

// PrintConfigTable prints the loaded configuration in a human-readable format
func PrintConfigTable(config Config) {
	fmt.Println("Loaded Configuration:")
	fmt.Printf("  Log Level        : %s\n", config.LogLevel)
	fmt.Printf("  Log File         : %s\n", config.LogFile)
	fmt.Printf("  Max Connections  : %d\n", config.MaxConnections)

	fmt.Println("  Ingestor:")
	fmt.Printf("    Protocol       : %s\n", config.Ingestor.Protocol)
	fmt.Printf("    Port           : %d\n", config.Ingestor.Port)

	fmt.Println("  Database:")
	fmt.Printf("    Type           : %s\n", config.Database.Type)
	fmt.Printf("    SQLite Filepath: %s\n", config.Database.SQLiteFilepath)
}
