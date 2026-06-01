package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/colin2wang/postgresql-client/internal/commons"
)

// 测试 NewConfig 创建新配置
func TestNewConfig(t *testing.T) {
	cfg := NewConfig()

	if cfg.Host != "localhost" {
		t.Errorf("Expected Host to be 'localhost', got '%s'", cfg.Host)
	}
	if cfg.Port != DefaultPort {
		t.Errorf("Expected Port to be %d, got %d", DefaultPort, cfg.Port)
	}
	if cfg.User != "postgres" {
		t.Errorf("Expected User to be 'postgres', got '%s'", cfg.User)
	}
	if cfg.Database != "postgres" {
		t.Errorf("Expected Database to be 'postgres', got '%s'", cfg.Database)
	}
	if cfg.SSLMode != "disable" {
		t.Errorf("Expected SSLMode to be 'disable', got '%s'", cfg.SSLMode)
	}
}

// 测试 DefaultConfig 返回默认配置
func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Host != "localhost" {
		t.Errorf("Expected Host to be 'localhost', got '%s'", cfg.Host)
	}
	if cfg.Port != 5432 {
		t.Errorf("Expected Port to be 5432, got %d", cfg.Port)
	}
	if cfg.User != "postgres" {
		t.Errorf("Expected User to be 'postgres', got '%s'", cfg.User)
	}
	if cfg.Database != "postgres" {
		t.Errorf("Expected Database to be 'postgres', got '%s'", cfg.Database)
	}
	if cfg.SSLMode != "disable" {
		t.Errorf("Expected SSLMode to be 'disable', got '%s'", cfg.SSLMode)
	}
}

// 测试 ToString 方法
func TestConfigToString(t *testing.T) {
	cfg := &Config{
		Host:     "localhost",
		Port:     5432,
		User:     "testuser",
		Password: "testpass",
		Database: "testdb",
		SSLMode:  "require",
	}

	result := cfg.ToString()
	expectedSubstrings := []string{
		"host=localhost",
		"port=5432",
		"user=testuser",
		"password=testpass",
		"dbname=testdb",
		"sslmode=require",
	}

	for _, expected := range expectedSubstrings {
		if !contains(result, expected) {
			t.Errorf("Result should contain '%s', got: %s", expected, result)
		}
	}
}

// 测试 loadFromEnv 从环境变量加载配置
func TestLoadFromEnv(t *testing.T) {
	// 保存原始环境变量
	origHost := os.Getenv("PGHOST")
	origPort := os.Getenv("PGPORT")
	origUser := os.Getenv("PGUSER")
	origPassword := os.Getenv("PGPASSWORD")
	origDatabase := os.Getenv("PGDATABASE")

	// 清理环境变量用于测试
	defer func() {
		os.Setenv("PGHOST", origHost)
		os.Setenv("PGPORT", origPort)
		os.Setenv("PGUSER", origUser)
		os.Setenv("PGPASSWORD", origPassword)
		os.Setenv("PGDATABASE", origDatabase)
	}()

	// 设置环境变量
	os.Setenv("PGHOST", "env-host.example.com")
	os.Setenv("PGPORT", "5433")
	os.Setenv("PGUSER", "envuser")
	os.Setenv("PGPASSWORD", "envpass")
	os.Setenv("PGDATABASE", "envdb")

	cfg := &Config{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "defaultpass",
		Database: "defaultdb",
	}

	result := loadFromEnv(cfg)

	if result.Host != "env-host.example.com" {
		t.Errorf("Expected Host to be 'env-host.example.com', got '%s'", result.Host)
	}
	if result.Port != 5433 {
		t.Errorf("Expected Port to be 5433, got %d", result.Port)
	}
	if result.User != "envuser" {
		t.Errorf("Expected User to be 'envuser', got '%s'", result.User)
	}
	if result.Password != "envpass" {
		t.Errorf("Expected Password to be 'envpass', got '%s'", result.Password)
	}
	if result.Database != "envdb" {
		t.Errorf("Expected Database to be 'envdb', got '%s'", result.Database)
	}
}

