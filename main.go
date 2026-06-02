package main

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/colin2wang/postgresql-client/internal/cli"
	"github.com/colin2wang/postgresql-client/internal/commons"
	"github.com/colin2wang/postgresql-client/internal/config"
	"github.com/colin2wang/postgresql-client/internal/database"
)

func main() {
	ctx := context.Background()

	// Parse command line flags
	configPath := flag.String("c", "", "Path to config file")
	host := flag.String("h", "", "Database host")
	port := flag.Int("p", 0, "Database port")
	user := flag.String("U", "", "Database user")
	password := flag.String("W", "", "Database password")
	databaseName := flag.String("d", "", "Database name")
	flag.Parse()

	// Initialize log file in the same directory as the executable
	if logDir, err := os.Getwd(); err == nil {
		logPath := logDir + "\\postgresql-client.log"
		if err := commons.DefaultLogger.SetLogFile(logPath); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to open log file: %v\n", err)
		} else {
			commons.DefaultLogger.Info("Log file: %s", logPath)
		}
	}

	// Load configuration
	cfg, err := loadConfig(*configPath, *host, *port, *user, *password, *databaseName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		commons.DefaultLogger.Error("Failed to load configuration: %v", err)
		os.Exit(1)
	}

	// Create database connection
	db, err := database.NewDatabase(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect: %v\n", err)
		commons.DefaultLogger.Error("Failed to connect to database: %v", err)
		os.Exit(1)
	}
	defer db.Close()

	commons.DefaultLogger.Info("Configuration loaded successfully")

	// Check if running in non-interactive mode
	if flag.NArg() > 0 {
		handleNonInteractive(ctx, db, flag.Args())
		return
	}

	// Run interactive mode
	runInteractive(ctx, db, cfg)
}

func loadConfig(configPath string, host string, port int, user, password, databaseName string) (*config.Config, error) {
	commons.DefaultLogger.Info("Loading configuration...")
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		commons.DefaultLogger.Error("Failed to load configuration: %v", err)
		return nil, &commons.ConfigError{
			Message:     "failed to load configuration",
			OriginalErr: err,
		}
	}

	// Override with command line arguments if provided
	if host != "" {
		cfg.Host = host
	}
	if port > 0 {
		cfg.Port = port
	}
	if user != "" {
		cfg.User = user
	}
	if password != "" {
		cfg.Password = password
	}
	if databaseName != "" {
		cfg.Database = databaseName
	}

	return cfg, nil
}

func handleNonInteractive(ctx context.Context, db *database.Database, args []string) {
	query := strings.Join(args, " ")

	switch args[0] {
	case "-f", "--file":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Error: -f requires a filename")
			os.Exit(1)
		}
		if err := runScript(ctx, db, args[1]); err != nil {
			fmt.Fprintf(os.Stderr, "Script execution failed: %v\n", err)
			os.Exit(1)
		}

	case "-c", "--command":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Error: -c requires a SQL command")
			os.Exit(1)
		}
		executeAndPrint(ctx, db, query)

	case "--json":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Error: --json requires a SQL query")
			os.Exit(1)
		}
		exportJSON(ctx, db, query)

	case "--csv":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Error: --csv requires a SQL query")
			os.Exit(1)
		}
		exportCSV(ctx, db, query)

	default:
		executeAndPrint(ctx, db, query)
	}
}

