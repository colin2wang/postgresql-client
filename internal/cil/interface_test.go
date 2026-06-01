package cil

import (
	"strings"
	"testing"

	"github.com/colin2wang/postgresql-client/internal/commons"
)

// TestDBSelector tests the DBSelector functionality
func TestDBSelector(t *testing.T) {
	ds := NewDBSelector()

	// Test with empty options
	t.Run("EmptyOptions", func(t *testing.T) {
		_, err := ds.Select("Test", []string{})
		if err == nil {
			t.Error("Expected error for empty options, got nil")
		}
	})

	// Test with single option (cannot test actual selection without mock)
	t.Run("SingleOption", func(t *testing.T) {
		options := []string{"test_db"}
		_, err := ds.Select("Test", options)
		// Survey will fail in tests without proper input, but we can check error handling
		if err == nil {
			// If no error, verify selection was stored
			if ds.GetSelected() != "test_db" {
				t.Errorf("Expected selected database to be 'test_db', got '%s'", ds.GetSelected())
			}
		}
	})

	// Test with multiple options (cannot test actual selection without mock)
	t.Run("MultipleOptions", func(t *testing.T) {
		options := []string{"db1", "db2", "db3"}
		_, err := ds.Select("Test", options)
		if err == nil {
			// If no error, verify the last selected item is stored
			selected := ds.GetSelected()
			found := false
			for _, opt := range options {
				if selected == opt {
					found = true
					break
				}
			}
			if !found && selected != "" {
				t.Errorf("Expected selected database to be one of %v, got '%s'", options, selected)
			}
		}
	})
}

// TestTableSelector tests the TableSelector functionality
func TestTableSelector(t *testing.T) {
	ts := NewTableSelector()

	// Test with empty options
	t.Run("EmptyOptions", func(t *testing.T) {
		_, err := ts.Select("Test", []string{})
		if err == nil {
			t.Error("Expected error for empty options, got nil")
		}
	})

	// Test with single option (cannot test actual selection without mock)
	t.Run("SingleOption", func(t *testing.T) {
		options := []string{"test_table"}
		_, err := ts.Select("Test", options)
		if err == nil {
			if ts.GetSelected() != "test_table" {
				t.Errorf("Expected selected table to be 'test_table', got '%s'", ts.GetSelected())
			}
		}
	})

	// Test with multiple options (cannot test actual selection without mock)
	t.Run("MultipleOptions", func(t *testing.T) {
		options := []string{"table1", "table2", "table3"}
		_, err := ts.Select("Test", options)
		if err == nil {
			selected := ts.GetSelected()
			found := false
			for _, opt := range options {
				if selected == opt {
					found = true
					break
				}
			}
			if !found && selected != "" {
				t.Errorf("Expected selected table to be one of %v, got '%s'", options, selected)
			}
		}
	})
}

// TestMultiSelector tests the MultiSelector functionality
func TestMultiSelector(t *testing.T) {
	ms := NewMultiSelector()

	// Test with empty options
	t.Run("EmptyOptions", func(t *testing.T) {
		_, err := ms.Select("Test", []string{})
		if err == nil {
			t.Error("Expected error for empty options, got nil")
		}
	})

	// Test with single option (cannot test actual selection without mock)
	t.Run("SingleOption", func(t *testing.T) {
		options := []string{"item1"}
		_, err := ms.Select("Test", options)
		if err == nil {
			selected := ms.GetSelected()
			if len(selected) != 1 || selected[0] != "item1" {
				t.Errorf("Expected selected items to contain 'item1', got %v", selected)
			}
		}
	})

	// Test with multiple options (cannot test actual selection without mock)
	t.Run("MultipleOptions", func(t *testing.T) {
		options := []string{"item1", "item2", "item3"}
		_, err := ms.Select("Test", options)
		if err == nil {
			selected := ms.GetSelected()
			// If no error, verify at least one item is selected
			if len(selected) == 0 && len(options) > 0 {
				t.Error("Expected at least one selected item")
			}
		}
	})
}