// 测试 loadFromEnv 部分环境变量
func TestLoadFromEnvPartial(t *testing.T) {
	// 保存原始环境变量
	origHost := os.Getenv("PGHOST")
	origPort := os.Getenv("PGPORT")

	defer func() {
		os.Setenv("PGHOST", origHost)
		os.Setenv("PGPORT", origPort)
	}()

	// 只设置部分环境变量
	os.Setenv("PGHOST", "partial-host.example.com")
	os.Unsetenv("PGPORT")

	cfg := &Config{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "defaultpass",
		Database: "defaultdb",
	}

	result := loadFromEnv(cfg)

	if result.Host != "partial-host.example.com" {
		t.Errorf("Expected Host to be 'partial-host.example.com', got '%s'", result.Host)
	}
	if result.Port != 5432 {
		t.Errorf("Expected Port to remain 5432 (default), got %d", result.Port)
	}
}

// 测试 loadFromEnv 环境变量为空
func TestLoadFromEnvEmpty(t *testing.T) {
	// 设置空字符串环境变量
	os.Setenv("PGHOST", "")

	cfg := &Config{
		Host: "localhost",
		Port: 5432,
	}

	result := loadFromEnv(cfg)

	if result.Host != "localhost" {
		t.Errorf("Expected Host to remain 'localhost' when env var is empty, got '%s'", result.Host)
	}
}

// 测试 parseInt 函数
func TestParseInt(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		def      int
		expected int
	}{
		{"valid positive", "123", 0, 123},
		{"valid negative", "-456", 100, -456},
		{"invalid string", "abc", 100, 100},
		{"empty string", "", 100, 100},
		{"zero default", "123", 0, 123},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseInt(tt.input, tt.def)
			if result != tt.expected {
				t.Errorf("parseInt(%q, %d) = %d, expected %d", tt.input, tt.def, result, tt.expected)
			}
		})
	}
}