func runInteractive(ctx context.Context, db *database.Database, cfg *config.Config) {
	commons.DefaultLogger.Info("Starting interactive mode")

	printWelcome(cfg)

	history := commons.NewHistory(100)
	reader := bufio.NewReader(os.Stdin)
	prompt := fmt.Sprintf("%s@%s> ", cfg.User, cfg.Database)

	for {
		fmt.Print(prompt)

		text, err := reader.ReadString('\n')
		if err != nil {
			commons.DefaultLogger.Error("Error reading input: %v", err)
			fmt.Printf("\nError reading input: %v\n", err)
			continue
		}

		query := strings.TrimSpace(strings.TrimSuffix(text, "\n"))

		switch query {
		case "\\m", "\\menu":
			newDb, err := showMainMenu(ctx, db, cfg)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			}
			if newDb != nil && newDb != db {
				db = newDb
				prompt = fmt.Sprintf("%s@%s> ", cfg.User, cfg.Database)
			}

		case "\\q", "quit", "exit":
			fmt.Println("Goodbye!")
			return

		case "\\h", "\\?", "help":
			printWelcome(cfg)

		case "\\l", "\\list":
			if err := showDatabases(ctx, db); err != nil {
				fmt.Printf("Error: %v\n", err)
			}

		case "\\dt", "\\d tables":
			if err := listTables(ctx, db); err != nil {
				fmt.Printf("Error: %v\n", err)
			}

		case "\\t", "\\select-table":
			if err := selectTableInteractive(ctx, db, cfg); err != nil {
				fmt.Printf("Error: %v\n", err)
			}

		case "\\s", "\\select-db":
			newDb, err := selectDatabaseInteractive(ctx, db, cfg)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else if newDb != nil {
				db = newDb
				prompt = fmt.Sprintf("%s@%s> ", cfg.User, cfg.Database)
			}

		case "\\c", "\\C":
			newDb, err := selectDatabaseInteractive(ctx, db, cfg)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else if newDb != nil {
				db = newDb
				prompt = fmt.Sprintf("%s@%s> ", cfg.User, cfg.Database)
			}

		case "\\d", "\\D":
			if err := selectAndDescribeTable(ctx, db, cfg); err != nil {
				fmt.Printf("Error: %v\n", err)
			}

		default:
			if query != "" {
				history.Add(query)
				executeAndPrint(ctx, db, query)
			}
		}
	}
}

func printWelcome(cfg *config.Config) {
	fmt.Println("========================================")
	fmt.Println("  PostgreSQL CLI Client")
	fmt.Printf("  Connected to: %s:%d/%s\n", cfg.Host, cfg.Port, cfg.Database)
	fmt.Println("========================================")
	fmt.Println("Supported commands:")
	fmt.Println("  - SQL statements (SELECT, INSERT, UPDATE, DELETE, etc.)")
	fmt.Println("  - \\m              - Show main menu (interactive)")
	fmt.Println("  - \\q              - Quit the client")
	fmt.Println("  - \\h              - Show this help message")
	fmt.Println("  - \\l              - List all databases")
	fmt.Println("  - \\dt             - List all tables")
	fmt.Println("  - \\d              - Select and describe table structure (interactive)")
	fmt.Println("  - \\c              - Select and connect to database (interactive)")
	fmt.Println("  - \\s              - Select database with arrow keys")
	fmt.Println("  - \\t              - Select table with arrow keys and choose action")
	fmt.Println("                    (show structure, show content, or edit content)")
}

func executeAndPrint(ctx context.Context, db *database.Database, query string) {
	commons.DefaultLogger.Debug("Executing query: %s", query[:min(len(query), 50)])

	result, err := db.ExecuteQuery(ctx, query)
	if err != nil {
		commons.DefaultLogger.Error("Query execution failed: %v", err)
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer result.Close()

	// TODO: Implement table rendering with pretty table library
	// This would require integrating the go-pretty/table library
	fmt.Println("Query executed successfully")
}

func listTables(ctx context.Context, db *database.Database) error {
	commons.DefaultLogger.Debug("Listing tables...")
	query := `
		SELECT table_name, table_type 
		FROM information_schema.tables 
		WHERE table_schema = 'public' 
		ORDER BY table_name
	`

	rows, err := db.ExecuteQuery(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to list tables: %w", err)
	}
	defer rows.Close()

	fmt.Println("Found table(s) in database:")
	fmt.Printf("%-30s %-10s %s\n", "Table Name", "Type", "Row Count")
	fmt.Println(strings.Repeat("-", 55))

	rowCount := 0
	for rows.Next() {
		var name, typ string
		if err := rows.Scan(&name, &typ); err != nil {
			return fmt.Errorf("scan failed: %w", err)
		}

		count, err := getRowCountForTable(ctx, db, name)
		if err != nil {
			fmt.Printf("%-30s %-10s %s\n", name, typ, "N/A")
		} else {
			fmt.Printf("%-30s %-10s %d\n", name, typ, count)
		}
		rowCount++
	}

	fmt.Printf("%d table(s) found\n", rowCount)
	return nil
}

func describeTable(ctx context.Context, db *database.Database, tableName string) error {
	commons.DefaultLogger.Debug("Describing table: %s", tableName)
	query := `
		SELECT 
			column_name,
			data_type,
			is_nullable,
			column_default
		FROM information_schema.columns
		WHERE table_schema = 'public' AND table_name = $1
		ORDER BY ordinal_position
	`

	rows, err := db.ExecuteQuery(ctx, query, tableName)
	if err != nil {
		return fmt.Errorf("failed to describe table: %w", err)
	}
	defer rows.Close()

	fmt.Printf("Structure of table '%s':\n", tableName)
	fmt.Printf("%-20s %-15s %-8s %s\n", "Column Name", "Type", "Nullable", "Default")
	fmt.Println(strings.Repeat("-", 70))

	rowCount := 0
	for rows.Next() {
		var name, typ string
		var isNullable, columnDefault sql.NullString
		if err := rows.Scan(&name, &typ, &isNullable, &columnDefault); err != nil {
			return fmt.Errorf("scan failed: %w", err)
		}

		nullStr := "NO"
		if isNullable.Valid && isNullable.String == "YES" {
			nullStr = "YES"
		}
		defaultVal := columnDefault.String
		if defaultVal == "" {
			defaultVal = "<NULL>"
		}

		fmt.Printf("%-20s %-15s %-8s %s\n", name, typ, nullStr, defaultVal)
		rowCount++
	}

	fmt.Printf("%d column(s) found\n", rowCount)
	return nil
}

func showDatabases(ctx context.Context, db *database.Database) error {
	commons.DefaultLogger.Debug("Showing databases...")
	query := `SELECT datname FROM pg_database WHERE datistemplate = false ORDER BY datname`

	rows, err := db.ExecuteQuery(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to show databases: %w", err)
	}
	defer rows.Close()

	fmt.Println("Found database(s):")
	fmt.Printf("%-30s\n", "Database Name")
	fmt.Println(strings.Repeat("-", 32))

	rowCount := 0
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return fmt.Errorf("scan failed: %w", err)
		}
		fmt.Printf("%-30s\n", name)
		rowCount++
	}

	fmt.Printf("%d database(s) found\n", rowCount)
	return nil
}

