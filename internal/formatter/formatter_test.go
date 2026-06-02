package formatter

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestFormatValue_Nil(t *testing.T) {
	got := FormatValue(nil)
	if got != "NULL" {
		t.Errorf("FormatValue(nil) = %q, want %q", got, "NULL")
	}
}

func TestFormatValue_ByteSlice(t *testing.T) {
	got := FormatValue([]byte("hello"))
	if got != "hello" {
		t.Errorf("FormatValue([]byte) = %q, want %q", got, "hello")
	}
}

func TestFormatValue_String(t *testing.T) {
	got := FormatValue("world")
	if got != "world" {
		t.Errorf("FormatValue(string) = %q, want %q", got, "world")
	}
}

func TestFormatValue_Int(t *testing.T) {
	got := FormatValue(42)
	if got != "42" {
		t.Errorf("FormatValue(int) = %q, want %q", got, "42")
	}
}

func TestFormatValue_Float(t *testing.T) {
	got := FormatValue(3.14)
	exp := fmt.Sprintf("%v", 3.14)
	if got != exp {
		t.Errorf("FormatValue(float) = %q, want %q", got, exp)
	}
}

func TestFormatValue_Bool(t *testing.T) {
	got := FormatValue(true)
	if got != "true" {
		t.Errorf("FormatValue(bool) = %q, want %q", got, "true")
	}
}

func TestMin(t *testing.T) {
	tests := []struct {
		a, b, want int
	}{
		{1, 2, 1},
		{2, 1, 1},
		{0, 0, 0},
		{-1, 1, -1},
		{100, 50, 50},
	}
	for _, tt := range tests {
		got := min(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("min(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func captureStdout(fn func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	fn()

	w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	os.Stdout = old
	return buf.String()
}

func TestPrintTable_EmptyData(t *testing.T) {
	out := captureStdout(func() {
		PrintTable([]string{"col1", "col2"}, []map[string]interface{}{})
	})
	if !strings.Contains(out, "col1") || !strings.Contains(out, "col2") {
		t.Errorf("PrintTable with empty data should show columns, got: %s", out)
	}
}

func TestPrintTable_SingleRow(t *testing.T) {
	data := []map[string]interface{}{
		{"name": "Alice", "age": 30},
	}
	out := captureStdout(func() {
		PrintTable([]string{"name", "age"}, data)
	})
	if !strings.Contains(out, "Alice") || !strings.Contains(out, "30") {
		t.Errorf("PrintTable should render row values, got: %s", out)
	}
	if !strings.Contains(out, "name") || !strings.Contains(out, "age") {
		t.Errorf("PrintTable should render column headers, got: %s", out)
	}
}

func TestPrintTable_MultipleRows(t *testing.T) {
	data := []map[string]interface{}{
		{"id": 1, "val": "foo"},
		{"id": 2, "val": "bar"},
		{"id": 3, "val": "baz"},
	}
	out := captureStdout(func() {
		PrintTable([]string{"id", "val"}, data)
	})
	lines := strings.Split(strings.TrimSpace(out), "\n")
	// header + separator + 3 data rows = 5 lines
	if len(lines) < 5 {
		t.Errorf("PrintTable expected at least 5 lines, got %d", len(lines))
	}
	if !strings.Contains(out, "foo") || !strings.Contains(out, "bar") || !strings.Contains(out, "baz") {
		t.Errorf("PrintTable should contain all row values, got: %s", out)
	}
}

func TestPrintTable_SpecialValues(t *testing.T) {
	data := []map[string]interface{}{
		{"col": nil},
		{"col": []byte("bytes")},
		{"col": "text"},
	}
	out := captureStdout(func() {
		PrintTable([]string{"col"}, data)
	})
	if !strings.Contains(out, "NULL") {
		t.Errorf("PrintTable should show NULL for nil, got: %s", out)
	}
	if !strings.Contains(out, "bytes") {
		t.Errorf("PrintTable should show []byte content, got: %s", out)
	}
	if !strings.Contains(out, "text") {
		t.Errorf("PrintTable should show string content, got: %s", out)
	}
}

func TestPrintTableWithTruncation(t *testing.T) {
	longStr := "this is a very long string that should be truncated"
	data := []map[string]interface{}{
		{"col": longStr},
	}
	out := captureStdout(func() {
		PrintTableWithTruncation([]string{"col"}, data, 10)
	})
	if strings.Contains(out, longStr) {
		t.Errorf("PrintTableWithTruncation should truncate long values, got full string")
	}
	if !strings.Contains(out, "...") {
		t.Errorf("PrintTableWithTruncation should show ellipsis for truncated values")
	}
}

func TestPrintTableWithTruncation_ShortValue(t *testing.T) {
	shortStr := "short"
	data := []map[string]interface{}{
		{"col": shortStr},
	}
	out := captureStdout(func() {
		PrintTableWithTruncation([]string{"col"}, data, 20)
	})
	if !strings.Contains(out, shortStr) {
		t.Errorf("PrintTableWithTruncation should keep short values intact, got: %s", out)
	}
}

func TestPrintTableWithTruncation_EmptyData(t *testing.T) {
	out := captureStdout(func() {
		PrintTableWithTruncation([]string{"col"}, []map[string]interface{}{}, 10)
	})
	if !strings.Contains(out, "col") {
		t.Errorf("PrintTableWithTruncation with empty data should show columns, got: %s", out)
	}
}
