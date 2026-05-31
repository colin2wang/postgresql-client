package utils

import (
	"time"

	"github.com/colin2wang/postgresql-client/commons"
)

// Formatters for data formatting
type Formatter struct{}

// FormatValue formats a value for display
func (f *Formatter) FormatValue(v interface{}) string {
	commons.DefaultLogger.Debug("Formatting value for display")
	if v == nil {
		return "NULL"
	}

	switch val := v.(type) {
	case []byte:
		return string(val)
	case time.Time:
		return val.Format("2006-01-02 15:04:05")
	default:
		return commons.ToString(val)
	}
}

// FormatDuration formats a duration for display
func FormatDuration(duration time.Duration) string {
	commons.DefaultLogger.Debug("Formatting duration: %v", duration)
	return commons.FormatDuration(duration)
}

// ToString converts any value to string
func ToString(v interface{}) string {
	return commons.ToString(v)
}

// ToStringDefault converts any value to string with default value
func ToStringDefault(v interface{}, def string) string {
	if v == nil {
		return def
	}
	switch val := v.(type) {
	case []byte:
		return string(val)
	case string:
		return val
	default:
		return commons.ToString(val)
	}
}

// CSVFormatter for CSV output
type CSVFormatter struct{}

// FormatValue formats a value for CSV output
func (f *CSVFormatter) FormatValue(v interface{}) string {
	if v == nil {
		return "NULL"
	}

	str := ToString(v)
	if commons.Contains(str, ",") || commons.Contains(str, "\"") || commons.Contains(str, "\n") {
		str = "\"" + commons.ReplaceAll(str, "\"", "\"\"") + "\""
	}
	return str
}

// Contains checks if a string contains a substring
func Contains(s, substr string) bool {
	return commons.Contains(s, substr)
}

// ReplaceAll replaces all occurrences of old in s with new
func ReplaceAll(s, old, new string) string {
	return commons.ReplaceAll(s, old, new)
}

// Index returns the index of the first instance of substr in s, or -1 if not found
func Index(s, substr string, start int) int {
	if start < 0 {
		start = 0
	}
	for i := start; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// History manages command history
type History struct {
	commands []string
	maxSize  int
}

// NewHistory creates a new history manager
func NewHistory(maxSize int) *History {
	commons.DefaultLogger.Debug("Creating history with maxSize=%d", maxSize)
	return &History{
		commands: make([]string, 0, maxSize),
		maxSize:  maxSize,
	}
}

// Add adds a command to history
func (h *History) Add(cmd string) {
	if cmd == "" {
		return
	}
	h.commands = append(h.commands, cmd)
	if len(h.commands) > h.maxSize {
		h.commands = h.commands[1:]
	}
}

// List returns all commands in history
func (h *History) List() []string {
	return h.commands
}

// Clear clears the history
func (h *History) Clear() {
	h.commands = nil
}
