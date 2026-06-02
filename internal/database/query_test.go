package database

import (
	"context"
	"os"
	"testing"

	"github.com/colin2wang/postgresql-client/internal/commons"
)

func TestNewQuery(t *testing.T) {
	q := NewQuery(nil)
	if q == nil {
		t.Fatal("NewQuery(nil) should return a non-nil Query")
	}
	if q.DB != nil {
		t.Errorf("NewQuery(nil).DB should be nil, got %v", q.DB)
	}
}

func TestNewQuery_WithDB(t *testing.T) {
	var db *Database
	q := NewQuery(db)
	if q.DB != nil {
		t.Errorf("NewQuery(nil *Database).DB should be nil")
	}
}

func TestDatabaseInfo_Defaults(t *testing.T) {
	info := DatabaseInfo{}
	if info.Name != "" {
		t.Errorf("DatabaseInfo{}.Name should be empty, got %q", info.Name)
	}
	if info.TableCount != 0 {
		t.Errorf("DatabaseInfo{}.TableCount should be 0, got %d", info.TableCount)
	}
}

func TestTableInfo_Defaults(t *testing.T) {
	info := TableInfo{}
	if info.Name != "" {
		t.Errorf("TableInfo{}.Name should be empty, got %q", info.Name)
	}
	if info.RowCount != 0 {
		t.Errorf("TableInfo{}.RowCount should be 0, got %d", info.RowCount)
	}
}

func TestColumnInfo_Defaults(t *testing.T) {
	col := ColumnInfo{}
	if col.Name != "" {
		t.Errorf("ColumnInfo{}.Name should be empty, got %q", col.Name)
	}
	if col.IsNullable {
		t.Errorf("ColumnInfo{}.IsNullable should be false")
	}
}

func TestColumnInfo_WithValues(t *testing.T) {
	col := ColumnInfo{
		Name:         "id",
		DataType:     "integer",
		IsNullable:   false,
		DefaultValue: "nextval('id_seq')",
	}
	if col.Name != "id" {
		t.Errorf("expected name 'id', got %q", col.Name)
	}
	if col.DataType != "integer" {
		t.Errorf("expected type 'integer', got %q", col.DataType)
	}
	if col.IsNullable {
		t.Error("expected IsNullable false")
	}
	if col.DefaultValue != "nextval('id_seq')" {
		t.Errorf("expected default 'nextval(...)', got %q", col.DefaultValue)
	}
}

func TestExecuteScript_FileNotFound(t *testing.T) {
	q := NewQuery(nil)
	err := q.ExecuteScript(context.Background(), "/nonexistent/file.sql")
	if err == nil {
		t.Fatal("ExecuteScript with nonexistent file should return error")
	}
	// Verify it's a FileError
	if _, ok := err.(*commons.FileError); !ok {
		t.Errorf("expected FileError type, got %T: %v", err, err)
	}
}

func TestExecuteScript_EmptyFile(t *testing.T) {
	// Create a temp empty file
	tmpFile, err := os.CreateTemp("", "empty-*.sql")
	if err != nil {
		t.Fatal(err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	q := NewQuery(nil)
	err = q.ExecuteScript(context.Background(), tmpPath)
	if err == nil {
		t.Fatal("ExecuteScript with empty file should return error")
	}
	if err.Error() != "empty SQL file" {
		t.Errorf("expected 'empty SQL file', got %q", err.Error())
	}
}

func TestExecuteScript_WhitespaceOnlyFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "whitespace-*.sql")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = tmpFile.WriteString("  \n  \t  \n")
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	q := NewQuery(nil)
	err = q.ExecuteScript(context.Background(), tmpPath)
	if err == nil {
		t.Fatal("ExecuteScript with whitespace-only file should return error")
	}
	if err.Error() != "empty SQL file" {
		t.Errorf("expected 'empty SQL file', got %q", err.Error())
	}
}
