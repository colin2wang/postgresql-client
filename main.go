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

	"github.com/colin2wang/postgresql-client/commons"
	"github.com/colin2wang/postgresql-client/config"
	"github.com/colin2wang/postgresql-client/database"
	"github.com/colin2wang/postgresql-client/utils"
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

	history := utils.NewHistory(100)
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

		default:
			if strings.HasPrefix(query, "\\d ") || strings.HasPrefix(query, "\\D ") {
				tableName := strings.TrimSpace(strings.TrimPrefix(query, "\\d "))
				tableName = strings.ToLower(tableName)
				if err := describeTable(ctx, db, tableName); err != nil {
					fmt.Printf("Error: %v\n", err)
				}
			} else if query != "" {
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
	fmt.Println("  - \\q              - Quit the client")
	fmt.Println("  - \\h              - Show this help message")
	fmt.Println("  - \\l              - List all databases")
	fmt.Println("  - \\dt             - List all tables")
	fmt.Println("  - \\d <table>      - Describe table structure")
	// Empty line for readability
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
	fmt.Printf("%-30s %s\n", "Table Name", "Type")
	fmt.Println(strings.Repeat("-", 50))

	rowCount := 0
	for rows.Next() {
		var name, typ string
		if err := rows.Scan(&name, &typ); err != nil {
			return fmt.Errorf("scan failed: %w", err)
		}
		fmt.Printf("%-30s %s\n", name, typ)
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

	csvFormatter := utils.CSVFormatter{}
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

// min returns the smaller of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