// selectDatabaseInteractive interactively selects a database using arrow keys
func selectDatabaseInteractive(ctx context.Context, db *database.Database, cfg *config.Config) (*database.Database, error) {
	commons.DefaultLogger.Debug("Starting interactive database selection")

	databases, err := getAllDatabases(ctx, db)
	if err != nil {
		return nil, fmt.Errorf("failed to get databases: %w", err)
	}

	if len(databases) == 0 {
		fmt.Println("No databases found.")
		return db, nil
	}

	dbOptions := make([]string, len(databases))
	for i, dbInfo := range databases {
		dbOptions[i] = fmt.Sprintf("%s (%d tables)", dbInfo.Name, dbInfo.TableCount)
	}

	selector := cli.NewDBSelector()
	selected, err := selector.Select("Select a database:", dbOptions)
	if err != nil {
		return nil, fmt.Errorf("database selection failed: %w", err)
	}

	selectedDBName := strings.Split(selected, " (")[0]

	newCfg := *cfg
	newCfg.Database = selectedDBName
	db.Close()

	newDb, err := database.NewDatabase(&newCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database '%s': %w", selectedDBName, err)
	}

	cfg.Database = selectedDBName
	fmt.Printf("Successfully connected to database: %s\n", selectedDBName)
	return newDb, nil
}

// DatabaseInfo represents database information including name and table count
type DatabaseInfo struct {
	Name       string // Name of the database
	TableCount int    // Number of tables in the database
}

// getAllDatabases retrieves all available databases and their table counts
func getAllDatabases(ctx context.Context, db *database.Database) ([]DatabaseInfo, error) {
	query := `SELECT datname FROM pg_database WHERE datistemplate = false AND datallowconn = true ORDER BY datname`

	rows, err := db.ExecuteQuery(ctx, query)
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

		tableCount, err := getTableCountForDatabase(ctx, db, name)
		if err != nil {
			commons.DefaultLogger.Warn("Failed to get table count for database %s: %v", name, err)
			tableCount = 0
		}

		databases = append(databases, DatabaseInfo{
			Name:       name,
			TableCount: tableCount,
		})
	}

	return databases, nil
}

