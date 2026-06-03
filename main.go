package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/colin2wang/postgresql-client/internal/cli"
	"github.com/colin2wang/postgresql-client/internal/commons"
	"github.com/colin2wang/postgresql-client/internal/config"
	"github.com/colin2wang/postgresql-client/internal/database"
	"github.com/colin2wang/postgresql-client/internal/formatter"
	"github.com/colin2wang/postgresql-client/internal/importer"
)

func main() {
	ctx := context.Background()

	configPath := flag.String("c", "", "Path to config file")
	host := flag.String("h", "", "Database host")
	port := flag.Int("p", 0, "Database port")
	user := flag.String("U", "", "Database user")
	password := flag.String("W", "", "Database password")
	databaseName := flag.String("d", "", "Database name")
	importDir := flag.String("i", "", "Import file or directory path")
	flag.Parse()

	if logDir, err := os.Getwd(); err == nil {
		logPath := logDir + "\\postgresql-client.log"
		if err := commons.DefaultLogger.SetLogFile(logPath); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to open log file: %v\n", err)
		} else {
			commons.DefaultLogger.Info("Log file: %s", logPath)
		}
	}

	cfg, err := loadConfig(*configPath, *host, *port, *user, *password, *databaseName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	db, q, err := database.NewDatabaseAndQuery(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	commons.DefaultLogger.Info("Configuration loaded successfully")

	// Handle non-interactive import mode
	if *importDir != "" {
		handleImportMode(ctx, db, q, cfg, *importDir)
		return
	}

	if flag.NArg() > 0 {
		handleNonInteractive(ctx, db, q, flag.Args())
		return
	}
	runInteractive(ctx, db, q, cfg)
}

func loadConfig(configPath string, host string, port int, user, password, databaseName string) (*config.Config, error) {
	commons.DefaultLogger.Info("Loading configuration...")
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return nil, &commons.ConfigError{
			Message:     "failed to load configuration",
			OriginalErr: err,
		}
	}
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

func handleNonInteractive(ctx context.Context, db *database.Database, q *database.Query, args []string) {
	query := strings.Join(args, " ")

	switch args[0] {
	case "-f", "--file":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Error: -f requires a filename")
			os.Exit(1)
		}
		if err := q.ExecuteScript(ctx, args[1]); err != nil {
			fmt.Fprintf(os.Stderr, "Script execution failed: %v\n", err)
			os.Exit(1)
		}
	case "-c", "--command":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Error: -c requires a SQL command")
			os.Exit(1)
		}
		formatter.ExecuteAndPrint(ctx, db, query)
	case "--json":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Error: --json requires a SQL query")
			os.Exit(1)
		}
		formatter.ExportJSON(ctx, db, query)
	case "--csv":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Error: --csv requires a SQL query")
			os.Exit(1)
		}
		formatter.ExportCSV(ctx, db, query)
	default:
		formatter.ExecuteAndPrint(ctx, db, query)
	}
}

func runInteractive(ctx context.Context, db *database.Database, q *database.Query, cfg *config.Config) {
	commons.DefaultLogger.Info("Starting interactive mode")
	printWelcome(cfg)

	reader := bufio.NewReader(os.Stdin)
	prompt := fmt.Sprintf("%s@%s> ", cfg.User, cfg.Database)

	for {
		fmt.Print(prompt)
		text, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("\nError reading input: %v\n", err)
			return
		}

		query := strings.TrimSpace(strings.TrimSuffix(text, "\n"))
		if query == "" {
			continue
		}

		switch query {
		case "\\m", "\\menu":
			newDb, err := showMainMenu(ctx, db, q, cfg)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			}
			if newDb != nil && newDb != db {
				db = newDb
				q = database.NewQuery(db)
				prompt = fmt.Sprintf("%s@%s> ", cfg.User, cfg.Database)
			}
		case "\\q", "quit", "exit":
			fmt.Println("Goodbye!")
			return
		case "\\h", "\\?", "help":
			printWelcome(cfg)
		case "\\l", "\\list":
			showDatabases(ctx, q)
		case "\\dt", "\\d tables":
			listTables(ctx, q)
		case "\\t", "\\select-table":
			selectTableInteractive(ctx, db, q, cfg)
		case "\\s", "\\select-db":
			newDb, err := selectDatabaseInteractive(ctx, db, q, cfg)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else if newDb != nil {
				db = newDb
				q = database.NewQuery(db)
				prompt = fmt.Sprintf("%s@%s> ", cfg.User, cfg.Database)
			}
		case "\\c", "\\C":
			newDb, err := selectDatabaseInteractive(ctx, db, q, cfg)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else if newDb != nil {
				db = newDb
				q = database.NewQuery(db)
				prompt = fmt.Sprintf("%s@%s> ", cfg.User, cfg.Database)
			}
		case "\\d", "\\D":
			selectAndDescribeTable(ctx, q, cfg)
		case "\\i", "\\import":
			showImportMenu(ctx, db, q, cfg)
		default:
			formatter.ExecuteAndPrint(ctx, db, query)
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
	fmt.Println("  - \\h              - Show this help message")
	fmt.Println("  - \\l              - List all databases")
	fmt.Println("  - \\dt             - List all tables")
	fmt.Println("  - \\d              - Describe a table")
	fmt.Println("  - \\t              - Select and show table content")
	fmt.Println("  - \\s, \\c          - Select and connect to database")
	fmt.Println("  - \\i              - Open import menu")
	fmt.Println("  - \\m              - Show main menu")
	fmt.Println("  - \\q              - Quit the client")
	fmt.Println("========================================")
}

