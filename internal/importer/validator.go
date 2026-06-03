package importer

import (
	"fmt"
	"regexp"
	"strings"
)

// validateSQL performs basic validation on SQL syntax
func validateSQL(sql string) error {
	trimmed := strings.TrimSpace(sql)
	if trimmed == "" {
		return fmt.Errorf("empty SQL statement")
	}

	// Check for balanced quotes
	if err := checkBalancedQuotes(trimmed); err != nil {
		return err
	}

	// For DDL, check basic CREATE TABLE structure
	if isDDLStatement(trimmed) {
		return validateDDL(trimmed)
	}

	return nil
}

// checkBalancedQuotes verifies that single quotes are balanced
func checkBalancedQuotes(s string) error {
	inSingleQuote := false
	inDoubleQuote := false
	escapeNext := false

	for _, ch := range s {
		if escapeNext {
			escapeNext = false
			continue
		}

		switch ch {
		case '\\':
			escapeNext = true
		case '\'':
			if !inDoubleQuote {
				inSingleQuote = !inSingleQuote
			}
		case '"':
			if !inSingleQuote {
				inDoubleQuote = !inDoubleQuote
			}
		}
	}

	if inSingleQuote {
		return fmt.Errorf("unterminated single-quoted string")
	}
	if inDoubleQuote {
		return fmt.Errorf("unterminated double-quoted identifier")
	}
	return nil
}

// isDDLStatement checks if the SQL contains DDL statements
func isDDLStatement(sql string) bool {
	upper := strings.ToUpper(strings.TrimSpace(sql))
	ddlPrefixes := []string{"CREATE", "ALTER", "DROP", "TRUNCATE", "COMMENT"}
	for _, prefix := range ddlPrefixes {
		if strings.HasPrefix(upper, prefix) {
			return true
		}
	}
	return false
}

// validateDDL performs additional validation on DDL statements
func validateDDL(sql string) error {
	upper := strings.ToUpper(strings.TrimSpace(sql))

	switch {
	case strings.HasPrefix(upper, "CREATE TABLE"):
		return validateCreateTable(sql)
	case strings.HasPrefix(upper, "CREATE"):
		// Other CREATE statements (index, view, etc.)
		return nil
	case strings.HasPrefix(upper, "ALTER"):
		return nil
	case strings.HasPrefix(upper, "DROP"):
		return nil
	}

	return nil
}

// validateCreateTable validates a CREATE TABLE statement
func validateCreateTable(sql string) error {
	// Check for table name
	if !regexp.MustCompile(`(?i)CREATE\s+TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?\S+\s*\(`).MatchString(sql) {
		return fmt.Errorf("invalid CREATE TABLE syntax: missing table name or opening parenthesis")
	}

	// Check for balanced parentheses (rough check)
	openCount := strings.Count(sql, "(")
	closeCount := strings.Count(sql, ")")
	if openCount != closeCount {
		return fmt.Errorf("unbalanced parentheses in CREATE TABLE statement: %d open vs %d close", openCount, closeCount)
	}

	// Check for at least one column definition
	parenContent := extractParenthesesContent(sql)
	if parenContent == "" || strings.TrimSpace(parenContent) == "" {
		return fmt.Errorf("CREATE TABLE must have at least one column definition")
	}

	return nil
}

// extractParenthesesContent extracts content between outermost parentheses
func extractParenthesesContent(s string) string {
	start := strings.Index(s, "(")
	if start == -1 {
		return ""
	}
	end := strings.LastIndex(s, ")")
	if end == -1 || end <= start {
		return ""
	}
	return s[start+1 : end]
}

// validateCSV validates CSV header and data rows
func validateCSV(headers []string, dataRows [][]string) error {
	if len(headers) == 0 {
		return fmt.Errorf("CSV file has no headers")
	}

	// Check for empty or duplicate headers
	headerSet := make(map[string]bool)
	for i, h := range headers {
		h = strings.TrimSpace(h)
		if h == "" {
			return fmt.Errorf("empty header at column %d", i+1)
		}
		if headerSet[h] {
			return fmt.Errorf("duplicate header '%s' at column %d", h, i+1)
		}
		headerSet[h] = true
	}

	// Validate data rows
	expectedColumns := len(headers)
	for i, row := range dataRows {
		if len(row) == 0 {
			return fmt.Errorf("empty data row at line %d", i+2)
		}
		if len(row) != expectedColumns {
			return fmt.Errorf("data row %d has %d columns, expected %d (header count)",
				i+2, len(row), expectedColumns)
		}
	}

	return nil
}

// splitSQLStatements splits SQL text into individual statements
func splitSQLStatements(sql string) []string {
	var statements []string
	currentStmt := strings.Builder{}
	inSingleQuote := false
	inDoubleQuote := false
	escapeNext := false
	parenDepth := 0

	for _, ch := range sql {
		if escapeNext {
			currentStmt.WriteRune(ch)
			escapeNext = false
			continue
		}

		switch ch {
		case '\\':
			currentStmt.WriteRune(ch)
			escapeNext = true
		case '\'':
			if !inDoubleQuote {
				inSingleQuote = !inSingleQuote
			}
			currentStmt.WriteRune(ch)
		case '"':
			if !inSingleQuote {
				inDoubleQuote = !inDoubleQuote
			}
			currentStmt.WriteRune(ch)
		case '(':
			if !inSingleQuote && !inDoubleQuote {
				parenDepth++
			}
			currentStmt.WriteRune(ch)
		case ')':
			if !inSingleQuote && !inDoubleQuote && parenDepth > 0 {
				parenDepth--
			}
			currentStmt.WriteRune(ch)
		case ';':
			if !inSingleQuote && !inDoubleQuote && parenDepth == 0 {
				stmt := strings.TrimSpace(currentStmt.String())
				if stmt != "" {
					statements = append(statements, stmt)
				}
				currentStmt.Reset()
			} else {
				currentStmt.WriteRune(ch)
			}
		default:
			currentStmt.WriteRune(ch)
		}
	}

	// Handle last statement without trailing semicolon
	remaining := strings.TrimSpace(currentStmt.String())
	if remaining != "" {
		statements = append(statements, remaining)
	}

	return statements
}