// getTableCountForDatabase retrieves the number of tables in a specified database
func getTableCountForDatabase(ctx context.Context, db *database.Database, dbName string) (int, error) {
	query := `
		SELECT COUNT(*) 
		FROM information_schema.tables 
		WHERE table_schema = 'public' AND table_type = 'BASE TABLE'
	`

	rows, err := db.ExecuteQuery(ctx, query)
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

// selectTableInteractive interactively selects a table using arrow keys for user selection
func selectTableInteractive(ctx context.Context, db *database.Database, cfg *config.Config) error {
	commons.DefaultLogger.Debug("Starting interactive table selection")

	tables, err := getAllTables(ctx, db)
	if err != nil {
		return fmt.Errorf("failed to get tables: %w", err)
	}

	if len(tables) == 0 {
		fmt.Println("No tables found.")
		return nil
	}

	tableOptions := make([]string, len(tables))
	for i, tableInfo := range tables {
		tableOptions[i] = fmt.Sprintf("%s (%d rows)", tableInfo.Name, tableInfo.RowCount)
	}

	selector := cli.NewTableSelector()
	selected, err := selector.Select(fmt.Sprintf("Select a table (current: %s):", cfg.Database), tableOptions)
	if err != nil {
		return fmt.Errorf("table selection failed: %w", err)
	}

	selectedTableName := strings.Split(selected, " (")[0]
	fmt.Printf("Selected table: %s\n", selectedTableName)

	actionSelector := cli.NewTableActionSelector()
	action, err := actionSelector.Select(fmt.Sprintf("What would you like to do with table '%s'?", selectedTableName))
	if err != nil {
		return fmt.Errorf("action selection failed: %w", err)
	}

	switch action {
	case "Show table structure":
		if err := describeTable(ctx, db, selectedTableName); err != nil {
			return fmt.Errorf("failed to describe table: %w", err)
		}
	case "Show table content":
		if err := showTableContent(ctx, db, selectedTableName); err != nil {
			return fmt.Errorf("failed to show table content: %w", err)
		}
	}

	return nil
}

// TableInfo represents table information including name and row count
type TableInfo struct {
	Name     string // Name of the table
	RowCount int    // Number of rows in the table
}

// getAllTables retrieves all available tables and their row counts
func getAllTables(ctx context.Context, db *database.Database) ([]TableInfo, error) {
	query := `
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = 'public' AND table_type = 'BASE TABLE'
		ORDER BY table_name
	`

	rows, err := db.ExecuteQuery(ctx, query)
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

		rowCount, err := getRowCountForTable(ctx, db, name)
		if err != nil {
			commons.DefaultLogger.Warn("Failed to get row count for table %s: %v", name, err)
			rowCount = 0
		}

		tables = append(tables, TableInfo{
			Name:     name,
			RowCount: rowCount,
		})
	}

	return tables, nil
}

// getRowCountForTable retrieves the number of rows in a specified table
func getRowCountForTable(ctx context.Context, db *database.Database, tableName string) (int, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)

	rows, err := db.ExecuteQuery(ctx, query)
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

// selectAndDescribeTable selects a table from interactive list and displays its structure
func selectAndDescribeTable(ctx context.Context, db *database.Database, cfg *config.Config) error {
	commons.DefaultLogger.Debug("Starting interactive table selection for description")

	tables, err := getAllTables(ctx, db)
	if err != nil {
		return fmt.Errorf("failed to get tables: %w", err)
	}

	if len(tables) == 0 {
		fmt.Println("No tables found.")
		return nil
	}

	tableOptions := make([]string, len(tables))
	for i, tableInfo := range tables {
		tableOptions[i] = fmt.Sprintf("%s (%d rows)", tableInfo.Name, tableInfo.RowCount)
	}

	selector := cli.NewTableSelector()
	selected, err := selector.Select(fmt.Sprintf("Select a table to describe (current: %s):", cfg.Database), tableOptions)
	if err != nil {
		return fmt.Errorf("table selection failed: %w", err)
	}

	selectedTableName := strings.Split(selected, " (")[0]
	return describeTable(ctx, db, selectedTableName)
}

// showMainMenu displays the main menu with available actions for user selection
func showMainMenu(ctx context.Context, db *database.Database, cfg *config.Config) (*database.Database, error) {
	commons.DefaultLogger.Debug("Showing main menu")

	menu := cli.NewMenu("Select an action:", []string{
		"List all databases",
		"List all tables",
		"Select and describe table structure",
		"Select and show table content",
		"Select and connect to database",
		"Execute custom SQL query",
		"Show help",
		"Quit",
	})

	selected, err := menu.Select()
	if err != nil {
		return nil, fmt.Errorf("menu selection failed: %w", err)
	}

	switch selected {
	case "List all databases":
		if err := showDatabases(ctx, db); err != nil {
			return nil, fmt.Errorf("failed to show databases: %w", err)
		}
	case "List all tables":
		if err := listTables(ctx, db); err != nil {
			return nil, fmt.Errorf("failed to list tables: %w", err)
		}
	case "Select and describe table structure":
		if err := selectAndDescribeTable(ctx, db, cfg); err != nil {
			return nil, fmt.Errorf("failed to select and describe table: %w", err)
		}
	case "Select and show table content":
		if err := selectAndShowTableContent(ctx, db, cfg); err != nil {
			return nil, fmt.Errorf("failed to select and show table content: %w", err)
		}
	case "Select and connect to database":
		newDb, err := selectDatabaseInteractive(ctx, db, cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to select database: %w", err)
		}
		if newDb != db {
			return newDb, nil
		}
		return db, nil
	case "Execute custom SQL query":
		if err := executeCustomQuery(ctx, db); err != nil {
			return nil, fmt.Errorf("failed to execute custom query: %w", err)
		}
	case "Show help":
		printWelcome(cfg)
	case "Quit":
		fmt.Println("Goodbye!")
		os.Exit(0)
	}

	return db, nil
}

