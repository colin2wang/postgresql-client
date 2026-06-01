package commons

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"
)

// 测试 Logger
func TestLogger(t *testing.T) {
	logger := NewLogger(WithLogLevel(Debug))

	// 测试不同级别的日志记录
	logger.Debug("Debug message")
	logger.Info("Info message")
	logger.Warn("Warning message")
	logger.Error("Error message")

	// 测试 Log 方法直接调用
	logger.Log(Debug, "Direct debug log")
	logger.Log(Info, "Direct info log")
}

func TestLoggerWithLogLevel(t *testing.T) {
	logger := NewLogger(WithLogLevel(Warn))

	// Debug 和 Info 应该被忽略（级别低于 Warn）
	buf := logger.SetOutputString()

	logger.Debug("Should not be logged")
	logger.Info("Should not be logged")
	if buf.Len() != 0 {
		t.Errorf("Expected no output for Debug/Info when level is Warn, got: %s", buf.String())
	}

	buf.Reset()

	// Warn 和 Error 应该被记录
	logger.Warn("Warning message")
	logger.Error("Error message")

	expectedContains := []string{"WARN", "ERROR"}
	for _, expected := range expectedContains {
		if !strings.Contains(buf.String(), expected) {
			t.Errorf("Expected output to contain '%s', got: %s", expected, buf.String())
		}
	}
}

func TestLoggerWithLogFile(t *testing.T) {
	// 创建临时日志文件
	tmpFile, err := os.CreateTemp("", "test-*.log")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	buf := &strings.Builder{}
	logger := NewLogger(WithLogLevel(Debug), WithLogFile(tmpFile), WithOutput(buf))

	logger.Info("Test log message")

	// 读取文件内容
	tmpFile.Seek(0, 0)
	content, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to read temp file: %v", err)
	}

	if !strings.Contains(string(content), "INFO") {
		t.Errorf("Expected log file to contain 'INFO', got: %s", string(content))
	}

	// 检查缓冲区是否有输出
	if !strings.Contains(buf.String(), "INFO") {
		t.Errorf("Expected buffer to contain 'INFO', got: %s", buf.String())
	}
}

func TestLoggerWithContext(t *testing.T) {
	logger := NewLogger(WithLogLevel(Debug))

	ctx := context.Background()
	ctx = context.WithValue(ctx, RequestIDContextKey, "req-123")
	ctx = context.WithValue(ctx, UserIDContextKey, "user-456")

	buf := logger.SetOutputString()

	logger.LogWithContext(ctx, Info, "Test message with context")

	output := buf.String()
	if !strings.Contains(output, "INFO") {
		t.Errorf("Expected output to contain 'INFO', got: %s", output)
	}

	if !strings.Contains(output, "request_id=req-123") {
		t.Errorf("Expected output to contain request_id, got: %s", output)
	}

	if !strings.Contains(output, "user_id=user-456") {
		t.Errorf("Expected output to contain user_id, got: %s", output)
	}
}

func TestLoggerWithEmptyContext(t *testing.T) {
	logger := NewLogger(WithLogLevel(Debug))

	buf := logger.SetOutputString()

	logger.LogWithContext(nil, Info, "Test message with nil context")

	output := buf.String()
	if !strings.Contains(output, "INFO") {
		t.Errorf("Expected output to contain 'INFO', got: %s", output)
	}
}

func TestDefaultLogger(t *testing.T) {
	if DefaultLogger == nil {
		t.Error("DefaultLogger should not be nil")
	}
}

// 测试日志级别转换
func TestLevelToString(t *testing.T) {
	tests := []struct {
		level    LogLevel
		expected string
	}{
		{Debug, "DEBUG"},
		{Info, "INFO"},
		{Warn, "WARN"},
		{Error, "ERROR"},
	}

	for _, tt := range tests {
		result := levelToString(tt.level)
		if result != tt.expected {
			t.Errorf("levelToString(%v) = %s, expected %s", tt.level, result, tt.expected)
		}
	}
}

// 测试错误类型
func TestDatabaseError(t *testing.T) {
	err := &DatabaseError{
		Message:     "database connection failed",
		ErrorCode:   "500",
		SQLState:    "08006",
		StatusCode:  500,
		OriginalErr: os.ErrNotExist,
	}

	output := err.Error()
	expectedSubstrings := []string{"database connection failed", "500", "08006"}
	for _, sub := range expectedSubstrings {
		if !strings.Contains(output, sub) {
			t.Errorf("Error output should contain '%s', got: %s", sub, output)
		}
	}

	// 测试 Unwrap
	unwrapped := err.Unwrap()
	if unwrapped != os.ErrNotExist {
		t.Errorf("Unwrap() returned wrong error")
	}
}

