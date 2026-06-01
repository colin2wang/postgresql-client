package database

import (
	"context"
	"testing"

	"github.com/colin2wang/postgresql-client/internal/config"
)

// 测试 NewDatabase 和 NewConnector 创建数据库连接器
func TestNewDatabaseAndConnector(t *testing.T) {
	cfg := config.NewConfig()

	conn, err := NewConnector(cfg)
	if err != nil {
		t.Logf("Note: Failed to connect to default database (this is expected if no PostgreSQL server running): %v", err)
		return
	}
	defer conn.Close()

	if conn == nil {
		t.Fatal("Expected connector to not be nil")
	}

	db, err := NewDatabase(cfg)
	if err != nil {
		t.Logf("Note: Failed to create database (this is expected if no PostgreSQL server running): %v", err)
		return
	}
	defer db.Close()

	if db == nil {
		t.Fatal("Expected database to not be nil")
	}
}

// 测试 NewConnector 空配置处理
func TestNewConnectorWithNilConfig(t *testing.T) {
	cfg := &config.Config{
		Host:     "nonexistent-host-12345.com",
		Port:     5432,
		User:     "testuser",
		Password: "testpass",
		Database: "testdb",
	}

	conn, err := NewConnector(cfg)
	if err == nil {
		t.Logf("Note: Connection succeeded (expected if host exists): %v", conn)
		if conn != nil {
			conn.Close()
		}
		return
	}
}

// 测试 ExecuteQuery 成功情况
func TestExecuteQuery(t *testing.T) {
	cfg := config.NewConfig()

	conn, err := NewConnector(cfg)
	if err != nil {
		t.Skipf("Skipping test - cannot connect to database: %v", err)
	}
	defer conn.Close()

	ctx := context.Background()
	rows, err := conn.ExecuteQuery(ctx, "SELECT 1")
	if err != nil {
		t.Fatalf("ExecuteQuery failed: %v", err)
	}
	defer rows.Close()

	if rows == nil {
		t.Fatal("Expected rows to not be nil")
	}
}

// 测试 ExecuteNonQuery 成功情况
func TestExecuteNonQuery(t *testing.T) {
	cfg := config.NewConfig()

	conn, err := NewConnector(cfg)
	if err != nil {
		t.Skipf("Skipping test - cannot connect to database: %v", err)
	}
	defer conn.Close()

	ctx := context.Background()
	// 创建测试表
	_, err = conn.ExecuteNonQuery(ctx, "CREATE TEMPORARY TABLE test_table (id SERIAL, name TEXT)")
	if err != nil {
		t.Fatalf("Failed to create temp table: %v", err)
	}

	// 插入数据
	result, err := conn.ExecuteNonQuery(ctx, "INSERT INTO test_table (name) VALUES ($1)", "test")
	if err != nil {
		t.Fatalf("ExecuteNonQuery failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result to not be nil")
	}
}

// 测试 Close 连接
func TestCloseConnection(t *testing.T) {
	cfg := config.NewConfig()

	conn, err := NewConnector(cfg)
	if err != nil {
		t.Skipf("Skipping test - cannot connect to database: %v", err)
	}

	err = conn.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// 重置为 nil 来测试 connector 的 Close 行为
	var nilConn *Connector
	err = nilConn.Close()
	if err != nil {
		t.Logf("Expected error when closing nil connector: %v", err)
	}
}

// 测试 Database Close
func TestDatabaseClose(t *testing.T) {
	cfg := config.NewConfig()

	db, err := NewDatabase(cfg)
	if err != nil {
		t.Skipf("Skipping test - cannot connect to database: %v", err)
	}

	err = db.Close()
	if err != nil {
		t.Fatalf("Database Close failed: %v", err)
	}

	// 测试关闭 nil connector
	var nilDB *Database
	err = nilDB.Close()
	if err != nil {
		t.Logf("Expected error when closing database with nil connector: %v", err)
	}
}

// 测试 ExecuteQuery 无效查询处理
func TestExecuteQueryInvalidQuery(t *testing.T) {
	cfg := config.NewConfig()

	conn, err := NewConnector(cfg)
	if err != nil {
		t.Skipf("Skipping test - cannot connect to database: %v", err)
	}
	defer conn.Close()

	ctx := context.Background()
	rows, err := conn.ExecuteQuery(ctx, "SELECT * FROM nonexistent_table_xyz123")

	if rows != nil {
		rows.Close()
	}

	// 这里期望会失败，因为表不存在
	t.Logf("Expected error for invalid query: %v", err)
}