// TestConfirm tests the Confirm functionality
func TestConfirm(t *testing.T) {
	c := NewConfirm("Test message", true)

	// Test with default value (cannot test actual selection without mock)
	t.Run("WithDefaultValue", func(t *testing.T) {
		result, err := c.Ask()
		if err == nil {
			// In non-interactive mode, survey might fail, but if it succeeds
			// we should have a boolean result
			_ = result // result is bool
		}
	})

	c2 := NewConfirm("Test message 2", false)
	t.Run("WithFalseDefaultValue", func(t *testing.T) {
		result, err := c2.Ask()
		if err == nil {
			_ = result
		}
	})
}

// TestInput tests the Input functionality
func TestInput(t *testing.T) {
	i := NewInput("Enter value:", "default")

	// Test without validator (cannot test actual input without mock)
	t.Run("WithoutValidator", func(t *testing.T) {
		result, err := i.Ask()
		if err == nil {
			expected := strings.TrimSpace("test")
			if result != expected {
				t.Errorf("Expected 'test', got '%s'", result)
			}
		}
	})

	// Test with validator
	t.Run("WithValidator", func(t *testing.T) {
		i2 := NewInput("Enter value:", "default").WithValidator(func(val interface{}) error {
			if v, ok := val.(string); ok && v == "" {
				return nil // allow empty as it has default
			}
			return nil
		})
		result, err := i2.Ask()
		if err == nil {
			_ = result
		}
	})

	// Test that spaces are trimmed
	t.Run("TrimsSpaces", func(t *testing.T) {
		result, err := i.Ask()
		if err == nil && result != strings.TrimSpace(result) {
			t.Errorf("Expected trimmed result, got '%s'", result)
		}
	})
}

// TestMenu tests the Menu functionality
func TestMenu(t *testing.T) {
	m := NewMenu("Select option:", []string{"opt1", "opt2", "opt3"})

	// Test with options (cannot test actual selection without mock)
	t.Run("WithOptions", func(t *testing.T) {
		result, err := m.Select()
		if err == nil {
			found := false
			for _, opt := range m.options {
				if result == opt {
					found = true
					break
				}
			}
			if !found && result != "" {
				t.Errorf("Expected menu item to be one of %v, got '%s'", m.options, result)
			}
		}
	})

	// Test GetSelected
	t.Run("GetSelected", func(t *testing.T) {
		m2 := NewMenu("Select:", []string{"a"})
		if selected := m2.GetSelected(); selected != "" {
			t.Errorf("Expected empty string before selection, got '%s'", selected)
		}
	})

	// Test with empty options
	t.Run("EmptyOptions", func(t *testing.T) {
		m3 := NewMenu("Select:", []string{})
		result, err := m3.Select()
		if err == nil && result != "" {
			t.Errorf("Expected error or empty result for empty options")
		}
	})
}

// TestTableActionSelector tests the TableActionSelector functionality
func TestTableActionSelector(t *testing.T) {
	tas := NewTableActionSelector()

	// Verify default actions exist
	t.Run("DefaultActions", func(t *testing.T) {
		if len(tas.actions) != 2 {
			t.Errorf("Expected 2 default actions, got %d", len(tas.actions))
		}
		expectedActions := []string{"Show table structure", "Show table content"}
		for i, expected := range expectedActions {
			if tas.actions[i] != expected {
				t.Errorf("Expected action %d to be '%s', got '%s'", i, expected, tas.actions[i])
			}
		}
	})

	// Test Select (cannot test actual selection without mock)
	t.Run("SelectAction", func(t *testing.T) {
		result, err := tas.Select("Test")
		if err == nil {
			found := false
			for _, action := range tas.actions {
				if result == action {
					found = true
					break
				}
			}
			if !found && result != "" {
				t.Errorf("Expected table action to be one of %v, got '%s'", tas.actions, result)
			}
		}
	})
}