// selectAndShowTableContent selects a table from interactive list and displays its content
func selectAndShowTableContent(ctx context.Context, db *database.Database, cfg *config.Config) error {
	commons.DefaultLogger.Debug("Starting interactive table selection for content display")

	tables, err := getAllTables(ctx, db)
	if err != nil {
		return fmt.Errorf("failed to get tables: %w", err)
	}

	if len(tables) == 0 {
		fmt.Println("No tables found.")
		return nil
	}

	tableOptions := make([]string, len(tables))
	for i, tableInfo := range tables {
		tableOptions[i] = fmt.Sprintf("%s (%d rows)", tableInfo.Name, tableInfo.RowCount)
	}

	selector := cli.NewTableSelector()
	selected, err := selector.Select(fmt.Sprintf("Select a table to show content (current: %s):", cfg.Database), tableOptions)
	if err != nil {
		return fmt.Errorf("table selection failed: %w", err)
	}

	selectedTableName := strings.Split(selected, " (")[0]
	return showTableContent(ctx, db, selectedTableName)
}

// executeCustomQuery prompts user for a custom SQL query and executes it
func executeCustomQuery(ctx context.Context, db *database.Database) error {
	input := cli.NewInput("Enter your SQL query:", "")
	query, err := input.Ask()
	if err != nil {
		return fmt.Errorf("failed to get query input: %w", err)
	}

	if query == "" {
		fmt.Println("Query cancelled.")
		return nil
	}

	executeAndPrint(ctx, db, query)
	return nil
}

// runScript executes SQL commands from a file
func runScript(ctx context.Context, db *database.Database, filename string) error {
	commons.DefaultLogger.Debug("Running script: %s", filename)

	data, err := os.ReadFile(filename)
	if err != nil {
		commons.DefaultLogger.Error("Failed to read file '%s': %v", filename, err)
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

	result, err := db.ExecuteQuery(ctx, query)
	if err != nil {
		return err
	}
	defer result.Close()

	fmt.Println("Script executed successfully")
	return nil
}

// exportJSON executes a query and exports results as JSON formatted output
func exportJSON(ctx context.Context, db *database.Database, query string) {
	commons.DefaultLogger.Debug("Exporting JSON for query")
	rows, err := db.ExecuteQuery(ctx, query)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Export failed: %v\n", err)
		return
	}
	defer rows.Close()

	columns, _ := rows.Columns()
	var result []map[string]interface{}

	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	for rows.Next() {
		if err := rows.Scan(valuePtrs...); err != nil {
			fmt.Fprintf(os.Stderr, "Export failed: %v\n", err)
			return
		}

		rowMap := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			rowMap[col] = formatValue(val)
		}
		result = append(result, rowMap)
	}

	if len(result) == 0 {
		fmt.Println("[]")
		return
	}

	jsonBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Export failed: %v\n", err)
		return
	}
	fmt.Println(string(jsonBytes))
}

// exportCSV executes a query and exports results as CSV formatted output
func exportCSV(ctx context.Context, db *database.Database, query string) {
	commons.DefaultLogger.Debug("Exporting CSV for query")
	rows, err := db.ExecuteQuery(ctx, query)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Export failed: %v\n", err)
		return
	}
	defer rows.Close()

	columns, _ := rows.Columns()

	// Print header
	fmt.Println(strings.Join(columns, ","))

	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	csvFormatter := commons.CSVFormatter{}
	for rows.Next() {
		if err := rows.Scan(valuePtrs...); err != nil {
			fmt.Fprintf(os.Stderr, "Export failed: %v\n", err)
			return
		}

		var row []string
		for i := range values {
			row = append(row, csvFormatter.FormatValue(values[i]))
		}
		fmt.Println(strings.Join(row, ","))
	}
}