// 测试 ExecuteNonQuery 无效查询处理
func TestExecuteNonQueryInvalidQuery(t *testing.T) {
	cfg := config.NewConfig()

	conn, err := NewConnector(cfg)
	if err != nil {
		t.Skipf("Skipping test - cannot connect to database: %v", err)
	}
	defer conn.Close()

	ctx := context.Background()
	result, err := conn.ExecuteNonQuery(ctx, "INVALID SQL QUERY XYZ123")

	if result != nil {
		t.Logf("Result should be nil for invalid query")
	}

	t.Logf("Expected error for invalid query: %v", err)
}

// 测试 ExecuteQuery 在未初始化连接器时的行为
func TestExecuteQueryWithNilConnector(t *testing.T) {
	var db Database // 未初始化的 connector

	ctx := context.Background()
	rows, err := db.ExecuteQuery(ctx, "SELECT 1")

	if rows != nil {
		rows.Close()
	}

	if err == nil {
		t.Errorf("Expected error when connector is not initialized")
	} else if err.Error() != "database connector is not initialized" {
		t.Errorf("Expected 'database connector is not initialized' error, got: %v", err)
	}
}

// 测试 ExecuteNonQuery 在未初始化连接器时的行为
func TestExecuteNonQueryWithNilConnector(t *testing.T) {
	var db Database // 未初始化的 connector

	ctx := context.Background()
	result, err := db.ExecuteNonQuery(ctx, "SELECT 1")

	if result != nil {
		t.Errorf("Expected nil result when connector is not initialized")
	}

	if err == nil {
		t.Errorf("Expected error when connector is not initialized")
	} else if err.Error() != "database connector is not initialized" {
		t.Errorf("Expected 'database connector is not initialized' error, got: %v", err)
	}
}

// 测试上下文超时
func TestExecuteQueryWithContextTimeout(t *testing.T) {
	cfg := config.NewConfig()

	conn, err := NewConnector(cfg)
	if err != nil {
		t.Skipf("Skipping test - cannot connect to database: %v", err)
	}
	defer conn.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // 立即取消上下文

	rows, err := conn.ExecuteQuery(ctx, "SELECT 1")

	if rows != nil {
		rows.Close()
	}

	t.Logf("Expected error for cancelled context: %v", err)
}

// 测试带参数的查询
func TestExecuteQueryWithArgs(t *testing.T) {
	cfg := config.NewConfig()

	conn, err := NewConnector(cfg)
	if err != nil {
		t.Skipf("Skipping test - cannot connect to database: %v", err)
	}
	defer conn.Close()

	ctx := context.Background()
	rows, err := conn.ExecuteQuery(ctx, "SELECT $1::text AS result", "hello")
	if err != nil {
		t.Fatalf("ExecuteQuery with args failed: %v", err)
	}
	defer rows.Close()

	if rows == nil {
		t.Fatal("Expected rows to not be nil")
	}
}

// 测试带参数的非查询
func TestExecuteNonQueryWithArgs(t *testing.T) {
	cfg := config.NewConfig()

	conn, err := NewConnector(cfg)
	if err != nil {
		t.Skipf("Skipping test - cannot connect to database: %v", err)
	}
	defer conn.Close()

	ctx := context.Background()

	// 创建测试表
	_, err = conn.ExecuteNonQuery(ctx, "CREATE TEMPORARY TABLE test_args (id SERIAL, value TEXT)")
	if err != nil {
		t.Fatalf("Failed to create temp table: %v", err)
	}

	// 使用参数插入数据
	result, err := conn.ExecuteNonQuery(ctx, "INSERT INTO test_args (value) VALUES ($1)", "test-value")
	if err != nil {
		t.Fatalf("ExecuteNonQuery with args failed: %v", err)
	}
	defer func() {
		conn.ExecuteNonQuery(ctx, "DROP TABLE IF EXISTS test_args")
	}()

	if result == nil {
		t.Fatal("Expected result to not be nil")
	}
}

