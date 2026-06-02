package formatter

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/colin2wang/postgresql-client/internal/cli"
	"github.com/colin2wang/postgresql-client/internal/commons"
)

// QueryExecutor defines the interface for executing queries
type QueryExecutor interface {
	ExecuteQuery(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}

// FormatValue converts an interface value to its string representation
func FormatValue(v interface{}) string {
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// PrintTable renders table data in a formatted ASCII table to stdout
func PrintTable(columns []string, tableData []map[string]interface{}) {
	colWidths := make([]int, len(columns))
	for i, col := range columns {
		colWidths[i] = len(col)
	}
	for _, row := range tableData {
		for i, col := range columns {
			val := FormatValue(row[col])
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
			val := FormatValue(row[col])
			fmt.Printf("%-*s", colWidths[i]+2, val)
		}
		fmt.Println()
	}
}

// PrintTableWithTruncation prints table with truncated values
func PrintTableWithTruncation(columns []string, tableData []map[string]interface{}, maxLen int) {
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
	for i, col := range columns {
		fmt.Printf("%-*s", colWidths[i], col)
	}
	fmt.Println()
	for _, width := range colWidths {
		fmt.Printf("%s", strings.Repeat("-", width))
	}
	fmt.Println()
	for _, row := range tableData {
		for i, col := range columns {
			val := cli.TruncateValue(row[col], maxLen)
			fmt.Printf("%-*s", colWidths[i], val)
		}
		fmt.Println()
	}
}

// ExecuteAndPrint executes a query and prints results
func ExecuteAndPrint(ctx context.Context, db QueryExecutor, query string) {
	commons.DefaultLogger.Debug("Executing query: %s", query[:min(len(query), 50)])
	result, err := db.ExecuteQuery(ctx, query)
	if err != nil {
		commons.DefaultLogger.Error("Query execution failed: %v", err)
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer result.Close()
	fmt.Println("Query executed successfully")
}

// ExportJSON executes a query and exports results as JSON
func ExportJSON(ctx context.Context, db QueryExecutor, query string) {
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
			rowMap[col] = FormatValue(values[i])
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

// ExportCSV executes a query and exports results as CSV
func ExportCSV(ctx context.Context, db QueryExecutor, query string) {
	rows, err := db.ExecuteQuery(ctx, query)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Export failed: %v\n", err)
		return
	}
	defer rows.Close()

	columns, _ := rows.Columns()
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