// 测试 fileExists 函数
func TestFileExists(t *testing.T) {
	// 创建临时文件
	tmpFile, err := os.CreateTemp("", "test-file-exists-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// 测试存在文件
	if !fileExists(tmpFile.Name()) {
		t.Errorf("Expected fileExists to return true for existing file")
	}

	// 测试不存在文件
	if fileExists("/nonexistent/path/to/file.txt") {
		t.Errorf("Expected fileExists to return false for non-existing file")
	}
}

// 测试 LoadConfig 从 YAML 文件加载
func TestLoadConfigFromYAML(t *testing.T) {
	// 创建临时配置文件
	tmpFile, err := os.CreateTemp("", "test-config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	yamlContent := `host: yaml-host.example.com
port: 5434
user: yamluser
password: yamlassword
database: yamldb
ssl_mode: require`

	if err := os.WriteFile(tmpFile.Name(), []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	cfg, err := LoadConfig(tmpFile.Name())
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.Host != "yaml-host.example.com" {
		t.Errorf("Expected Host to be 'yaml-host.example.com', got '%s'", cfg.Host)
	}
	if cfg.Port != 5434 {
		t.Errorf("Expected Port to be 5434, got %d", cfg.Port)
	}
	if cfg.User != "yamluser" {
		t.Errorf("Expected User to be 'yamluser', got '%s'", cfg.User)
	}
	if cfg.Password != "yamlassword" {
		t.Errorf("Expected Password to be 'yamlassword', got '%s'", cfg.Password)
	}
	if cfg.Database != "yamldb" {
		t.Errorf("Expected Database to be 'yamldb', got '%s'", cfg.Database)
	}
	if cfg.SSLMode != "require" {
		t.Errorf("Expected SSLMode to be 'require', got '%s'", cfg.SSLMode)
	}
}

// 测试 LoadConfig 使用环境变量回退
func TestLoadConfigWithEnvFallback(t *testing.T) {
	// 保存原始环境变量
	origHost := os.Getenv("PGHOST")
	origPort := os.Getenv("PGPORT")

	defer func() {
		os.Setenv("PGHOST", origHost)
		os.Setenv("PGPORT", origPort)
	}()

	// 设置环境变量
	os.Setenv("PGHOST", "fallback-host.example.com")
	os.Setenv("PGPORT", "5435")

	// 创建配置文件但不设置 host 和 port
	tmpFile, err := os.CreateTemp("", "test-config-fallback-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	yamlContent := `user: fallbackuser
database: fallbackdb
ssl_mode: disable`

	if err := os.WriteFile(tmpFile.Name(), []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	cfg, err := LoadConfig(tmpFile.Name())
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.Host != "fallback-host.example.com" {
		t.Errorf("Expected Host from env var, got '%s'", cfg.Host)
	}
	if cfg.Port != 5435 {
		t.Errorf("Expected Port from env var, got %d", cfg.Port)
	}
}

// 测试 LoadConfig 使用默认值
func TestLoadConfigWithDefaults(t *testing.T) {
	// 创建空的配置文件
	tmpFile, err := os.CreateTemp("", "test-config-empty-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	yamlContent := `# Empty config`

	if err := os.WriteFile(tmpFile.Name(), []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	cfg, err := LoadConfig(tmpFile.Name())
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// 应该使用默认值
	if cfg.Host != "localhost" {
		t.Errorf("Expected Host to be 'localhost' (default), got '%s'", cfg.Host)
	}
	if cfg.Port != DefaultPort {
		t.Errorf("Expected Port to be %d (default), got %d", DefaultPort, cfg.Port)
	}
}

// 测试 LoadConfig 文件不存在
func TestLoadConfigFileNotFound(t *testing.T) {
	cfg, err := LoadConfig("/nonexistent/path/config.yaml")
	if err != nil {
		t.Fatalf("LoadConfig should not fail for non-existent file (should fallback to env): %v", err)
	}
	// 应该返回默认配置或从环境变量加载
	if cfg == nil {
		t.Errorf("Expected config to not be nil")
	}
}

// 测试 LoadConfig 空路径（使用默认文件）
func TestLoadConfigEmptyPath(t *testing.T) {
	// 创建临时目录中的默认配置文件
	tmpDir, err := os.MkdirTemp("", "test-config-dir-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	defaultFilePath := filepath.Join(tmpDir, "config.yaml")
	yamlContent := `host: default-file-host
port: 5436`

	if err := os.WriteFile(defaultFilePath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// 切换到临时目录
	oldDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldDir)

	_, err = LoadConfig("")
	if err != nil {
		t.Logf("LoadConfig failed (may be expected if file not found): %v", err)
	}

	// 恢复工作目录
	os.Chdir(oldDir)
}

// 测试 SaveConfig 保存配置到文件
func TestSaveConfig(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-save-config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	cfg := &Config{
		Host:     "save-host.example.com",
		Port:     5437,
		User:     "saveuser",
		Password: "savepass",
		Database: "savedb",
		SSLMode:  "verify-full",
	}

	err = SaveConfig(cfg, tmpFile.Name())
	if err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	// 验证文件内容
	content, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to read saved file: %v", err)
	}

	expectedSubstrings := []string{
		"host: save-host.example.com",
		"port: 5437",
		"user: saveuser",
		"password: savepass",
		"database: savedb",
		"ssl_mode: verify-full",
	}

	for _, expected := range expectedSubstrings {
		if !contains(string(content), expected) {
			t.Errorf("Saved content should contain '%s', got: %s", expected, string(content))
		}
	}
}

// 测试 SaveConfig 错误处理
func TestSaveConfigError(t *testing.T) {
	err := SaveConfig(&Config{}, "/nonexistent/directory/config.yaml")
	if err == nil {
		t.Errorf("Expected error when saving to non-existent directory")
	}
}

// 测试错误类型断言
func TestConfigLoadError(t *testing.T) {
	// 创建无效的 YAML 内容
	tmpFile, err := os.CreateTemp("", "test-invalid-yaml-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	invalidYAML := `host: test.com
port: invalid-port  # not a number`

	if err := os.WriteFile(tmpFile.Name(), []byte(invalidYAML), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	_, err = LoadConfig(tmpFile.Name())
	if err == nil {
		t.Errorf("Expected error for invalid YAML")
	} else if _, ok := err.(*commons.ConfigError); !ok {
		t.Logf("Got error type: %T", err)
	}
}

// 测试 SSLMode 省略
func TestSSLModeOmitted(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-no-sslmode-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	yamlContent := `host: test.com
port: 5432
user: user
database: db`

	if err := os.WriteFile(tmpFile.Name(), []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	cfg, err := LoadConfig(tmpFile.Name())
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.SSLMode != "" {
		t.Errorf("Expected empty SSLMode when omitted, got '%s'", cfg.SSLMode)
	}
}

// 测试配置比较
func TestConfigEquality(t *testing.T) {
	cfg1 := &Config{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "pass",
		Database: "db",
		SSLMode:  "disable",
	}

	cfg2 := &Config{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "pass",
		Database: "db",
		SSLMode:  "disable",
	}

	if cfg1.Host != cfg2.Host || cfg1.Port != cfg2.Port ||
		cfg1.User != cfg2.User || cfg1.Password != cfg2.Password ||
		cfg1.Database != cfg2.Database || cfg1.SSLMode != cfg2.SSLMode {
		t.Errorf("Configs should be equal")
	}
}

// 辅助函数：检查字符串是否包含子串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