func showDatabases(ctx context.Context, q *database.Query) {
	databases, err := q.GetAllDatabases(ctx)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Println("Found database(s):")
	fmt.Printf("%-30s\n", "Database Name")
	fmt.Println(strings.Repeat("-", 32))
	for _, db := range databases {
		fmt.Printf("%-30s\n", db.Name)
	}
	fmt.Printf("%d database(s) found\n", len(databases))
}

func listTables(ctx context.Context, q *database.Query) {
	tables, err := q.GetAllTables(ctx)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Println("Found table(s) in database:")
	fmt.Printf("%-30s %-10s %s\n", "Table Name", "Type", "Row Count")
	fmt.Println(strings.Repeat("-", 55))
	for _, table := range tables {
		fmt.Printf("%-30s %-10s %d\n", table.Name, "TABLE", table.RowCount)
	}
	fmt.Printf("%d table(s) found\n", len(tables))
}

func describeTable(ctx context.Context, q *database.Query, tableName string) error {
	columns, err := q.DescribeTable(ctx, tableName)
	if err != nil {
		return fmt.Errorf("failed to describe table: %w", err)
	}
	fmt.Printf("Structure of table '%s':\n", tableName)
	fmt.Printf("%-20s %-15s %-8s %s\n", "Column Name", "Type", "Nullable", "Default")
	fmt.Println(strings.Repeat("-", 70))
	for _, col := range columns {
		nullStr := "NO"
		if col.IsNullable {
			nullStr = "YES"
		}
		defaultVal := col.DefaultValue
		if defaultVal == "" {
			defaultVal = "<NULL>"
		}
		fmt.Printf("%-20s %-15s %-8s %s\n", col.Name, col.DataType, nullStr, defaultVal)
	}
	fmt.Printf("%d column(s) found\n", len(columns))
	return nil
}

func selectDatabaseInteractive(ctx context.Context, db *database.Database, q *database.Query, cfg *config.Config) (*database.Database, error) {
	commons.DefaultLogger.Debug("Starting interactive database selection")
	databases, err := q.GetAllDatabases(ctx)
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

func selectTableInteractive(ctx context.Context, db *database.Database, q *database.Query, cfg *config.Config) error {
	tables, err := q.GetAllTables(ctx)
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
		return describeTable(ctx, q, selectedTableName)
	case "Show table content":
		return showTableContent(ctx, db, q, selectedTableName)
	}
	return nil
}

func selectAndDescribeTable(ctx context.Context, q *database.Query, cfg *config.Config) error {
	tables, err := q.GetAllTables(ctx)
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
	return describeTable(ctx, q, selectedTableName)
}

func showMainMenu(ctx context.Context, db *database.Database, q *database.Query, cfg *config.Config) (*database.Database, error) {
	commons.DefaultLogger.Debug("Showing main menu")
	menu := cli.NewMenu("Select an action:", []string{
		"List all databases",
		"List all tables",
		"Select and describe table structure",
		"Select and show table content",
		"Select and connect to database",
		"Execute custom SQL query",
		"Import data",
		"Show help",
		"Quit",
	})

	selected, err := menu.Select()
	if err != nil {
		return nil, fmt.Errorf("menu selection failed: %w", err)
	}

	switch selected {
	case "List all databases":
		showDatabases(ctx, q)
	case "List all tables":
		listTables(ctx, q)
	case "Select and describe table structure":
		if err := selectAndDescribeTable(ctx, q, cfg); err != nil {
			return nil, fmt.Errorf("failed to select and describe table: %w", err)
		}
	case "Select and show table content":
		if err := selectAndShowTableContent(ctx, db, q, cfg); err != nil {
			return nil, fmt.Errorf("failed to select and show table content: %w", err)
		}
	case "Select and connect to database":
		newDb, err := selectDatabaseInteractive(ctx, db, q, cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to select database: %w", err)
		}
		if newDb != db {
			return newDb, nil
		}
		return db, nil
	case "Execute custom SQL query":
		executeCustomQuery(ctx, db)
	case "Import data":
		showImportMenu(ctx, db, q, cfg)
	case "Show help":
		printWelcome(cfg)
	case "Quit":
		fmt.Println("Goodbye!")
		os.Exit(0)
	}
	return db, nil
}