// formatValue converts an interface value to its string representation for display
func formatValue(v interface{}) string {
	commons.DefaultLogger.Debug("Formatting value for output")

	if v == nil {
		return "NULL"
	}

	switch val := v.(type) {
	case []byte:
		return string(val)
	default:
		return fmt.Sprintf("%v", val)
	}
}

// min returns the smaller of two integer values
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// showTableContent retrieves and displays all rows from a specified table with pagination
func showTableContent(ctx context.Context, db *database.Database, tableName string) error {
	commons.DefaultLogger.Debug("Showing table content: %s", tableName)

	// First get total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)
	countRows, err := db.ExecuteQuery(ctx, countQuery)
	if err != nil {
		return fmt.Errorf("failed to get row count: %w", err)
	}
	var totalRecords int
	if countRows.Next() {
		countRows.Scan(&totalRecords)
	}
	countRows.Close()

	if totalRecords == 0 {
		fmt.Println("No data found in table.")
		return nil
	}

	pageSize := 20
	pagination := cli.NewPaginationSelector(totalRecords, pageSize)

	for {
		// Query current page
		offset := (pagination.GetCurrentPage() - 1) * pageSize
		query := fmt.Sprintf("SELECT * FROM %s LIMIT %d OFFSET %d", tableName, pageSize, offset)
		rows, err := db.ExecuteQuery(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to query table content: %w", err)
		}

		columns, err := rows.Columns()
		if err != nil {
			rows.Close()
			return fmt.Errorf("failed to get columns: %w", err)
		}

		tableData, err := parseTableData(rows, columns)
		rows.Close()
		if err != nil {
			return fmt.Errorf("failed to parse table data: %w", err)
		}

		if len(tableData) == 0 {
			fmt.Println("No data found on this page.")
			continue
		}

		// Set current page data for pagination selector
		pagination.SetCurrentPageData(columns, tableData)

		// Show combined page/row selection
		result, err := pagination.SelectPage()
		if err != nil {
			return fmt.Errorf("selection failed: %w", err)
		}

		switch result.Action {
		case "row":
			// Convert absolute row index to page-relative index
			pageRelativeIdx := result.RowIndex - (pagination.GetCurrentPage()-1)*pageSize
			if pageRelativeIdx >= 0 && pageRelativeIdx < len(tableData) {
				err := showRowDetail(ctx, db, tableName, columns, tableData[pageRelativeIdx])
				if err != nil {
					return fmt.Errorf("failed to show row detail: %w", err)
				}
			}
		case "add-row":
			err := addNewRow(ctx, db, tableName, columns, pagination)
			if err != nil {
				fmt.Printf("Failed to add new row: %v\n", err)
			}
		case "page":
			// Page changed, continue to next iteration to reload data
		case "exit":
			// User chose to go back
			return nil
		}
	}
}

// addNewRow handles adding a new row to the table
func addNewRow(ctx context.Context, db *database.Database, tableName string, columns []string, pagination *cli.PaginationSelector) error {
	creator := cli.NewRowCreator(columns)

	// Let user choose method and potentially get a row number to copy from
	newRow, err := creator.CreateWithMethod()
	if err != nil {
		return fmt.Errorf("failed to create row: %w", err)
	}

	// Check if the result is a copy-row sentinel
	if rowNumVal, ok := newRow["__COPY_ROW__"]; ok {
		rowNum, _ := rowNumVal.(int)
		// Fetch the source row from the database
		offset := rowNum - 1
		query := fmt.Sprintf("SELECT * FROM %s LIMIT 1 OFFSET %d", tableName, offset)
		rows, err := db.ExecuteQuery(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to fetch source row: %w", err)
		}
		defer rows.Close()

		if !rows.Next() {
			return fmt.Errorf("row %d not found", rowNum)
		}

		rowData := make(map[string]interface{})
		scanArgs := make([]interface{}, len(columns))
		scanTargets := make([]interface{}, len(columns))
		for i := range columns {
			scanTargets[i] = &scanArgs[i]
		}
		if err := rows.Scan(scanTargets...); err != nil {
			return fmt.Errorf("failed to scan source row: %w", err)
		}
		for i, col := range columns {
			if scanArgs[i] != nil {
				rowData[col] = scanArgs[i]
			}
		}

		// Set defaults from the copied row
		creator2 := cli.NewRowCreator(columns)
		creator2.SetDefaults(rowData)
		newRow, err = creator2.Create()
		if err != nil {
			return fmt.Errorf("failed to create row from copy: %w", err)
		}
	}

	// Build INSERT statement
	placeholders := make([]string, len(columns))
	values := make([]interface{}, len(columns))
	for i, col := range columns {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		values[i] = newRow[col]
	}

	insertQuery := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		tableName,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "))

	_, err = db.ExecuteNonQuery(ctx, insertQuery, values...)
	if err != nil {
		return fmt.Errorf("insert query failed: %w", err)
	}

	fmt.Println("✅ New row added successfully!")

	// Refresh the current page — reload by resetting pagination state is handled by the loop
	// The loop will re-fetch the page data on next iteration
	return nil
}