func TestConfigError(t *testing.T) {
	err := &ConfigError{
		Message:     "configuration missing",
		ConfigKey:   "database.host",
		EnvVar:      "DB_HOST",
		StatusCode:  400,
		OriginalErr: os.ErrNotExist,
	}

	output := err.Error()
	expectedSubstrings := []string{"configuration missing", "config key: database.host", "env var: DB_HOST"}
	for _, sub := range expectedSubstrings {
		if !strings.Contains(output, sub) {
			t.Errorf("Error output should contain '%s', got: %s", sub, output)
		}
	}
}

func TestConnectionError(t *testing.T) {
	err := &ConnectionError{
		Message:     "connection timeout",
		Host:        "localhost",
		Port:        5432,
		Database:    "testdb",
		Timeout:     time.Minute,
		StatusCode:  504,
		OriginalErr: os.ErrNotExist,
	}

	output := err.Error()
	expectedSubstrings := []string{"connection timeout", "host=localhost", "port=5432", "db=testdb"}
	for _, sub := range expectedSubstrings {
		if !strings.Contains(output, sub) {
			t.Errorf("Error output should contain '%s', got: %s", sub, output)
		}
	}
}

func TestValidationError(t *testing.T) {
	err := &ValidationError{
		Message: "invalid input",
		Field:   "email",
		Value:   "not-an-email",
		Code:    "INVALID_EMAIL",
		Details: []string{"must contain @", "must not be empty"},
	}

	output := err.Error()
	expectedSubstrings := []string{"invalid input", "field=email"}
	for _, sub := range expectedSubstrings {
		if !strings.Contains(output, sub) {
			t.Errorf("Error output should contain '%s', got: %s", sub, output)
		}
	}
}

func TestFileError(t *testing.T) {
	err := &FileError{
		Message:     "file not found",
		Path:        "/tmp/test.txt",
		Action:      "read",
		OriginalErr: os.ErrNotExist,
	}

	output := err.Error()
	expectedSubstrings := []string{"file not found", "path=/tmp/test.txt", "action=read"}
	for _, sub := range expectedSubstrings {
		if !strings.Contains(output, sub) {
			t.Errorf("Error output should contain '%s', got: %s", sub, output)
		}
	}
}

// 测试 Formatter
func TestFormatter(t *testing.T) {
	f := &Formatter{}

	tests := []struct {
		input    interface{}
		expected string
	}{
		{nil, "NULL"},
		{"hello", "hello"},
		{123, "123"},
		{[]byte("bytes"), "bytes"},
		{time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC), "2024-01-01 12:00:00"},
	}

	for _, tt := range tests {
		result := f.FormatValue(tt.input)
		if result != tt.expected {
			t.Errorf("FormatValue(%v) = %s, expected %s", tt.input, result, tt.expected)
		}
	}
}

func TestCSVFormatter(t *testing.T) {
	f := &CSVFormatter{}

	tests := []struct {
		input    interface{}
		expected string
	}{
		{nil, "NULL"},
		{"hello", "hello"},
		{"hello,world", `"hello,world"`},
		{`hello"world`, `"hello""world"`},
	}

	for _, tt := range tests {
		result := f.FormatValue(tt.input)
		if result != tt.expected {
			t.Errorf("FormatValue(%v) = %s, expected %s", tt.input, result, tt.expected)
		}
	}
}

// 测试上下文相关函数
func TestWithRequestID(t *testing.T) {
	ctx := context.Background()
	ctx = WithRequestID(ctx, "req-123")

	requestID, ok := GetRequestID(ctx)
	if !ok || requestID != "req-123" {
		t.Errorf("Expected 'req-123', got '%s' with ok=%v", requestID, ok)
	}
}

func TestGetRequestIDWithNilContext(t *testing.T) {
	requestID, ok := GetRequestID(nil)
	if ok || requestID != "" {
		t.Errorf("Expected empty string with ok=false for nil context")
	}
}

func TestGetRequestIDNoKey(t *testing.T) {
	ctx := context.Background()
	requestID, ok := GetRequestID(ctx)
	if ok || requestID != "" {
		t.Errorf("Expected empty string with ok=false when key not in context")
	}
}

// 测试 IsContextCancelled
func TestIsContextCancelled(t *testing.T) {
	tests := []struct {
		err      error
		expected bool
	}{
		{nil, false},
		{context.Canceled, true},
		{context.DeadlineExceeded, true},
		{os.ErrNotExist, false},
	}

	for _, tt := range tests {
		result := IsContextCancelled(tt.err)
		if result != tt.expected {
			t.Errorf("IsContextCancelled(%v) = %v, expected %v", tt.err, result, tt.expected)
		}
	}
}

