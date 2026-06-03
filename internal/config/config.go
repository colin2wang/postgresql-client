package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/colin2wang/postgresql-client/internal/commons"
	"gopkg.in/yaml.v3"
)

const DefaultPort = 5432

// ImportConfig holds directory paths for import operations
type ImportConfig struct {
	DDLDir string `yaml:"ddl"`
	CSVDir string `yaml:"csv"`
	SQLDir string `yaml:"sql"`
}

// DefaultImportConfig returns default import directories
func DefaultImportConfig() ImportConfig {
	return ImportConfig{
		DDLDir: "ddl",
		CSVDir: "csv",
		SQLDir: "sql",
	}
}

// Config stores database connection and import configuration
type Config struct {
	Host     string       `yaml:"host"`
	Port     int          `yaml:"port"`
	User     string       `yaml:"user"`
	Password string       `yaml:"password"`
	Database string       `yaml:"database"`
	SSLMode  string       `yaml:"ssl_mode,omitempty"`
	Import   ImportConfig `yaml:"import"`
}

// NewConfig creates a new configuration
func NewConfig() *Config {
	defaultImport := DefaultImportConfig()
	return &Config{
		Host:     "localhost",
		Port:     DefaultPort,
		User:     "postgres",
		Database: "postgres",
		SSLMode:  "disable",
		Import:   defaultImport,
	}
}

// DefaultConfig returns default configuration values
func DefaultConfig() *Config {
	return &Config{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Database: "postgres",
		SSLMode:  "disable",
		Import:   DefaultImportConfig(),
	}
}

// LoadConfig loads configuration from a YAML file
func LoadConfig(filePath string) (*Config, error) {
	commons.DefaultLogger.Debug("Loading configuration from file: %s", filePath)
	config := DefaultConfig()

	// Check for default config file if no path provided
	if filePath == "" {
		if fileExists("config.yaml") {
			filePath = "config.yaml"
		} else if fileExists("config.yml") {
			filePath = "config.yml"
		}
	}

	// If no file path provided or file doesn't exist, use environment variables
	if filePath == "" || !fileExists(filePath) {
		return loadFromEnv(config), nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		commons.DefaultLogger.Error("Failed to read config file: %v", err)
		return nil, &commons.FileError{
			Message:     "failed to read config file",
			Path:        filePath,
			Action:      "read",
			OriginalErr: err,
		}
	}

	if err := yaml.Unmarshal(data, config); err != nil {
		commons.DefaultLogger.Error("Failed to parse config file: %v", err)
		return nil, &commons.ConfigError{
			Message:     "failed to parse config file",
			ConfigKey:   "",
			OriginalErr: err,
		}
	}

	// Fall back to environment variables if not set in config
	return loadFromEnv(config), nil
}

func loadFromEnv(config *Config) *Config {
	commons.DefaultLogger.Debug("Loading configuration from environment variables")
	if envHost := os.Getenv("PGHOST"); envHost != "" {
		config.Host = envHost
	}
	if envPort := os.Getenv("PGPORT"); envPort != "" {
		config.Port = parseInt(envPort, config.Port)
	}
	if envUser := os.Getenv("PGUSER"); envUser != "" {
		config.User = envUser
	}
	if envPassword := os.Getenv("PGPASSWORD"); envPassword != "" {
		config.Password = envPassword
	}
	if envDatabase := os.Getenv("PGDATABASE"); envDatabase != "" {
		config.Database = envDatabase
	}
	return config
}

func fileExists(path string) bool {
	commons.DefaultLogger.Debug("Checking if file exists: %s", path)
	_, err := os.Stat(path)
	return err == nil
}

func parseInt(s string, def int) int {
	if s == "" {
		return def
	}
	val, err := strconv.Atoi(s)
	if err != nil {
		commons.DefaultLogger.Warn("Failed to parse config value: %v, using default: %d", err, def)
		return def
	}
	return val
}

// SaveConfig saves configuration to a YAML file
func SaveConfig(config *Config, filePath string) error {
	commons.DefaultLogger.Debug("Saving configuration to file: %s", filePath)

	data, err := yaml.Marshal(config)
	if err != nil {
		commons.DefaultLogger.Error("Failed to marshal config: %v", err)
		return &commons.ConfigError{
			Message:     "failed to marshal config",
			ConfigKey:   "",
			OriginalErr: err,
		}
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		commons.DefaultLogger.Error("Failed to write config file: %v", err)
		return &commons.FileError{
			Message:     "failed to write config file",
			Path:        filePath,
			Action:      "write",
			OriginalErr: err,
		}
	}
	return nil
}

// ToString returns connection string
func (c *Config) ToString() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Database, c.SSLMode)
}
