package commons

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"
)

// Logger provides a centralized logging interface
type Logger struct {
	level      LogLevel
	logFile    *os.File
	output     *os.File
	timeFormat string
}

// LogLevel represents the log level
type LogLevel int

const (
	Debug LogLevel = iota
	Info
	Warn
	Error
)

// LogEntry represents a single log entry
type LogEntry struct {
	Timestamp time.Time
	Level     LogLevel
	Message   string
	Context   map[string]interface{}
}

// LoggerOption configures the logger
type LoggerOption func(*Logger)

// WithLogLevel sets the log level
func WithLogLevel(level LogLevel) LoggerOption {
	return func(l *Logger) {
		l.level = level
	}
}

// WithLogFile sets the log file
func WithLogFile(file *os.File) LoggerOption {
	return func(l *Logger) {
		l.logFile = file
	}
}

// NewLogger creates a new logger instance
func NewLogger(opts ...LoggerOption) *Logger {
	logger := &Logger{
		output:     os.Stdout,
		level:      Info,
		timeFormat: "2006-01-02 15:04:05",
	}

	for _, opt := range opts {
		opt(logger)
	}

	return logger
}

// Log logs a message with the given level
func (l *Logger) Log(level LogLevel, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	entry := &LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   fmt.Sprintf(format, args...),
		Context:   make(map[string]interface{}),
	}

	l.write(entry)
}

// LogWithContext logs a message with context information
func (l *Logger) LogWithContext(ctx context.Context, level LogLevel, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	entry := &LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   fmt.Sprintf(format, args...),
		Context:   extractContext(ctx),
	}

	l.write(entry)
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	l.Log(Debug, format, args...)
}

// Info logs an info message
func (l *Logger) Info(format string, args ...interface{}) {
	l.Log(Info, format, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(format string, args ...interface{}) {
	l.Log(Warn, format, args...)
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	l.Log(Error, format, args...)
}

func (l *Logger) write(entry *LogEntry) {
	msg := fmt.Sprintf("[%s] [%s] %s",
		entry.Timestamp.Format(l.timeFormat),
		levelToString(entry.Level),
		entry.Message,
	)

	if entry.Context != nil && len(entry.Context) > 0 {
		msg += formatContext(entry.Context)
	}

	fmt.Fprintln(l.output, msg)

	if l.logFile != nil {
		fmt.Fprintln(l.logFile, msg)
	}
}

func levelToString(level LogLevel) string {
	switch level {
	case Debug:
		return "DEBUG"
	case Info:
		return "INFO"
	case Warn:
		return "WARN"
	case Error:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

func formatContext(ctx map[string]interface{}) string {
	if ctx == nil || len(ctx) == 0 {
		return ""
	}

	var parts []string
	for k, v := range ctx {
		parts = append(parts, fmt.Sprintf("%s=%v", k, v))
	}

	return " [" + join(parts, ", ") + "]"
}

func join(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	if len(strs) == 1 {
		return strs[0]
	}

	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}

func extractContext(ctx context.Context) map[string]interface{} {
	result := make(map[string]interface{})
	if ctx == nil {
		return result
	}

	// Add standard context values if available
	if val := ctx.Value("request_id"); val != nil {
		result["request_id"] = val
	}
	if val := ctx.Value("user_id"); val != nil {
		result["user_id"] = val
	}

	return result
}

// DefaultLogger is the default logger instance
var DefaultLogger *Logger

func init() {
	DefaultLogger = NewLogger()
}

// ContextKey represents a context key type for context values
type ContextKey string

const (
	RequestIDContextKey ContextKey = "request_id"
	UserIDContextKey    ContextKey = "user_id"
)

// WithRequestID adds a request ID to the context
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDContextKey, requestID)
}

// GetRequestID extracts the request ID from context
func GetRequestID(ctx context.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}
	val := ctx.Value(RequestIDContextKey)
	str, ok := val.(string)
	return str, ok
}

// Error types for common error cases

// DatabaseError represents a database-related error
type DatabaseError struct {
	Message     string
	ErrorCode   string
	SQLState    string
	StatusCode  int
	OriginalErr error
}

func (e *DatabaseError) Error() string {
	if e.OriginalErr != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.OriginalErr)
	}
	return e.Message
}

func (e *DatabaseError) Unwrap() error {
	return e.OriginalErr
}

// ConfigError represents a configuration-related error
type ConfigError struct {
	Message     string
	ConfigKey   string
	EnvVar      string
	StatusCode  int
	OriginalErr error
}