// printTableWithTruncation prints table with truncated values
func printTableWithTruncation(columns []string, tableData []map[string]interface{}, maxLen int) {
	colWidths := make([]int, len(columns))
	for i, col := range columns {
		colWidths[i] = min(len(col)+2, maxLen+2)
	}

	for _, row := range tableData {
		for i, col := range columns {
			val := cli.TruncateValue(row[col], maxLen)
			if len(val)+2 > colWidths[i] {
				colWidths[i] = min(len(val)+2, maxLen+2)
			}
		}
	}

	// Print header
	for i, col := range columns {
		fmt.Printf("%-*s", colWidths[i], col)
	}
	fmt.Println()

	// Print separator
	for _, width := range colWidths {
		fmt.Printf("%s", strings.Repeat("-", width))
	}
	fmt.Println()

	// Print rows
	for _, row := range tableData {
		for i, col := range columns {
			val := cli.TruncateValue(row[col], maxLen)
			fmt.Printf("%-*s", colWidths[i], val)
		}
		fmt.Println()
	}
}

// selectTableRow allows user to select a row or exit
func selectTableRow(tableData []map[string]interface{}, columns []string) (int, error) {
	options := make([]string, len(tableData)+1)

	for i, row := range tableData {
		var values []string
		for _, col := range columns {
			values = append(values, cli.TruncateValue(row[col], 30))
		}
		options[i] = fmt.Sprintf("Row %d: %s", i+1, strings.Join(values, ", "))
	}
	options[len(tableData)] = "[Back] Back"

	var selected string
	err := survey.AskOne(
		&survey.Select{
			Message:  "Select a row to view details or go back:",
			Options:  options,
			PageSize: 15,
		},
		&selected,
	)
	if err != nil {
		return -1, err
	}

	if selected == "[Back] Back" {
		return -1, nil
	}

	for i, option := range options {
		if option == selected {
			return i, nil
		}
	}

	return -1, nil
}

// showRowDetail shows row details with edit options
func showRowDetail(ctx context.Context, db *database.Database, tableName string, columns []string, row map[string]interface{}) error {
	for {
		viewer := cli.NewRowDetailViewer(row, columns)
		action, err := viewer.Show()
		if err != nil {
			return fmt.Errorf("row detail view failed: %w", err)
		}

		switch action {
		case "BACK":
			return nil
		case "EDIT":
			err := editRowDetail(ctx, db, tableName, columns, row)
			if err != nil {
				fmt.Printf("Edit failed: %v\n", err)
			} else {
				fmt.Println("Row updated successfully!")
				// Refresh row data
				row, err = getRowByValues(ctx, db, tableName, columns, row)
				if err != nil {
					fmt.Printf("Failed to refresh row: %v\n", err)
				}
			}
		case "DELETE":
			confirm := cli.NewConfirm("Are you sure you want to delete this row?", false)
			confirmed, err := confirm.Ask()
			if err != nil {
				return fmt.Errorf("confirmation failed: %w", err)
			}
			if confirmed {
				err := deleteRow(ctx, db, tableName, columns, row)
				if err != nil {
					fmt.Printf("Delete failed: %v\n", err)
				} else {
					fmt.Println("Row deleted successfully!")
					return nil
				}
			}
		}
	}
}