// 测试工具函数
func TestFormatDuration(t *testing.T) {
	tests := []struct {
		duration time.Duration
		expected string
	}{
		{time.Millisecond, "1.00 ms"},
		{500 * time.Millisecond, "500.00 ms"},
		{time.Second, "1.00 sec"},
		{2 * time.Second, "2.00 sec"},
		{time.Minute, "1.00 min"},
		{2 * time.Minute, "2.00 min"},
		{time.Hour, "1.00 hours"},
	}

	for _, tt := range tests {
		result := FormatDuration(tt.duration)
		if result != tt.expected {
			t.Errorf("FormatDuration(%v) = %s, expected %s", tt.duration, result, tt.expected)
		}
	}
}

func TestRepeatString(t *testing.T) {
	tests := []struct {
		s        string
		n        int
		expected string
	}{
		{"a", 3, "aaa"},
		{"hello", 2, "hellohello"},
		{"", 5, ""},
		{"test", 0, ""},
		{"test", -1, ""},
	}

	for _, tt := range tests {
		result := RepeatString(tt.s, tt.n)
		if result != tt.expected {
			t.Errorf("RepeatString(%q, %d) = %s, expected %s", tt.s, tt.n, result, tt.expected)
		}
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		s        string
		maxLen   int
		expected string
	}{
		{"hello world", 5, "hello"},
		{"hi", 10, "hi"},
		{"test", 4, "test"},
		{"test", 2, "te"},
	}

	for _, tt := range tests {
		result := TruncateString(tt.s, tt.maxLen)
		if result != tt.expected {
			t.Errorf("TruncateString(%q, %d) = %s, expected %s", tt.s, tt.maxLen, result, tt.expected)
		}
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		s        string
		substr   string
		expected bool
	}{
		{"hello world", "world", true},
		{"hello", "lo", true},
		{"hello", "xyz", false},
		{"", "", true},
		{"test", "", true},
	}

	for _, tt := range tests {
		result := Contains(tt.s, tt.substr)
		if result != tt.expected {
			t.Errorf("Contains(%q, %q) = %v, expected %v", tt.s, tt.substr, result, tt.expected)
		}
	}
}

func TestIndex(t *testing.T) {
	tests := []struct {
		s        string
		substr   string
		start    int
		expected int
	}{
		{"hello world", "world", 0, 6},
		{"hello hello", "hello", 1, 6},
		{"test", "xyz", 0, -1},
		{"test", "", 0, 0},
	}

	for _, tt := range tests {
		result := Index(tt.s, tt.substr, tt.start)
		if result != tt.expected {
			t.Errorf("Index(%q, %q, %d) = %d, expected %d", tt.s, tt.substr, tt.start, result, tt.expected)
		}
	}
}

func TestReplaceAll(t *testing.T) {
	tests := []struct {
		s        string
		old      string
		new      string
		expected string
	}{
		{"hello world", "world", "go", "hello go"},
		{"test test", "test", "pass", "pass pass"},
		{"aaa", "a", "b", "bbb"},
		{"test", "xyz", "abc", "test"},
	}

	for _, tt := range tests {
		result := ReplaceAll(tt.s, tt.old, tt.new)
		if result != tt.expected {
			t.Errorf("ReplaceAll(%q, %q, %q) = %s, expected %s", tt.s, tt.old, tt.new, result, tt.expected)
		}
	}
}

func TestSafeClose(t *testing.T) {
	// 测试非 nil 的 closable
	tmpFile, err := os.CreateTemp("", "test-*.log")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	err = SafeClose(tmpFile)
	if err != nil {
		t.Errorf("SafeClose should not return error for valid file, got: %v", err)
	}

	// 测试 nil
	err = SafeClose(nil)
	if err != nil {
		t.Errorf("SafeClose should not return error for nil, got: %v", err)
	}
}

// 测试 History
func TestHistory(t *testing.T) {
	h := NewHistory(3)

	h.Add("cmd1")
	h.Add("cmd2")
	h.Add("cmd3")

	cmds := h.List()
	if len(cmds) != 3 {
		t.Errorf("Expected 3 commands, got %d", len(cmds))
	}

	if cmds[0] != "cmd1" || cmds[1] != "cmd2" || cmds[2] != "cmd3" {
		t.Errorf("Expected [cmd1, cmd2, cmd3], got %v", cmds)
	}

	// 添加第4个命令，应该移除第一个
	h.Add("cmd4")
	cmds = h.List()
	if len(cmds) != 3 {
		t.Errorf("Expected 3 commands after adding cmd4, got %d", len(cmds))
	}
	if cmds[0] != "cmd2" || cmds[1] != "cmd3" || cmds[2] != "cmd4" {
		t.Errorf("Expected [cmd2, cmd3, cmd4], got %v", cmds)
	}

	// 添加空字符串应该被忽略
	h.Add("")
	cmds = h.List()
	if len(cmds) != 3 {
		t.Errorf("Expected 3 commands after adding empty string, got %d", len(cmds))
	}

	// 清除历史
	h.Clear()
	cmds = h.List()
	if len(cmds) != 0 {
		t.Errorf("Expected 0 commands after clear, got %d", len(cmds))
	}
}

