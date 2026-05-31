package database

import (
	"context"
	"database/sql"

	"github.com/colin2wang/postgresql-client/commons"
	"github.com/colin2wang/postgresql-client/config"
)

// Database handles all database operations
type Database struct {
	connector *Connector
	config    *config.Config
}

// ExecuteQuery executes a query and returns rows
func (db *Database) ExecuteQuery(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	if db.connector == nil {
		commons.DefaultLogger.Error("Database connector is not initialized")
		return nil, &commons.DatabaseError{
			Message: "database connector is not initialized",
		}
	}
	commons.DefaultLogger.Debug("Executing query via database: %s", query[:min(len(query), 50)])
	return db.connector.ExecuteQuery(ctx, query, args...)
}

// ExecuteNonQuery executes a non-query statement
func (db *Database) ExecuteNonQuery(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	if db.connector == nil {
		commons.DefaultLogger.Error("Database connector is not initialized")
		return nil, &commons.DatabaseError{
			Message: "database connector is not initialized",
		}
	}
	commons.DefaultLogger.Debug("Executing non-query via database: %s", query[:min(len(query), 50)])
	return db.connector.ExecuteNonQuery(ctx, query, args...)
}

// Close closes the database connection
func (db *Database) Close() error {
	if db.connector != nil {
		commons.DefaultLogger.Debug("Closing database")
		return db.connector.Close()
	}
	return nil
}

// Connector wraps sql.DB for database operations
type Connector struct {
	db     *sql.DB
	config *config.Config
}

// NewDatabase creates a new database handler
func NewDatabase(cfg *config.Config) (*Database, error) {
	commons.DefaultLogger.Info("Initializing database connection")
	conn, err := NewConnector(cfg)
	if err != nil {
		return nil, err
	}
	return &Database{
		connector: conn,
		config:    cfg,
	}, nil
}

// NewConnector creates a new connector
func NewConnector(cfg *config.Config) (*Connector, error) {
	connStr := cfg.ToString()

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		commons.DefaultLogger.Error("Failed to connect to database: %v", err)
		return nil, &commons.DatabaseError{
			Message:     "failed to connect to database",
			OriginalErr: err,
		}
	}

	if err := db.PingContext(context.Background()); err != nil {
		commons.DefaultLogger.Error("Database ping failed: %v", err)
		return nil, &commons.DatabaseError{
			Message:     "database ping failed",
			OriginalErr: err,
		}
	}

	return &Connector{
		db:     db,
		config: cfg,
	}, nil
}

// ExecuteQuery executes a query and returns rows
func (c *Connector) ExecuteQuery(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	commons.DefaultLogger.Debug("Executing query via connector: %s", query[:min(len(query), 50)])
	return c.db.QueryContext(ctx, query, args...)
}

// ExecuteNonQuery executes a non-query statement
func (c *Connector) ExecuteNonQuery(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	commons.DefaultLogger.Debug("Executing non-query via connector: %s", query[:min(len(query), 50)])
	return c.db.ExecContext(ctx, query, args...)
}

// Close closes the database connection
func (c *Connector) Close() error {
	if c.db != nil {
		commons.DefaultLogger.Debug("Closing connector")
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