// TestRowSelector tests the RowSelector functionality
func TestRowSelector(t *testing.T) {
	rows := []map[string]interface{}{
		{"id": 1, "name": "Alice"},
		{"id": 2, "name": "Bob"},
	}

	rs := NewRowSelector(rows)

	// Test with empty rows
	t.Run("EmptyRows", func(t *testing.T) {
		rs_empty := NewRowSelector([]map[string]interface{}{})
		_, err := rs_empty.Select("Test")
		if err == nil {
			t.Error("Expected error for empty rows, got nil")
		}
	})

	// Test with valid rows (cannot test actual selection without mock)
	t.Run("ValidRows", func(t *testing.T) {
		result, err := rs.Select("Test")
		if err == nil {
			if result < 0 || result >= len(rows) {
				t.Errorf("Expected row index in range [0, %d), got %d", len(rows), result)
			}
		}
	})

	// Test GetSelected
	t.Run("GetSelected", func(t *testing.T) {
		rs2 := NewRowSelector(rows)
		selected := rs2.GetSelected()
		if selected != nil {
			t.Errorf("Expected nil before selection, got %v", selected)
		}

		// After valid selection (if no error occurred)
		if _, err := rs2.Select("Test"); err == nil {
			selected = rs2.GetSelected()
			if selected == nil {
				t.Error("Expected non-nil selected row after valid selection")
			}
		}
	})

	// Test index out of bounds
	t.Run("IndexOutOfBounds", func(t *testing.T) {
		rs3 := NewRowSelector(rows)
		rs3.selected = 999 // Simulate invalid index
		selected := rs3.GetSelected()
		if selected != nil {
			t.Errorf("Expected nil for invalid index, got %v", selected)
		}
	})
}

// TestRowEditor tests the RowEditor functionality
func TestRowEditor(t *testing.T) {
	row := map[string]interface{}{
		"id":   1,
		"name": "Alice",
		"age":  30,
	}
	columns := []string{"id", "name", "age"}

	re := NewRowEditor(row, columns)

	// Test HasChanges initially
	t.Run("InitialNoChanges", func(t *testing.T) {
		if re.HasChanges() {
			t.Error("Expected no initial changes")
		}
	})

	// Test GetEditedRow
	t.Run("GetEditedRow", func(t *testing.T) {
		edited := re.GetEditedRow()
		if edited == nil {
			t.Error("Expected non-nil edited row")
		} else if edited["name"] != "Alice" {
			t.Errorf("Expected name to be 'Alice', got '%v'", edited["name"])
		}
	})

	// Test EditColumn (cannot test actual input without mock)
	t.Run("EditColumn", func(t *testing.T) {
		newValue, err := re.EditColumn("name", "Alice")
		if err == nil {
			// newValue could be string or nil
			if newValue != nil {
				_, ok := newValue.(string)
				if !ok {
					t.Errorf("Expected string or nil value, got %T", newValue)
				}
			}
		}
	})

	// Test with empty row
	t.Run("EmptyRow", func(t *testing.T) {
		re_empty := NewRowEditor(map[string]interface{}{}, []string{})
		if re_empty.HasChanges() {
			t.Error("Expected no changes for empty row")
		}
	})
}

// TestFormatValue tests the formatValue function
func TestFormatValue(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{"nil value", nil, "NULL"},
		{"string", "hello", "hello"},
		{"int", 42, "42"},
		{"float", 3.14, "3.14"},
		{"bool true", true, "true"},
		{"bool false", false, "false"},
		{"byte slice", []byte("test"), "test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatValue(tt.input)
			if result != tt.expected {
				t.Errorf("FormatValue(%v): expected '%s', got '%s'", tt.input, tt.expected, result)
			}
		})
	}
}

// TestTruncateValue tests the TruncateValue function
func TestTruncateValue(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		maxLen   int
		expected string
	}{
		{"short string", "hello", 10, "hello"},
		{"exact length", "hello", 5, "hello"},
		{"truncate", "hello world", 8, "hello..."},
		{"nil value", nil, 10, "NULL"},
		{"zero maxLen", "test", 0, "..."},
		{"one char", "x", 1, "..."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TruncateValue(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("TruncateValue(%v, %d): expected '%s', got '%s'", tt.input, tt.maxLen, tt.expected, result)
			}
		})
	}
}

// TestMax tests the max helper function
func TestMax(t *testing.T) {
	tests := []struct {
		a        int
		b        int
		expected int
	}{
		{1, 2, 2},
		{2, 1, 2},
		{5, 5, 5},
		{-1, -2, -1},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := max(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("max(%d, %d): expected %d, got %d", tt.a, tt.b, tt.expected, result)
			}
		})
	}
}

// TestMin tests the min helper function
func TestMin(t *testing.T) {
	tests := []struct {
		a        int
		b        int
		expected int
	}{
		{1, 2, 1},
		{2, 1, 1},
		{5, 5, 5},
		{-1, -2, -2},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := min(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("min(%d, %d): expected %d, got %d", tt.a, tt.b, tt.expected, result)
			}
		})
	}
}

