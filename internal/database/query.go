package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"

	"github.com/colin2wang/postgresql-client/internal/commons"
	"github.com/colin2wang/postgresql-client/internal/config"
)

// DatabaseInfo represents database information including name and table count
type DatabaseInfo struct {
	Name       string
	TableCount int
}

// TableInfo represents table information including name and row count
type TableInfo struct {
	Name     string
	RowCount int
}

// ColumnInfo represents a column in a table
type ColumnInfo struct {
	Name         string
	DataType     string
	IsNullable   bool
	DefaultValue string
}

// Query provides database query helper methods
type Query struct {
	DB *Database
}

// NewQuery creates a new query helper
func NewQuery(db *Database) *Query {
	return &Query{DB: db}
}

// GetAllDatabases retrieves all available databases and their table counts
func (q *Query) GetAllDatabases(ctx context.Context) ([]DatabaseInfo, error) {
	query := `SELECT datname FROM pg_database WHERE datistemplate = false AND datallowconn = true ORDER BY datname`
	rows, err := q.DB.ExecuteQuery(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query databases: %w", err)
	}
	defer rows.Close()

	var databases []DatabaseInfo
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		tableCount, err := q.GetTableCountForDatabase(ctx)
		if err != nil {
			commons.DefaultLogger.Warn("Failed to get table count for database %s: %v", name, err)
			tableCount = 0
		}
		databases = append(databases, DatabaseInfo{Name: name, TableCount: tableCount})
	}
	return databases, nil
}

// GetTableCountForDatabase retrieves the number of tables in the current database
func (q *Query) GetTableCountForDatabase(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_type = 'BASE TABLE'`
	rows, err := q.DB.ExecuteQuery(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("failed to query table count: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		var count int
		if err := rows.Scan(&count); err != nil {
			return 0, fmt.Errorf("scan failed: %w", err)
		}
		return count, nil
	}
	return 0, nil
}

// GetAllTables retrieves all available tables and their row counts
func (q *Query) GetAllTables(ctx context.Context) ([]TableInfo, error) {
	query := `SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' AND table_type = 'BASE TABLE' ORDER BY table_name`
	rows, err := q.DB.ExecuteQuery(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query tables: %w", err)
	}
	defer rows.Close()

	var tables []TableInfo
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		rowCount, err := q.GetRowCountForTable(ctx, name)
		if err != nil {
			commons.DefaultLogger.Warn("Failed to get row count for table %s: %v", name, err)
			rowCount = 0
		}
		tables = append(tables, TableInfo{Name: name, RowCount: rowCount})
	}
	return tables, nil
}

// GetRowCountForTable retrieves the number of rows in a table
func (q *Query) GetRowCountForTable(ctx context.Context, tableName string) (int, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)
	rows, err := q.DB.ExecuteQuery(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("failed to query row count: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		var count int
		if err := rows.Scan(&count); err != nil {
			return 0, fmt.Errorf("scan failed: %w", err)
		}
		return count, nil
	}
	return 0, nil
}

// DescribeTable retrieves column information for a table
func (q *Query) DescribeTable(ctx context.Context, tableName string) ([]ColumnInfo, error) {
	query := `SELECT column_name, data_type, is_nullable, column_default FROM information_schema.columns WHERE table_schema = 'public' AND table_name = $1 ORDER BY ordinal_position`
	rows, err := q.DB.ExecuteQuery(ctx, query, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to describe table: %w", err)
	}
	defer rows.Close()

	var columns []ColumnInfo
	for rows.Next() {
		var name, typ string
		var isNullable, columnDefault sql.NullString
		if err := rows.Scan(&name, &typ, &isNullable, &columnDefault); err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		col := ColumnInfo{
			Name:     name,
			DataType: typ,
		}
		if isNullable.Valid && isNullable.String == "YES" {
			col.IsNullable = true
		}
		if columnDefault.Valid {
			col.DefaultValue = columnDefault.String
		}
		columns = append(columns, col)
	}
	return columns, nil
}

// GetRowByValues retrieves a row by its values (all columns as WHERE)
func (q *Query) GetRowByValues(ctx context.Context, tableName string, columns []string, row map[string]interface{}) (map[string]interface{}, error) {
	whereClauses := make([]string, 0, len(columns))
	args := make([]interface{}, 0, len(columns))
	for _, col := range columns {
		whereClauses = append(whereClauses, fmt.Sprintf("%s = $%d", col, len(args)+1))
		args = append(args, row[col])
	}
	query := fmt.Sprintf("SELECT * FROM %s WHERE %s", tableName, strings.Join(whereClauses, " AND "))
	rows, err := q.DB.ExecuteQuery(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query row: %w", err)
	}
	defer rows.Close()

	result, err := q.ParseTableData(rows, columns)
	if err != nil {
		return nil, fmt.Errorf("failed to parse row data: %w", err)
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("row not found")
	}
	return result[0], nil
}

// ParseTableData converts SQL rows into a slice of map[string]interface{}
func (q *Query) ParseTableData(rows *sql.Rows, columns []string) ([]map[string]interface{}, error) {
	var tableData []map[string]interface{}
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}
	for rows.Next() {
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		rowMap := make(map[string]interface{})
		for i, col := range columns {
			rowMap[col] = values[i]
		}
		tableData = append(tableData, rowMap)
	}
	return tableData, nil
}

// DeleteRow deletes a row from the table
func (q *Query) DeleteRow(ctx context.Context, tableName string, columns []string, row map[string]interface{}) error {
	whereClauses := make([]string, 0, len(columns))
	args := make([]interface{}, 0, len(columns))
	for _, col := range columns {
		whereClauses = append(whereClauses, fmt.Sprintf("%s = $%d", col, len(args)+1))
		args = append(args, row[col])
	}
	query := fmt.Sprintf("DELETE FROM %s WHERE %s", tableName, strings.Join(whereClauses, " AND "))
	_, err := q.DB.ExecuteNonQuery(ctx, query, args...)
	return err
}

// UpdateTableRow updates a row with new values, matching all columns
func (q *Query) UpdateTableRow(ctx context.Context, tableName string, columns []string, oldRow, newRow map[string]interface{}) error {
	setClauses := make([]string, 0, len(columns))
	whereClauses := make([]string, 0, len(columns))
	args := make([]interface{}, 0, len(columns)*2)
	for _, col := range columns {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", col, len(args)+1))
		args = append(args, newRow[col])
	}
	for _, col := range columns {
		whereClauses = append(whereClauses, fmt.Sprintf("%s = $%d", col, len(args)+1))
		args = append(args, oldRow[col])
	}
	query := fmt.Sprintf("UPDATE %s SET %s WHERE %s", tableName, strings.Join(setClauses, ", "), strings.Join(whereClauses, " AND "))
	_, err := q.DB.ExecuteNonQuery(ctx, query, args...)
	return err
}

// ExecuteScript executes SQL commands from a file
func (q *Query) ExecuteScript(ctx context.Context, filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return &commons.FileError{
			Message:     "failed to read SQL file",
			Path:        filename,
			Action:      "read",
			OriginalErr: err,
		}
	}
	query := strings.TrimSpace(string(data))
	if query == "" {
		return fmt.Errorf("empty SQL file")
	}
	result, err := q.DB.ExecuteQuery(ctx, query)
	if err != nil {
		return err
	}
	result.Close()
	return nil
}

// NewDatabase creates a new Database and Query helper from config
func NewDatabaseAndQuery(cfg *config.Config) (*Database, *Query, error) {
	db, err := NewDatabase(cfg)
	if err != nil {
		return nil, nil, err
	}
	return db, NewQuery(db), nil
}
