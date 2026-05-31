package database

import (
	"context"
	"database/sql"

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
		return nil, &DatabaseError{Message: "database connector is not initialized"}
	}
	return db.connector.ExecuteQuery(ctx, query, args...)
}

// ExecuteNonQuery executes a non-query statement
func (db *Database) ExecuteNonQuery(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	if db.connector == nil {
		return nil, &DatabaseError{Message: "database connector is not initialized"}
	}
	return db.connector.ExecuteNonQuery(ctx, query, args...)
}

// Close closes the database connection
func (db *Database) Close() error {
	if db.connector != nil {
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
		return nil, &DatabaseError{Message: "failed to connect to database: " + err.Error()}
	}

	if err := db.PingContext(context.Background()); err != nil {
		return nil, &DatabaseError{Message: "database ping failed: " + err.Error()}
	}

	return &Connector{
		db:     db,
		config: cfg,
	}, nil
}

// ExecuteQuery executes a query and returns rows
func (c *Connector) ExecuteQuery(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return c.db.QueryContext(ctx, query, args...)
}

// ExecuteNonQuery executes a non-query statement
func (c *Connector) ExecuteNonQuery(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return c.db.ExecContext(ctx, query, args...)
}

// Close closes the database connection
func (c *Connector) Close() error {
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}

// DatabaseError represents a database-related error
type DatabaseError struct {
	Message string
}

func (e *DatabaseError) Error() string {
	return e.Message
}