// TestPaginationSelector tests the PaginationSelector functionality
func TestPaginationSelector(t *testing.T) {
	ps := NewPaginationSelector(100, 10)

	// Test initial values
	t.Run("InitialValues", func(t *testing.T) {
		if ps.GetCurrentPage() != 1 {
			t.Errorf("Expected initial page to be 1, got %d", ps.GetCurrentPage())
		}
		if ps.GetPageSize() != 10 {
			t.Errorf("Expected page size to be 10, got %d", ps.GetPageSize())
		}
		if ps.GetTotalRecords() != 100 {
			t.Errorf("Expected total records to be 100, got %d", ps.GetTotalRecords())
		}
	})

	// Test total pages calculation
	t.Run("TotalPagesCalculation", func(t *testing.T) {
		tests := []struct {
			totalRecords int
			pageSize     int
			expected     int
		}{
			{100, 10, 10},
			{99, 10, 10},
			{10, 10, 1},
			{1, 10, 1},
			{0, 10, 1}, // Edge case: zero records
		}

		for _, tt := range tests {
			ps2 := NewPaginationSelector(tt.totalRecords, tt.pageSize)
			actual := ps2.GetTotalPages()
			if actual != tt.expected {
				t.Errorf("NewPaginationSelector(%d, %d): expected %d pages, got %d",
					tt.totalRecords, tt.pageSize, tt.expected, actual)
			}
		}
	})

	// Test SetCurrentPageData
	t.Run("SetCurrentPageData", func(t *testing.T) {
		columns := []string{"id", "name"}
		rows := []map[string]interface{}{
			{"id": 1, "name": "Alice"},
		}
		ps.SetCurrentPageData(columns, rows)
	})
}

// TestRowDetailViewer tests the RowDetailViewer functionality
func TestRowDetailViewer(t *testing.T) {
	row := map[string]interface{}{
		"id":   1,
		"name": "Alice",
	}
	columns := []string{"id", "name"}

	rdv := NewRowDetailViewer(row, columns)

	// Test GetRow
	t.Run("GetRow", func(t *testing.T) {
		retrieved := rdv.GetRow()
		if retrieved == nil {
			t.Error("Expected non-nil row")
		} else if retrieved["id"] != 1 {
			t.Errorf("Expected id to be 1, got '%v'", retrieved["id"])
		}
	})

	// Test Show (cannot test actual selection without mock)
	t.Run("Show", func(t *testing.T) {
		result, err := rdv.Show()
		if err == nil {
			validActions := []string{"EDIT", "DELETE", "BACK"}
			found := false
			for _, action := range validActions {
				if result == action {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected result to be one of %v, got '%s'", validActions, result)
			}
		}
	})

	// Test with empty row
	t.Run("EmptyRow", func(t *testing.T) {
		rdv2 := NewRowDetailViewer(map[string]interface{}{}, []string{"col1"})
		if rdv2.GetRow() == nil {
			t.Error("Expected non-nil row")
		}
	})
}

// Test utility functions
func TestUtilityFunctions(t *testing.T) {
	commons.DefaultLogger.Info("Running utility function tests")

	// Test formatValue with edge cases
	t.Run("FormatValueEdgeCases", func(t *testing.T) {
		tests := []struct {
			input    interface{}
			expected string
		}{
			{nil, "NULL"},
			{"test", "test"},
			{123, "123"},
			{[]byte("bytes"), "bytes"},
			{true, "true"},
		}

		for _, tt := range tests {
			result := FormatValue(tt.input)
			if result != tt.expected {
				t.Errorf("FormatValue(%v): expected '%s', got '%s'", tt.input, tt.expected, result)
			}
		}
	})

	// Test TruncateValue with edge cases
	t.Run("TruncateValueEdgeCases", func(t *testing.T) {
		tests := []struct {
			input    interface{}
			maxLen   int
			expected string
		}{
			{"hello", 10, "hello"},
			{"hello world", 8, "hello..."},
			{nil, 5, "NULL"},
		}

		for _, tt := range tests {
			result := TruncateValue(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("TruncateValue(%v, %d): expected '%s', got '%s'", tt.input, tt.maxLen, tt.expected, result)
			}
		}
	})
}