func TestToString(t *testing.T) {
	tests := []struct {
		input    interface{}
		expected string
	}{
		{nil, ""},
		{"hello", "hello"},
		{123, "123"},
		{[]byte("bytes"), "bytes"},
		{true, "true"},
	}

	for _, tt := range tests {
		result := ToString(tt.input)
		if result != tt.expected {
			t.Errorf("ToString(%v) = %s, expected %s", tt.input, result, tt.expected)
		}
	}
}

func TestToStringDefault(t *testing.T) {
	tests := []struct {
		input    interface{}
		def      string
		expected string
	}{
		{nil, "default", "default"},
		{"hello", "default", "hello"},
		{123, "default", "123"},
	}

	for _, tt := range tests {
		result := ToStringDefault(tt.input, tt.def)
		if result != tt.expected {
			t.Errorf("ToStringDefault(%v, %s) = %s, expected %s", tt.input, tt.def, result, tt.expected)
		}
	}
}

func TestParseInt(t *testing.T) {
	tests := []struct {
		s        string
		def      int
		expected int
	}{
		{"123", 0, 123},
		{"-456", 100, -456},
		{"abc", 100, 100},
		{"", 100, 100},
	}

	for _, tt := range tests {
		result := ParseInt(tt.s, tt.def)
		if result != tt.expected {
			t.Errorf("ParseInt(%q, %d) = %d, expected %d", tt.s, tt.def, result, tt.expected)
		}
	}
}

func TestToSliceString(t *testing.T) {
	tests := []struct {
		input    interface{}
		expected []string
	}{
		{[]interface{}{"a", "b", "c"}, []string{"a", "b", "c"}},
		{[]interface{}{1, 2, 3}, []string{"1", "2", "3"}},
		{"not a slice", nil},
	}

	for _, tt := range tests {
		result := ToSliceString(tt.input)
		if (result == nil) != (tt.expected == nil) {
			t.Errorf("ToSliceString(%v) = %v, expected %v", tt.input, result, tt.expected)
			continue
		}
		if len(result) != len(tt.expected) {
			t.Errorf("ToSliceString(%v) length = %d, expected %d", tt.input, len(result), len(tt.expected))
			continue
		}
		for i := range result {
			if result[i] != tt.expected[i] {
				t.Errorf("ToSliceString(%v)[%d] = %s, expected %s", tt.input, i, result[i], tt.expected[i])
			}
		}
	}
}

// 测试日志格式化输出
func TestLogFormat(t *testing.T) {
	logger := NewLogger(WithLogLevel(Debug))

	buf := logger.SetOutputString()

	logger.Info("test message")

	output := buf.String()
	if !strings.Contains(output, "INFO") {
		t.Errorf("Expected output to contain 'INFO', got: %s", output)
	}
	if !strings.Contains(output, "test message") {
		t.Errorf("Expected output to contain 'test message', got: %s", output)
	}

	// 验证时间格式
	parts := strings.Split(output, "] ")
	if len(parts) < 2 {
		t.Errorf("Unexpected log format: %s", output)
	}
	// 第一个部分应该是 [timestamp]，验证格式
	timeStr := parts[0]
	if !strings.HasPrefix(timeStr, "[") || !strings.HasSuffix(timeStr, "]") {
		t.Errorf("Time should be in brackets: %s", timeStr)
	}
}

func TestLogEntryWithEmptyContext(t *testing.T) {
	logger := NewLogger(WithLogLevel(Debug))

	buf := logger.SetOutputString()

	// 使用空 context
	ctx := context.Background()
	ctx = context.WithValue(ctx, "other_key", "value")
	logger.LogWithContext(ctx, Info, "Test message")

	output := buf.String()
	if !strings.Contains(output, "INFO") {
		t.Errorf("Expected output to contain 'INFO', got: %s", output)
	}
}

// 测试边界情况
func TestLoggerEdgeCases(t *testing.T) {
	logger := NewLogger(WithLogLevel(Error))

	buf := logger.SetOutputString()

	// 使用格式化字符串
	logger.Error("Error %d: %s", 500, "internal server error")

	output := buf.String()
	if !strings.Contains(output, "Error 500: internal server error") {
		t.Errorf("Expected formatted message in output, got: %s", output)
	}

	buf.Reset()

	// 测试空消息
	logger.Info("")
	output = buf.String()
	if !strings.Contains(output, "INFO") {
		t.Errorf("Expected output for empty message, got: %s", output)
	}
}
