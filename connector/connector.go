package connector

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/colin2wang/postgresql-client/commons"
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
	commons.DefaultLogger.Debug("Creating new connector for host=%s port=%d", config.Host, config.Port)
	connStr := buildConnString(config)

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		commons.DefaultLogger.Error("Failed to connect to database: %v", err)
		return nil, &commons.ConnectionError{
			Message:     "failed to connect to database",
			Host:        config.Host,
			Port:        config.Port,
			Database:    config.Database,
			OriginalErr: err,
		}
	}

	if err := db.PingContext(context.Background()); err != nil {
		commons.DefaultLogger.Error("Database ping failed: %v", err)
		return nil, &commons.ConnectionError{
			Message:     "database ping failed",
			Host:        config.Host,
			Port:        config.Port,
			Database:    config.Database,
			OriginalErr: err,
		}
	}

	return &Connector{
		config: config,
		db:     db,
	}, nil
}

// buildConnString builds the connection string from config
func buildConnString(config *Config) string {
	commons.DefaultLogger.Debug("Building connection string for host=%s port=%d", config.Host, config.Port)
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.Database, config.SSLMode)
}

// Execute executes a query and returns rows
func (c *Connector) Execute(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	commons.DefaultLogger.Debug("Executing query: %s", query[:min(len(query), 50)])
	return c.db.QueryContext(ctx, query, args...)
}

// Exec executes a non-query statement
func (c *Connector) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	commons.DefaultLogger.Debug("Executing non-query: %s", query[:min(len(query), 50)])
	return c.db.ExecContext(ctx, query, args...)
}

// Close closes the database connection
func (c *Connector) Close() error {
	commons.DefaultLogger.Debug("Closing database connection")
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
