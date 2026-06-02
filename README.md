# PostgreSQL CLI Client

A modular PostgreSQL command-line client written in Go with YAML configuration support.

## Features

- **Interactive and Non-Interactive Modes**
  - Interactive command-line interface
  - SQL query execution and script file running

- **Data Operations**
  - SQL query execution (SELECT, INSERT, UPDATE, DELETE, etc.)
  - JSON and CSV export
  - Table content pagination
  - Row viewing, editing, and deletion

- **Interactive Selection**
  - Database selection with arrow keys
  - Table selection with arrow keys
  - Table operation menu (view structure, view content)
  - Main menu integrating all available actions

- **Configuration Management**
  - YAML configuration file support
  - Environment variable support
  - Command-line argument overrides

## Project Structure

```
.
├── internal/              # Internal packages
│   ├── cli/               # CLI interactive components
│   │   └── interface.go   # Survey-based interactive selectors (DB, table, menu, row)
│   ├── commons/           # Shared utilities, logging, error types
│   │   └── commons.go     # Logger, error types, formatters, utility functions
│   ├── config/            # Configuration management
│   │   └── config.go      # Config loading from YAML/env
│   └── database/          # Database operations
│       └── database.go    # Database connection and query execution
├── main.go                # Main entry point
├── config.example.yaml    # Example configuration
├── build.sh               # Linux/macOS build script
├── build.ps1              # Windows build script
└── go.mod                 # Go module definition
```

## Configuration

### YAML Configuration File

Create a `config.yaml` file (or use `-c` flag to specify a path):

```yaml
host: localhost
port: 5432
user: postgres
password: your_password
database: postgres
ssl_mode: disable
```

### Environment Variables

Configure connection information using the following environment variables:

- `PGHOST` - Database host (default: localhost)
- `PGPORT` - Database port (default: 5432)
- `PGUSER` - Database user (default: postgres)
- `PGPASSWORD` - Database password
- `PGDATABASE` - Database name (default: postgres)

### Command-Line Arguments

Command-line arguments override values from config files and environment variables:

```
-c, --config      Path to configuration file
-h, --host        Database host
-p, --port        Database port (default: 5432)
-U, --user        Database user
-W, --password    Database password
-d, --database    Database name
```

## Usage

### Interactive Mode

```bash
./postgresql-client
```

Enter interactive command-line mode with the following commands:

| Command | Description |
|---------|-------------|
| `\q`, `quit`, `exit` | Quit the client |
| `\h`, `\?`, `help` | Show help message |
| `\m`, `\menu` | Open main menu (all available actions) |
| `\l`, `\list` | List all databases |
| `\dt`, `\d tables` | List all tables |
| `\d`, `\D` | Interactive table structure description |
| `\s`, `\select-db` | Select database with arrow keys |
| `\t`, `\select-table` | Select table (view structure/content) |
| `\c`, `\C` | Select and connect to database |

### Non-Interactive Mode

```bash
# Execute SQL command
./postgresql-client -c "SELECT * FROM table"

# Run SQL from file
./postgresql-client -f script.sql

# Export as JSON
./postgresql-client --json "SELECT * FROM table"

# Export as CSV
./postgresql-client --csv "SELECT * FROM table"
```

## Table Operations

### View Table Structure
```sql
\d table_name
```
Displays column names, data types, nullability, and default values.

### Browse Table Content with Pagination
Use `\t` or `\select-table` command to select a table for:
- Navigate through paginated results (20 rows per page)
- Jump to specific page or row number
- View row details, edit, or delete

### Main Menu Actions
Enter `\m` or `\menu` to open the main menu with options:
1. List all databases
2. List all tables
3. Select and describe table structure
4. Select and show table content
5. Select and connect to database
6. Execute custom SQL query
7. Show help information
8. Quit

## Building

### Using Build Scripts

**Linux/macOS:**
```bash
./build.sh
```

**Windows (PowerShell):**
```powershell
.\build.ps1
```

### Manual Build

```bash
# Default build
go build -o postgresql-client .

# Cross-compile for Linux
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o postgresql-client-linux .

# Cross-compile for Windows
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o postgresql-client.exe .
```

## Requirements

- Go 1.25+
- PostgreSQL 9.x+ server
- Network connectivity (for database connections)

## Dependencies

| Component | Technology |
|-----------|------------|
| Database Driver | pgx/v5 |
| YAML Parsing | yaml.v3 |
| Interactive UI | survey/v2 |

## License

MIT