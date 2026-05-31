package config

import (
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

const DefaultPort = 5432

// Config stores database connection configuration
type Config struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
	SSLMode  string `yaml:"ssl_mode,omitempty"`
}

// NewConfig creates a new configuration
func NewConfig() *Config {
	return &Config{
		Host:     "localhost",
		Port:     DefaultPort,
		User:     "postgres",
		Database: "postgres",
		SSLMode:  "disable",
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
	}
}

// LoadConfig loads configuration from a YAML file
func LoadConfig(filePath string) (*Config, error) {
	config := DefaultConfig()

	// If no file path provided or file doesn't exist, use environment variables
	if filePath == "" || !fileExists(filePath) {
		return loadFromEnv(config), nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Fall back to environment variables if not set in config
	return loadFromEnv(config), nil
}

func loadFromEnv(config *Config) *Config {
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
	_, err := os.Stat(path)
	return err == nil
}

func parseInt(s string, def int) int {
	if s == "" {
		return def
	}
	val, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return val
}

// SaveConfig saves configuration to a YAML file
func SaveConfig(config *Config, filePath string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	return nil
}

// ToString returns connection string
func (c *Config) ToString() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Database, c.SSLMode)
}
