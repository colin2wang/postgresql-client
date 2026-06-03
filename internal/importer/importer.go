package importer

import (
	"bufio"
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/colin2wang/postgresql-client/internal/cli"
	"github.com/colin2wang/postgresql-client/internal/commons"
	"github.com/colin2wang/postgresql-client/internal/config"
	"github.com/colin2wang/postgresql-client/internal/database"
)

// Importer handles data import operations
type Importer struct {
	DB  *database.Database
	Q   *database.Query
	Cfg *config.ImportConfig
}

// NewImporter creates a new Importer
func NewImporter(db *database.Database, q *database.Query, cfg *config.ImportConfig) *Importer {
	if cfg == nil {
		defaultCfg := config.DefaultImportConfig()
		cfg = &defaultCfg
	}
	return &Importer{
		DB:  db,
		Q:   q,
		Cfg: cfg,
	}
}

// extractTableName extracts the table name from a CREATE TABLE statement
func extractTableName(ddl string) string {
	re := regexp.MustCompile(`(?i)CREATE\s+TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?(?:public\.)?["\` + "`" + `]?(\w+)["\` + "`" + `]?`)
	matches := re.FindStringSubmatch(ddl)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// ImportDDL reads a DDL file and executes CREATE TABLE statements
func (imp *Importer) ImportDDL(ctx context.Context, filePath string) error {
	commons.DefaultLogger.Info("Importing DDL from file: %s", filePath)

	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read DDL file %s: %w", filePath, err)
	}

	content := strings.TrimSpace(string(data))
	if content == "" {
		return fmt.Errorf("DDL file is empty: %s", filePath)
	}

	// Validate SQL syntax before executing
	if err := validateSQL(content); err != nil {
		return fmt.Errorf("DDL validation failed: %w", err)
	}

	// Extract all CREATE TABLE statements
	statements := splitSQLStatements(content)
	if len(statements) == 0 {
		return fmt.Errorf("no valid SQL statements found in DDL file: %s", filePath)
	}

	for _, stmt := range statements {
		tableName := extractTableName(stmt)
		if tableName == "" {
			continue
		}

		// Check if table already exists
		exists, err := tableExists(ctx, imp.Q, tableName)
		if err != nil {
			return fmt.Errorf("failed to check table existence for %s: %w", tableName, err)
		}

		if exists {
			confirm := cli.NewConfirm(
				fmt.Sprintf("Table \"%s\" already exists. Import will be skipped. Continue?", tableName),
				false,
			)
			confirmed, err := confirm.Ask()
			if err != nil {
				return fmt.Errorf("confirmation failed: %w", err)
			}
			if !confirmed {
				fmt.Printf("Skipped DDL for table \"%s\".\n", tableName)
				continue
			}
			fmt.Printf("Skipping DDL for existing table \"%s\".\n", tableName)
			continue
		}

		// Execute the DDL
		_, err = imp.DB.ExecuteNonQuery(ctx, stmt)
		if err != nil {
			return fmt.Errorf("failed to execute DDL for table \"%s\": %w", tableName, err)
		}
		fmt.Printf("Table \"%s\" created successfully.\n", tableName)
	}

	return nil
}

// ImportCSV reads a CSV file and imports data into a table
func (imp *Importer) ImportCSV(ctx context.Context, filePath string) error {
	commons.DefaultLogger.Info("Importing CSV from file: %s", filePath)

	// Open and read CSV file
	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open CSV file %s: %w", filePath, err)
	}
	defer f.Close()

	reader := csv.NewReader(bufio.NewReader(f))
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true

	allRows, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to parse CSV file %s: %w", filePath, err)
	}

	if len(allRows) < 1 {
		return fmt.Errorf("CSV file is empty: %s", filePath)
	}

	headers := allRows[0]
	dataRows := allRows[1:]

	if len(dataRows) == 0 {
		return fmt.Errorf("CSV file has no data rows (header only): %s", filePath)
	}

	// Validate CSV data
	if err := validateCSV(headers, dataRows); err != nil {
		return fmt.Errorf("CSV validation failed: %w", err)
	}

	// Extract table name from filename (without extension)
	baseName := filepath.Base(filePath)
	tableName := strings.TrimSuffix(baseName, filepath.Ext(baseName))

	// Check if table exists
	exists, err := tableExists(ctx, imp.Q, tableName)
	if err != nil {
		return fmt.Errorf("failed to check table existence: %w", err)
	}

	compareImport := false

	if !exists {
		// Table doesn't exist - ask user if they want to create it
		confirm := cli.NewConfirm(
			fmt.Sprintf("Table \"%s\" does not exist. Create table with columns %v?", tableName, headers),
			false,
		)
		confirmed, err := confirm.Ask()
		if err != nil {
			return fmt.Errorf("confirmation failed: %w", err)
		}
		if !confirmed {
			fmt.Printf("Import cancelled for table \"%s\".\n", tableName)
			return nil
		}

		// Auto-create table with text columns based on CSV headers
		createStmt := buildCreateTableFromCSV(tableName, headers)
		_, err = imp.DB.ExecuteNonQuery(ctx, createStmt)
		if err != nil {
			return fmt.Errorf("failed to create table \"%s\": %w", tableName, err)
		}
		fmt.Printf("Table \"%s\" created automatically.\n", tableName)
	} else {
		// Table exists - check if it's empty
		rowCount, err := imp.Q.GetRowCountForTable(ctx, tableName)
		if err != nil {
			return fmt.Errorf("failed to get row count for table \"%s\": %w", tableName, err)
		}

		if rowCount > 0 {
			// Ask user whether to use compare import (skip identical rows)
			compareConfirm := cli.NewConfirm(
				fmt.Sprintf("Table \"%s\" already has %d row(s). Compare import (skip identical rows)?", tableName, rowCount),
				true, // default yes
			)
			compareImport, err = compareConfirm.Ask()
			if err != nil {
				return fmt.Errorf("confirmation failed: %w", err)
			}
		}
		// Empty table: import directly (compareImport stays false)
	}

	// Import data
	inserted, skipped, failed := 0, 0, 0

	if compareImport {
		// Per-row with duplicate check
		for _, row := range dataRows {
			exists, err := rowExists(ctx, imp.DB, tableName, headers, row)
			if err != nil {
				commons.DefaultLogger.Warn("Failed to check row existence, skipping: %v", err)
				failed++
				continue
			}
			if exists {
				skipped++
				continue
			}
			if err := imp.insertRow(ctx, tableName, headers, row); err != nil {
				commons.DefaultLogger.Warn("Failed to insert row, skipping: %v", err)
				failed++
				continue
			}
			inserted++
		}
	} else {
		// Batch insert with per-row fallback on error
		inserted = imp.batchInsertRows(ctx, tableName, headers, dataRows)
	}

	// Print summary
	if compareImport && skipped > 0 {
		fmt.Printf("Skipped %d duplicate row(s).\n", skipped)
	}
	fmt.Printf("Successfully imported %d row(s) into table \"%s\".\n", inserted, tableName)
	if failed > 0 {
		fmt.Printf("Warning: %d row(s) failed to import.\n", failed)
	}
	return nil
}