func (e *ConfigError) Error() string {
	msg := e.Message
	if e.ConfigKey != "" {
		msg += fmt.Sprintf(" (config key: %s)", e.ConfigKey)
	}
	if e.EnvVar != "" {
		msg += fmt.Sprintf(" (env var: %s)", e.EnvVar)
	}
	if e.OriginalErr != nil {
		msg += fmt.Sprintf(": %v", e.OriginalErr)
	}
	return msg
}

func (e *ConfigError) Unwrap() error {
	return e.OriginalErr
}

// ConnectionError represents a connection-related error
type ConnectionError struct {
	Message     string
	Host        string
	Port        int
	Database    string
	Timeout     time.Duration
	StatusCode  int
	OriginalErr error
}

func (e *ConnectionError) Error() string {
	msg := e.Message
	if e.Host != "" || e.Port > 0 {
		msg += fmt.Sprintf(" (host=%s, port=%d)", e.Host, e.Port)
	}
	if e.Database != "" {
		msg += fmt.Sprintf(", db=%s", e.Database)
	}
	if e.OriginalErr != nil {
		msg += fmt.Sprintf(": %v", e.OriginalErr)
	}
	return msg
}

func (e *ConnectionError) Unwrap() error {
	return e.OriginalErr
}

// ValidationError represents a validation-related error
type ValidationError struct {
	Message string
	Field   string
	Value   interface{}
	Code    string
	Details []string
}

func (e *ValidationError) Error() string {
	msg := e.Message
	if e.Field != "" {
		msg += fmt.Sprintf(" [field=%s]", e.Field)
	}
	if len(e.Details) > 0 {
		msg += ": " + join(e.Details, "; ")
	}
	return msg
}

// FileError represents a file-related error
type FileError struct {
	Message     string
	Path        string
	Action      string // read, write, create, delete
	OriginalErr error
}

func (e *FileError) Error() string {
	msg := e.Message
	if e.Path != "" {
		msg += fmt.Sprintf(" (path=%s)", e.Path)
	}
	if e.Action != "" {
		msg += fmt.Sprintf(", action=%s", e.Action)
	}
	if e.OriginalErr != nil {
		msg += fmt.Sprintf(": %v", e.OriginalErr)
	}
	return msg
}

func (e *FileError) Unwrap() error {
	return e.OriginalErr
}

// Utility functions

// IsContextCancelled checks if the context was cancelled
func IsContextCancelled(err error) bool {
	if err == nil {
		return false
	}
	return err.Error() == "context canceled" || err.Error() == "context deadline exceeded"
}

// SafeClose closes a closable resource and logs any errors
func SafeClose(closer interface{ Close() error }) error {
	if closer != nil {
		return closer.Close()
	}
	return nil
}

// FormatDuration formats a duration for display
func FormatDuration(duration time.Duration) string {
	if duration < time.Second {
		return fmt.Sprintf("%.2f ms", duration.Seconds()*1000)
	}
	if duration < time.Minute {
		return fmt.Sprintf("%.2f sec", duration.Seconds())
	}
	if duration < time.Hour {
		return fmt.Sprintf("%.2f min", duration.Minutes())
	}
	return fmt.Sprintf("%.2f hours", duration.Hours())
}

// RepeatString repeats a string n times
func RepeatString(s string, n int) string {
	if n <= 0 || s == "" {
		return ""
	}
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}

// TruncateString truncates a string to maxLen characters
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}

// Contains checks if a string contains a substring
func Contains(s, substr string) bool {
	return index(s, substr) >= 0
}

func index(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// ReplaceAll replaces all occurrences of old in s with new
func ReplaceAll(s, old, new string) string {
	result := ""
	i := 0
	for {
		j := index(s, old)
		if j < 0 {
			result += s[i:]
			break
		}
		result += s[i:j] + new
		i = j + len(old)
	}
	return result
}

// ToSliceString converts interface{} slice to string slice
func ToSliceString(slice interface{}) []string {
	switch v := slice.(type) {
	case []interface{}:
		result := make([]string, len(v))
		for i, val := range v {
			result[i] = ToString(val)
		}
		return result
	default:
		return nil
	}
}

// ToString converts any value to string
func ToString(v interface{}) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case []byte:
		return string(val)
	case string:
		return val
	default:
		return fmt.Sprintf("%v", val)
	}
}

// ParseInt parses a string to int with default value
func ParseInt(s string, def int) int {
	if s == "" {
		return def
	}
	val, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return val
}