// editRowDetail edits a row in detail view
func editRowDetail(ctx context.Context, db *database.Database, tableName string, columns []string, row map[string]interface{}) error {
	editor := cli.NewRowEditor(row, columns)

	for {
		selection, err := editor.SelectColumnsToEdit()
		if err != nil {
			return fmt.Errorf("column selection failed: %w", err)
		}

		if len(selection) == 0 {
			continue
		}

		action := selection[0]

		if action == "SAVE" {
			if !editor.HasChanges() {
				fmt.Println("No changes made to save.")
				return nil
			}

			confirm := cli.NewConfirm("Are you sure you want to update this row?", false)
			confirmed, err := confirm.Ask()
			if err != nil {
				return fmt.Errorf("confirmation failed: %w", err)
			}

			if !confirmed {
				fmt.Println("Update cancelled.")
				return nil
			}

			editedRow := editor.GetEditedRow()
			if err := updateTableRow(ctx, db, tableName, columns, row, editedRow); err != nil {
				return fmt.Errorf("failed to update row: %w", err)
			}

			// Update the row data
			for k, v := range editedRow {
				row[k] = v
			}

			return nil
		}

		if action == "DISCARD" {
			fmt.Println("Changes discarded.")
			return nil
		}

		column := action
		currentValue := editor.GetEditedRow()[column]
		newValue, err := editor.EditColumn(column, currentValue)
		if err != nil {
			fmt.Printf("Failed to edit column %s: %v\n", column, err)
			continue
		}

		editor.GetEditedRow()[column] = newValue
		fmt.Printf("Updated %s: %s -> %s\n", column, cli.FormatValue(currentValue), cli.FormatValue(newValue))
	}
}

// deleteRow deletes a row from the table
func deleteRow(ctx context.Context, db *database.Database, tableName string, columns []string, row map[string]interface{}) error {
	whereClauses := make([]string, 0, len(columns))
	args := make([]interface{}, 0, len(columns))

	for _, col := range columns {
		whereClauses = append(whereClauses, fmt.Sprintf("%s = $%d", col, len(args)+1))
		args = append(args, row[col])
	}

	query := fmt.Sprintf("DELETE FROM %s WHERE %s", tableName, strings.Join(whereClauses, " AND "))

	_, err := db.ExecuteNonQuery(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("delete query failed: %w", err)
	}

	return nil
}

// getRowByValues retrieves a row by its values (used to refresh after update)
func getRowByValues(ctx context.Context, db *database.Database, tableName string, columns []string, row map[string]interface{}) (map[string]interface{}, error) {
	whereClauses := make([]string, 0, len(columns))
	args := make([]interface{}, 0, len(columns))

	for _, col := range columns {
		whereClauses = append(whereClauses, fmt.Sprintf("%s = $%d", col, len(args)+1))
		args = append(args, row[col])
	}

	query := fmt.Sprintf("SELECT * FROM %s WHERE %s", tableName, strings.Join(whereClauses, " AND "))

	rows, err := db.ExecuteQuery(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query row: %w", err)
	}
	defer rows.Close()

	result, err := parseTableData(rows, columns)
	if err != nil {
		return nil, fmt.Errorf("failed to parse row data: %w", err)
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("row not found")
	}

	return result[0], nil
}

// parseTableData converts SQL rows into a slice of map[string]interface{} for further processing
func parseTableData(rows *sql.Rows, columns []string) ([]map[string]interface{}, error) {
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

// printTable renders table data in a formatted ASCII table to stdout
func printTable(columns []string, tableData []map[string]interface{}) {
	colWidths := make([]int, len(columns))
	for i, col := range columns {
		colWidths[i] = len(col)
	}

	for _, row := range tableData {
		for i, col := range columns {
			val := formatValue(row[col])
			if len(val) > colWidths[i] {
				colWidths[i] = len(val)
			}
		}
	}

	for i, col := range columns {
		fmt.Printf("%-*s", colWidths[i]+2, col)
	}
	fmt.Println()

	for _, width := range colWidths {
		fmt.Printf("%s", strings.Repeat("-", width+2))
	}
	fmt.Println()

	for _, row := range tableData {
		for i, col := range columns {
			val := formatValue(row[col])
			fmt.Printf("%-*s", colWidths[i]+2, val)
		}
		fmt.Println()
	}
}

// updateTableRow updates a row in the specified table with new values
func updateTableRow(ctx context.Context, db *database.Database, tableName string, columns []string, oldRow, newRow map[string]interface{}) error {
	setClauses := make([]string, 0, len(columns))
	whereClauses := make([]string, 0, len(columns))
	args := make([]interface{}, 0, len(columns)*2)

	for _, col := range columns {
		newValue := newRow[col]
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", col, len(args)+1))
		args = append(args, newValue)
	}

	for _, col := range columns {
		oldValue := oldRow[col]
		whereClauses = append(whereClauses, fmt.Sprintf("%s = $%d", col, len(args)+1))
		args = append(args, oldValue)
	}

	query := fmt.Sprintf("UPDATE %s SET %s WHERE %s", tableName, strings.Join(setClauses, ", "), strings.Join(whereClauses, " AND "))

	_, err := db.ExecuteNonQuery(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("update query failed: %w", err)
	}

	return nil
}