// ExecuteSQLFile reads and executes an SQL script file
func (imp *Importer) ExecuteSQLFile(ctx context.Context, filePath string) error {
	commons.DefaultLogger.Info("Executing SQL file: %s", filePath)

	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read SQL file %s: %w", filePath, err)
	}

	content := strings.TrimSpace(string(data))
	if content == "" {
		return fmt.Errorf("SQL file is empty: %s", filePath)
	}

	// Validate SQL syntax
	if err := validateSQL(content); err != nil {
		return fmt.Errorf("SQL validation failed: %w", err)
	}

	statements := splitSQLStatements(content)
	if len(statements) == 0 {
		return fmt.Errorf("no valid SQL statements found in file: %s", filePath)
	}

	executedCount := 0
	for _, stmt := range statements {
		_, err := imp.DB.ExecuteNonQuery(ctx, stmt)
		if err != nil {
			return fmt.Errorf("failed to execute statement [%s...]: %w", truncateString(stmt, 60), err)
		}
		executedCount++
	}

	fmt.Printf("Successfully executed %d SQL statement(s) from \"%s\".\n", executedCount, filePath)
	return nil
}

// tableExists checks if a table exists in the public schema
func tableExists(ctx context.Context, q *database.Query, tableName string) (bool, error) {
	query := "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name = $1"
	rows, err := q.DB.ExecuteQuery(ctx, query, tableName)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	if rows.Next() {
		var count int
		if err := rows.Scan(&count); err != nil {
			return false, err
		}
		return count > 0, nil
	}
	return false, nil
}