func selectAndShowTableContent(ctx context.Context, db *database.Database, q *database.Query, cfg *config.Config) error {
	tables, err := q.GetAllTables(ctx)
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
	return showTableContent(ctx, db, q, selectedTableName)
}

func executeCustomQuery(ctx context.Context, db *database.Database) {
	input := cli.NewInput("Enter your SQL query:", "")
	query, err := input.Ask()
	if err != nil {
		fmt.Printf("Failed to get query input: %v\n", err)
		return
	}
	if query == "" {
		fmt.Println("Query cancelled.")
		return
	}
	formatter.ExecuteAndPrint(ctx, db, query)
}

func showTableContent(ctx context.Context, db *database.Database, q *database.Query, tableName string) error {
	commons.DefaultLogger.Debug("Showing table content: %s", tableName)

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

		tableData, err := q.ParseTableData(rows, columns)
		rows.Close()
		if err != nil {
			return fmt.Errorf("failed to parse table data: %w", err)
		}

		if len(tableData) == 0 {
			fmt.Println("No data found on this page.")
			continue
		}

		pagination.SetCurrentPageData(columns, tableData)
		result, err := pagination.SelectPage()
		if err != nil {
			return fmt.Errorf("selection failed: %w", err)
		}

		switch result.Action {
		case "row":
			pageRelativeIdx := result.RowIndex - (pagination.GetCurrentPage()-1)*pageSize
			if pageRelativeIdx >= 0 && pageRelativeIdx < len(tableData) {
				err := showRowDetail(ctx, db, q, tableName, columns, tableData[pageRelativeIdx])
				if err != nil {
					return fmt.Errorf("failed to show row detail: %w", err)
				}
			}
		case "add-row":
			err := addNewRow(ctx, db, q, tableName, columns, pagination)
			if err != nil {
				fmt.Printf("Failed to add new row: %v\n", err)
			}
		case "page":
		case "exit":
			return nil
		}
	}
}

func addNewRow(ctx context.Context, db *database.Database, q *database.Query, tableName string, columns []string, pagination *cli.PaginationSelector) error {
	creator := cli.NewRowCreator(columns)
	newRow, err := creator.CreateWithMethod()
	if err != nil {
		return fmt.Errorf("failed to create row: %w", err)
	}

	if rowNumVal, ok := newRow["__COPY_ROW__"]; ok {
		rowNum, _ := rowNumVal.(int)
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

		creator2 := cli.NewRowCreator(columns)
		creator2.SetDefaults(rowData)
		newRow, err = creator2.Create()
		if err != nil {
			return fmt.Errorf("failed to create row from copy: %w", err)
		}
	}

	placeholders := make([]string, len(columns))
	values := make([]interface{}, len(columns))
	for i, col := range columns {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		values[i] = newRow[col]
	}
	insertQuery := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", tableName, strings.Join(columns, ", "), strings.Join(placeholders, ", "))
	_, err = db.ExecuteNonQuery(ctx, insertQuery, values...)
	if err != nil {
		return fmt.Errorf("insert query failed: %w", err)
	}
	fmt.Println("✅ New row added successfully!")
	return nil
}

