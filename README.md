# PostgreSQL CLI Client

A modular PostgreSQL command-line client written in Go with YAML configuration support.

## Features

- Interactive and non-interactive modes
- SQL query execution with formatted output
- Table listing and description
- Database listing
- JSON and CSV export
- Command history
- YAML configuration support
- **Interactive database selection with arrow keys**
- **Interactive table selection with arrow keys**

## Project Structure

```
.
├── config/          # Configuration management
│   └── config.go    # Config loading from YAML/env
├── connector/       # Database connection handling
│   └── connector.go
├── database/        # Database operations
│   └── database.go
├── ui/              # Interactive UI components
│   └── ui.go        # Survey-based interactive selectors
├── utils/           # Utility functions
│   └── utils.go     # Formatting, history, etc.
├── main.go          # Main entry point
├── config.example.yaml  # Example configuration
├── build.sh         # Linux/macOS build script
├── build.ps1        # Windows build script
└── go.mod           # Go module definition
```

## Configuration

Create a `config.yaml` file with your database connection details:

```yaml
host: localhost
port: 5432
user: postgres
password: your_password
database: postgres
ssl_mode: disable
```

Or use environment variables:
- `PGHOST`
- `PGPORT`
- `PGUSER`
- `PGPASSWORD`
- `PGDATABASE`

## Usage

### Interactive Mode

```bash
./postgresql-client
```

### Non-Interactive Mode

```bash
# Execute a SQL command
./postgresql-client -c "SELECT * FROM table"

# Run SQL from file
./postgresql-client -f script.sql

# Export as JSON
./postgresql-client --json "SELECT * FROM table"

# Export as CSV
./postgresql-client --csv "SELECT * FROM table"
```

### Command Line Flags

```
-c, --config      Path to config file
-h, --host        Database host
-p, --port        Database port (default: 5432)
-U, --user        Database user
-W, --password    Database password
-d, --database    Database name
```

### Interactive Commands

- `\q`, `quit`, `exit` - Quit the client
- `\h`, `\?`, `help` - Show help message
- `\l`, `\list` - List all databases
- `\dt` - List all tables
- `\d <table>` - Describe table structure
- `\s`, `\select-db` - Select database with arrow keys
- `\t`, `\select-table` - Select table with arrow keys
- `\c <db>` - Connect to database (triggers arrow key selection if empty)

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
```

## Requirements

- Go 1.25+
- PostgreSQL server

## Interactive Selection

The client supports interactive database and table selection using arrow keys:

### Select Database
Use `\s` or `\select-db` command to open a selector with all available databases. Use up/down arrows to navigate and Enter to select.

### Select Table
Use `\t` or `\select-table` command to open a selector with all available tables. Use up/down arrows to navigate and Enter to select. The table structure will be displayed after selection.

## License

MIT