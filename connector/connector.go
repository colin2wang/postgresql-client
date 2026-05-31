package connector

import (
	"context"
	"database/sql"
	"fmt"

	// PostgreSQL driver
	_ "github.com/jackc/pgx/v5/stdlib"
)

// Connector handles database connections
type Connector struct {
	config *Config
	db     *sql.DB
}

// Config represents database connection configuration
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
	SSLMode  string
}

// NewConnector creates a new database connector
func NewConnector(config *Config) (*Connector, error) {
	connStr := buildConnString(config)

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return nil, &ConnectionError{Message: "failed to connect to database: " + err.Error()}
	}

	if err := db.PingContext(context.Background()); err != nil {
		return nil, &ConnectionError{Message: "database ping failed: " + err.Error()}
	}

	return &Connector{
		config: config,
		db:     db,
	}, nil
}

// buildConnString builds the connection string from config
func buildConnString(config *Config) string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.Database, config.SSLMode)
}

// Execute executes a query and returns rows
func (c *Connector) Execute(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return c.db.QueryContext(ctx, query, args...)
}

// Exec executes a non-query statement
func (c *Connector) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return c.db.ExecContext(ctx, query, args...)
}

// Close closes the database connection
func (c *Connector) Close() error {
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}

// ConnectionError represents a connection-related error
type ConnectionError struct {
	Message string
}

func (e *ConnectionError) Error() string {
	return e.Message
}