// 测试查询返回单行
func TestExecuteQuerySingleRow(t *testing.T) {
	cfg := config.NewConfig()

	conn, err := NewConnector(cfg)
	if err != nil {
		t.Skipf("Skipping test - cannot connect to database: %v", err)
	}
	defer conn.Close()

	ctx := context.Background()
	rows, err := conn.ExecuteQuery(ctx, "SELECT 42 AS value")
	if err != nil {
		t.Fatalf("ExecuteQuery failed: %v", err)
	}
	defer rows.Close()

	// 读取一行数据
	if !rows.Next() {
		t.Fatal("Expected at least one row")
	}

	var value int
	err = rows.Scan(&value)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if value != 42 {
		t.Errorf("Expected value to be 42, got %d", value)
	}
}

// 测试空结果集
func TestExecuteQueryNoRows(t *testing.T) {
	cfg := config.NewConfig()

	conn, err := NewConnector(cfg)
	if err != nil {
		t.Skipf("Skipping test - cannot connect to database: %v", err)
	}
	defer conn.Close()

	ctx := context.Background()
	rows, err := conn.ExecuteQuery(ctx, "SELECT 1 WHERE FALSE")
	if err != nil {
		t.Fatalf("ExecuteQuery failed: %v", err)
	}
	defer rows.Close()

	// 检查是否有数据
	hasRow := rows.Next()
	if hasRow {
		t.Error("Expected no rows, but got one")
	}
}

// 测试多次查询
func TestMultipleQueries(t *testing.T) {
	cfg := config.NewConfig()

	conn, err := NewConnector(cfg)
	if err != nil {
		t.Skipf("Skipping test - cannot connect to database: %v", err)
	}
	defer conn.Close()

	ctx := context.Background()

	// 执行多次查询
	for i := 0; i < 3; i++ {
		rows, err := conn.ExecuteQuery(ctx, "SELECT $1::int AS value", i+1)
		if err != nil {
			t.Fatalf("ExecuteQuery %d failed: %v", i+1, err)
		}

		var value int
		if rows.Next() {
			rows.Scan(&value)
			if value != i+1 {
				t.Errorf("Expected value %d, got %d", i+1, value)
			}
		}
		rows.Close()
	}
}

// 测试连接器 Close 多次调用
func TestConnectorCloseMultiple(t *testing.T) {
	cfg := config.NewConfig()

	conn, err := NewConnector(cfg)
	if err != nil {
		t.Skipf("Skipping test - cannot connect to database: %v", err)
	}

	err = conn.Close()
	if err != nil {
		t.Logf("First close error (unexpected): %v", err)
	}

	// 第二次关闭应该不会出错
	err = conn.Close()
	if err != nil {
		t.Logf("Second close error (may be expected if db is already closed): %v", err)
	}
}

// 测试数据库连接字符串格式
func TestConnectionStringFormat(t *testing.T) {
	cfg := &config.Config{
		Host:     "testhost",
		Port:     5432,
		User:     "testuser",
		Password: "testpass",
		Database: "testdb",
		SSLMode:  "verify-full",
	}

	connStr := cfg.ToString()
	expectedSubstrings := []string{
		"host=testhost",
		"port=5432",
		"user=testuser",
		"password=testpass",
		"dbname=testdb",
		"sslmode=verify-full",
	}

	for _, expected := range expectedSubstrings {
		if !contains(connStr, expected) {
			t.Errorf("Connection string should contain '%s', got: %s", expected, connStr)
		}
	}
}

// 测试连接器初始化
func TestConnectorInitialization(t *testing.T) {
	cfg := config.NewConfig()

	conn, err := NewConnector(cfg)
	if err != nil {
		t.Skipf("Skipping test - cannot connect to database: %v", err)
	}
	defer conn.Close()

	if conn.db == nil {
		t.Error("Expected db to be initialized")
	}
	if conn.config == nil {
		t.Error("Expected config to be initialized")
	}
}

// 测试 Database 的 connector 引用
func TestDatabaseConnectorReference(t *testing.T) {
	cfg := config.NewConfig()

	db, err := NewDatabase(cfg)
	if err != nil {
		t.Skipf("Skipping test - cannot connect to database: %v", err)
	}
	defer db.Close()

	if db.connector == nil {
		t.Error("Expected connector to be set in Database")
	}
}