// buildCreateTableFromCSV generates a CREATE TABLE statement from CSV headers
func buildCreateTableFromCSV(tableName string, headers []string) string {
	var columns []string
	for _, h := range headers {
		col := fmt.Sprintf("%s TEXT", quoteIdentifier(h))
		columns = append(columns, col)
	}
	return fmt.Sprintf("CREATE TABLE %s (%s);", quoteIdentifier(tableName), strings.Join(columns, ", "))
}

// batchInsertRows inserts data rows in batches, falling back to per-row on error
func (imp *Importer) batchInsertRows(ctx context.Context, tableName string, headers []string, dataRows [][]string) int {
	const batchSize = 500
	inserted := 0

	quotedHeaders := make([]string, len(headers))
	for j, h := range headers {
		quotedHeaders[j] = quoteIdentifier(h)
	}
	quotedTable := quoteIdentifier(tableName)

	for i := 0; i < len(dataRows); i += batchSize {
		end := i + batchSize
		if end > len(dataRows) {
			end = len(dataRows)
		}
		batch := dataRows[i:end]

		// Try batch insert first
		inserted += imp.tryBatchInsert(ctx, quotedTable, quotedHeaders, batch)

		// Note: batch failures are handled inside tryBatchInsert by falling back per-row
	}
	return inserted
}

// tryBatchInsert attempts a batch INSERT; on failure falls back to per-row
func (imp *Importer) tryBatchInsert(ctx context.Context, quotedTable string, quotedHeaders []string, batch [][]string) int {
	inserted := 0

	// Build multi-row INSERT statement
	placeholders := make([]string, 0, len(batch))
	args := make([]interface{}, 0, len(quotedHeaders)*len(batch))
	argIdx := 1
	for _, row := range batch {
		rowPlaceholders := make([]string, len(quotedHeaders))
		for j, val := range row {
			rowPlaceholders[j] = fmt.Sprintf("$%d", argIdx)
			args = append(args, val)
			argIdx++
		}
		placeholders = append(placeholders, "("+strings.Join(rowPlaceholders, ",")+")")
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s",
		quotedTable,
		strings.Join(quotedHeaders, ","),
		strings.Join(placeholders, ","))

	_, err := imp.DB.ExecuteNonQuery(ctx, query, args...)
	if err != nil {
		// Batch failed, fall back to per-row insert
		commons.DefaultLogger.Warn("Batch insert failed, falling back to per-row: %v", err)
		for _, row := range batch {
			if err := imp.insertRow(ctx, quotedTable, quotedHeaders, row); err != nil {
				commons.DefaultLogger.Warn("Failed to insert row, skipping: %v", err)
				continue
			}
			inserted++
		}
		return inserted
	}
	return len(batch)
}

// insertRow inserts a single row into the table
func (imp *Importer) insertRow(ctx context.Context, quotedTable string, quotedHeaders []string, row []string) error {
	placeholders := make([]string, len(quotedHeaders))
	args := make([]interface{}, len(row))
	for i, val := range row {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = val
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		quotedTable,
		strings.Join(quotedHeaders, ","),
		strings.Join(placeholders, ","))

	_, err := imp.DB.ExecuteNonQuery(ctx, query, args...)
	return err
}

// rowExists checks if a row with identical values exists in the table
func rowExists(ctx context.Context, db *database.Database, tableName string, headers []string, row []string) (bool, error) {
	conditions := make([]string, len(headers))
	args := make([]interface{}, len(row))
	for i, val := range row {
		conditions[i] = fmt.Sprintf("%s = $%d", quoteIdentifier(headers[i]), i+1)
		args[i] = val
	}
	query := fmt.Sprintf("SELECT 1 FROM %s WHERE %s LIMIT 1",
		quoteIdentifier(tableName),
		strings.Join(conditions, " AND "))

	rows, err := db.ExecuteQuery(ctx, query, args...)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	return rows.Next(), nil
}

// quoteIdentifier quotes a PostgreSQL identifier
func quoteIdentifier(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}

// truncateString truncates a string to maxLen
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
