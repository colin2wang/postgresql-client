package utils

import (
	"time"
)

// Formatters for data formatting
type Formatter struct{}

// FormatValue formats a value for display
func (f *Formatter) FormatValue(v interface{}) string {
	if v == nil {
		return "NULL"
	}

	switch val := v.(type) {
	case []byte:
		return string(val)
	case time.Time:
		return val.Format("2006-01-02 15:04:05")
	default:
		return ToString(val)
	}
}

// FormatDuration formats a duration for display
func FormatDuration(duration time.Duration) string {
	return ToString(duration.Seconds()) + " sec"
}

// ToString converts any value to string
func ToString(v interface{}) string {
	return ToStringDefault(v, "")
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
		return ToString(val)
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
	if Contains(str, ",") || Contains(str, "\"") || Contains(str, "\n") {
		str = "\"" + ReplaceAll(str, "\"", "\"\"") + "\""
	}
	return str
}

// Contains checks if a string contains a substring
func Contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// ReplaceAll replaces all occurrences of old in s with new
func ReplaceAll(s, old, new string) string {
	result := ""
	i := 0
	for {
		j := Index(s, old, i)
		if j < 0 {
			result += s[i:]
			break
		}
		result += s[i:j] + new
		i = j + len(old)
	}
	return result
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