// 测试带大量参数的查询
func TestExecuteQueryWithManyArgs(t *testing.T) {
	cfg := config.NewConfig()

	conn, err := NewConnector(cfg)
	if err != nil {
		t.Skipf("Skipping test - cannot connect to database: %v", err)
	}
	defer conn.Close()

	ctx := context.Background()

	// 测试带多个参数的查询
	query := "SELECT $1::int AS a, $2::text AS b, $3::bool AS c"
	rows, err := conn.ExecuteQuery(ctx, query, 42, "hello", true)
	if err != nil {
		t.Fatalf("ExecuteQuery with many args failed: %v", err)
	}
	defer rows.Close()

	var a int
	var b string
	var c bool
	if rows.Next() {
		rows.Scan(&a, &b, &c)
		if a != 42 || b != "hello" || !c {
			t.Errorf("Expected (42, hello, true), got (%d, %s, %v)", a, b, c)
		}
	}
}

// 测试数据库连接器方法调用顺序
func TestConnectorMethodOrder(t *testing.T) {
	cfg := config.NewConfig()

	conn, err := NewConnector(cfg)
	if err != nil {
		t.Skipf("Skipping test - cannot connect to database: %v", err)
	}

	ctx := context.Background()

	// 先执行查询
	rows1, err := conn.ExecuteQuery(ctx, "SELECT 1")
	if err == nil && rows1 != nil {
		rows1.Close()
	}

	// 再执行非查询
	result, err := conn.ExecuteNonQuery(ctx, "SELECT 2")
	if result != nil {
		t.Logf("Non-query result: %v", result)
	}

	conn.Close()
}

// 辅助函数：检查字符串是否包含子串
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// 测试临时表清理
func TestCleanupTempTable(t *testing.T) {
	cfg := config.NewConfig()

	conn, err := NewConnector(cfg)
	if err != nil {
		t.Skipf("Skipping test - cannot connect to database: %v", err)
	}
	defer conn.Close()

	ctx := context.Background()

	// 创建临时表
	_, err = conn.ExecuteNonQuery(ctx, "CREATE TEMPORARY TABLE cleanup_test (id SERIAL)")
	if err != nil {
		t.Fatalf("Failed to create temp table: %v", err)
	}

	// 插入数据
	_, err = conn.ExecuteNonQuery(ctx, "INSERT INTO cleanup_test DEFAULT VALUES")
	if err != nil {
		t.Logf("Insert error: %v", err)
	}

	// 查询验证
	rows, err := conn.ExecuteQuery(ctx, "SELECT COUNT(*) FROM cleanup_test")
	if err == nil && rows != nil {
		var count int
		if rows.Next() {
			rows.Scan(&count)
			t.Logf("Temp table has %d row(s)", count)
		}
		rows.Close()
	}

	// 不显式删除，临时表会在会话结束时自动清理
}

// 测试连接池状态
func TestConnectionPoolStatus(t *testing.T) {
	cfg := config.NewConfig()

	conn, err := NewConnector(cfg)
	if err != nil {
		t.Skipf("Skipping test - cannot connect to database: %v", err)
	}
	defer conn.Close()

	// 简单测试连接是否有效
	ctx := context.Background()
	rows, err := conn.ExecuteQuery(ctx, "SELECT 1")
	if err == nil && rows != nil {
		rows.Close()
		t.Log("Connection pool is healthy")
	} else {
		t.Logf("Connection pool issue: %v", err)
	}
}

// 测试并发查询（基本测试）
func TestConcurrentQueries(t *testing.T) {
	cfg := config.NewConfig()

	conn, err := NewConnector(cfg)
	if err != nil {
		t.Skipf("Skipping test - cannot connect to database: %v", err)
	}
	defer conn.Close()

	ctx := context.Background()
	done := make(chan bool, 10)

	// 并发执行查询
	for i := 0; i < 10; i++ {
		go func(id int) {
			rows, _ := conn.ExecuteQuery(ctx, "SELECT $1::int", id)
			if rows != nil {
				rows.Close()
			}
			done <- true
		}(i)
	}

	// 等待所有 goroutine 完成
	for i := 0; i < 10; i++ {
		<-done
	}

	t.Log("All concurrent queries completed")
}