func showRowDetail(ctx context.Context, db *database.Database, q *database.Query, tableName string, columns []string, row map[string]interface{}) error {
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
			err := editRowDetail(ctx, db, q, tableName, columns, row)
			if err != nil {
				fmt.Printf("Edit failed: %v\n", err)
			} else {
				fmt.Println("Row updated successfully!")
				row, err = q.GetRowByValues(ctx, tableName, columns, row)
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
				err := q.DeleteRow(ctx, tableName, columns, row)
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

func editRowDetail(ctx context.Context, db *database.Database, q *database.Query, tableName string, columns []string, row map[string]interface{}) error {
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
			if err := q.UpdateTableRow(ctx, tableName, columns, row, editedRow); err != nil {
				return fmt.Errorf("failed to update row: %w", err)
			}
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

// showImportMenu displays the import sub-menu
func showImportMenu(ctx context.Context, db *database.Database, q *database.Query, cfg *config.Config) {
	commons.DefaultLogger.Debug("Showing import menu")

	for {
		// Collect files from all import directories
		type taggedFile struct {
			Path      string
			Label     string
			ImportDir string
		}

		var taggedFiles []taggedFile

		scanDir := func(dir, dirLabel, tag string) {
			if dir == "" {
				return
			}
			entries, err := os.ReadDir(dir)
			if err != nil {
				return
			}
			for _, entry := range entries {
				if entry.IsDir() {
					continue
				}
				taggedFiles = append(taggedFiles, taggedFile{
					Path:      filepath.Join(dir, entry.Name()),
					Label:     fmt.Sprintf("[%s] %s", tag, entry.Name()),
					ImportDir: dirLabel,
				})
			}
		}

		scanDir(cfg.Import.DDLDir, "ddl", "DDL")
		scanDir(cfg.Import.CSVDir, "csv", "CSV")
		scanDir(cfg.Import.SQLDir, "sql", "SQL")

		if len(taggedFiles) == 0 {
			fmt.Println("No import files found in configured directories (ddl/, csv/, sql/).")
			return
		}

		// Build menu options
		options := make([]string, len(taggedFiles))
		for i, tf := range taggedFiles {
			options[i] = tf.Label
		}
		options = append(options, "Back to main menu")

		selector := cli.NewTableSelector()
		selected, err := selector.Select("Select a file to import:", options)
		if err != nil {
			fmt.Printf("Selection failed: %v\n", err)
			return
		}

		if selected == "Back to main menu" {
			return
		}

		// Find the selected tagged file
		var tf *taggedFile
		for i, opt := range options {
			if opt == selected && i < len(taggedFiles) {
				tf = &taggedFiles[i]
				break
			}
		}
		if tf == nil {
			fmt.Println("Invalid selection.")
			continue
		}

		imp := importer.NewImporter(db, q, &cfg.Import)
		ext := strings.ToLower(filepath.Ext(tf.Path))

		switch ext {
		case ".sql", ".ddl":
			// .sql/.ddl files in sql/ dir → ExecuteSQLFile; in ddl/ dir → ImportDDL
			if tf.ImportDir == "sql" {
				fmt.Printf("Executing SQL script: %s...\n", filepath.Base(tf.Path))
				err = imp.ExecuteSQLFile(ctx, tf.Path)
			} else {
				fmt.Printf("Importing DDL: %s...\n", filepath.Base(tf.Path))
				err = imp.ImportDDL(ctx, tf.Path)
			}
		case ".csv":
			fmt.Printf("Importing CSV: %s...\n", filepath.Base(tf.Path))
			err = imp.ImportCSV(ctx, tf.Path)
		default:
			fmt.Printf("Executing SQL script: %s...\n", filepath.Base(tf.Path))
			err = imp.ExecuteSQLFile(ctx, tf.Path)
		}

		if err != nil {
			fmt.Printf("Import failed: %v\n", err)
		}
		fmt.Println()
	}
}

// handleImportMode handles non-interactive import via -i flag
func handleImportMode(ctx context.Context, db *database.Database, q *database.Query, cfg *config.Config, path string) {
	fi, err := os.Stat(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: cannot access path '%s': %v\n", path, err)
		os.Exit(1)
	}

	imp := importer.NewImporter(db, q, &cfg.Import)

	if fi.IsDir() {
		// Import all files from directory based on extension
		entries, err := os.ReadDir(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to read directory '%s': %v\n", path, err)
			os.Exit(1)
		}
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			ext := strings.ToLower(filepath.Ext(entry.Name()))
			filePath := filepath.Join(path, entry.Name())
			switch ext {
			case ".sql", ".ddl":
				if err := imp.ImportDDL(ctx, filePath); err != nil {
					fmt.Fprintf(os.Stderr, "Error importing DDL '%s': %v\n", entry.Name(), err)
				}
			case ".csv":
				if err := imp.ImportCSV(ctx, filePath); err != nil {
					fmt.Fprintf(os.Stderr, "Error importing CSV '%s': %v\n", entry.Name(), err)
				}
			}
		}
		return
	}

	// Single file
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".sql", ".ddl":
		if err := imp.ImportDDL(ctx, path); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case ".csv":
		if err := imp.ImportCSV(ctx, path); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	default:
		if err := imp.ExecuteSQLFile(ctx, path); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}
}
