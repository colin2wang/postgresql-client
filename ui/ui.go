package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/colin2wang/postgresql-client/commons"
)

// DBSelector database selector
type DBSelector struct {
	databases []string
	selected  string
}

// NewDBSelector creates a new database selector
func NewDBSelector() *DBSelector {
	return &DBSelector{}
}

// Select allows the user to select a database using arrow keys
func (ds *DBSelector) Select(prompt string, options []string) (string, error) {
	ds.databases = options

	if len(options) == 0 {
		return "", fmt.Errorf("no databases available")
	}

	var selected string
	err := survey.AskOne(
		&survey.Select{
			Message:  prompt,
			Options:  options,
			PageSize: 10,
		},
		&selected,
	)
	if err != nil {
		commons.DefaultLogger.Error("Failed to select database: %v", err)
		return "", err
	}

	ds.selected = selected
	return selected, nil
}

// GetSelected returns the selected database
func (ds *DBSelector) GetSelected() string {
	return ds.selected
}

// TableSelector table selector
type TableSelector struct {
	tables   []string
	selected string
}

// NewTableSelector creates a new table selector
func NewTableSelector() *TableSelector {
	return &TableSelector{}
}

// Select allows the user to select a table using arrow keys
func (ts *TableSelector) Select(prompt string, options []string) (string, error) {
	ts.tables = options

	if len(options) == 0 {
		return "", fmt.Errorf("no tables available")
	}

	var selected string
	err := survey.AskOne(
		&survey.Select{
			Message:  prompt,
			Options:  options,
			PageSize: 10,
		},
		&selected,
	)
	if err != nil {
		commons.DefaultLogger.Error("Failed to select table: %v", err)
		return "", err
	}

	ts.selected = selected
	return selected, nil
}

// GetSelected returns the selected table
func (ts *TableSelector) GetSelected() string {
	return ts.selected
}

// MultiSelector multi-select selector
type MultiSelector struct {
	items    []string
	selected []string
}

// NewMultiSelector creates a new multi-select selector
func NewMultiSelector() *MultiSelector {
	return &MultiSelector{}
}

// Select allows the user to multi-select items using arrow keys and spacebar
func (ms *MultiSelector) Select(prompt string, options []string) ([]string, error) {
	ms.items = options

	if len(options) == 0 {
		return nil, fmt.Errorf("no items available")
	}

	var selected []string
	err := survey.AskOne(
		&survey.MultiSelect{
			Message:  prompt,
			Options:  options,
			PageSize: 10,
		},
		&selected,
	)
	if err != nil {
		commons.DefaultLogger.Error("Failed to select items: %v", err)
		return nil, err
	}

	ms.selected = selected
	return selected, nil
}

// GetSelected returns the selected items
func (ms *MultiSelector) GetSelected() []string {
	return ms.selected
}

// Confirm confirmation dialog
type Confirm struct {
	message    string
	defaultVal bool
}

// NewConfirm creates a new confirmation dialog
func NewConfirm(message string, defaultVal bool) *Confirm {
	return &Confirm{
		message:    message,
		defaultVal: defaultVal,
	}
}

// Ask allows the user to confirm using arrow keys
func (c *Confirm) Ask() (bool, error) {
	var confirm bool
	err := survey.AskOne(
		&survey.Confirm{
			Message: c.message,
			Default: c.defaultVal,
		},
		&confirm,
	)
	if err != nil {
		commons.DefaultLogger.Error("Failed to confirm: %v", err)
		return false, err
	}
	return confirm, nil
}

// Input input field
type Input struct {
	message    string
	defaultVal string
	validate   survey.Validator
}

// NewInput creates a new input field
func NewInput(message, defaultVal string) *Input {
	return &Input{
		message:    message,
		defaultVal: defaultVal,
	}
}

// WithValidator adds a validator
func (i *Input) WithValidator(validate survey.Validator) *Input {
	i.validate = validate
	return i
}

// Ask allows the user to input via keyboard
func (i *Input) Ask() (string, error) {
	var input string
	var opts []survey.AskOpt
	if i.validate != nil {
		opts = append(opts, survey.WithValidator(i.validate))
	}
	err := survey.AskOne(
		&survey.Input{
			Message: i.message,
			Default: i.defaultVal,
		},
		&input,
		opts...,
	)
	if err != nil {
		commons.DefaultLogger.Error("Failed to get input: %v", err)
		return "", err
	}
	return strings.TrimSpace(input), nil
}

// Menu menu selector
type Menu struct {
	message  string
	options  []string
	selected string
}

// NewMenu creates a new menu selector
func NewMenu(message string, options []string) *Menu {
	return &Menu{
		message: message,
		options: options,
	}
}

// Select allows the user to select a menu item using arrow keys
func (m *Menu) Select() (string, error) {
	var selected string
	err := survey.AskOne(
		&survey.Select{
			Message:  m.message,
			Options:  m.options,
			PageSize: 10,
		},
		&selected,
	)
	if err != nil {
		commons.DefaultLogger.Error("Failed to select menu item: %v", err)
		return "", err
	}
	m.selected = selected
	return selected, nil
}

// GetSelected returns the selected menu item
func (m *Menu) GetSelected() string {
	return m.selected
}

// TableActionSelector table action selector
type TableActionSelector struct {
	actions  []string
	selected string
}

// NewTableActionSelector creates a new table action selector
func NewTableActionSelector() *TableActionSelector {
	return &TableActionSelector{
		actions: []string{
			"Show table structure",
			"Show table content",
		},
	}
}

// Select allows the user to select a table action
func (tas *TableActionSelector) Select(prompt string) (string, error) {
	var selected string
	err := survey.AskOne(
		&survey.Select{
			Message:  prompt,
			Options:  tas.actions,
			PageSize: 10,
		},
		&selected,
	)
	if err != nil {
		commons.DefaultLogger.Error("Failed to select table action: %v", err)
		return "", err
	}
	tas.selected = selected
	return selected, nil
}

// GetSelected returns the selected action
func (tas *TableActionSelector) GetSelected() string {
	return tas.selected
}

// RowSelector row selector
type RowSelector struct {
	rows     []map[string]interface{}
	selected int
}

// NewRowSelector creates a new row selector
func NewRowSelector(rows []map[string]interface{}) *RowSelector {
	return &RowSelector{
		rows: rows,
	}
}

// Select allows the user to select a row number
func (rs *RowSelector) Select(prompt string) (int, error) {
	if len(rs.rows) == 0 {
		return -1, fmt.Errorf("no rows available")
	}

	options := make([]string, len(rs.rows))
	for i, row := range rs.rows {
		var values []string
		for _, v := range row {
			values = append(values, formatValue(v))
		}
		options[i] = fmt.Sprintf("Row %d: %s", i+1, strings.Join(values, ", "))
	}

	var selected string
	err := survey.AskOne(
		&survey.Select{
			Message:  prompt,
			Options:  options,
			PageSize: 10,
		},
		&selected,
	)
	if err != nil {
		commons.DefaultLogger.Error("Failed to select row: %v", err)
		return -1, err
	}

	for i, option := range options {
		if option == selected {
			rs.selected = i
			return i, nil
		}
	}

	return -1, fmt.Errorf("row not found")
}

// GetSelected returns the selected row
func (rs *RowSelector) GetSelected() map[string]interface{} {
	if rs.selected >= 0 && rs.selected < len(rs.rows) {
		return rs.rows[rs.selected]
	}
	return nil
}

// RowEditor row editor
type RowEditor struct {
	originalRow map[string]interface{}
	editedRow   map[string]interface{}
	columns     []string
}

// NewRowEditor creates a new row editor
func NewRowEditor(row map[string]interface{}, columns []string) *RowEditor {
	editedRow := make(map[string]interface{})
	for k, v := range row {
		editedRow[k] = v
	}
	return &RowEditor{
		originalRow: row,
		editedRow:   editedRow,
		columns:     columns,
	}
}

// SelectColumnsToEdit selects columns to edit
func (re *RowEditor) SelectColumnsToEdit() ([]string, error) {
	options := make([]string, len(re.columns)+2)
	for i, column := range re.columns {
		currentValue := formatValue(re.editedRow[column])
		originalValue := formatValue(re.originalRow[column])
		if currentValue != originalValue {
			options[i] = fmt.Sprintf("%s: %s *", column, currentValue)
		} else {
			options[i] = fmt.Sprintf("%s: %s", column, currentValue)
		}
	}
	options[len(re.columns)] = "[Save] Save changes"
	options[len(re.columns)+1] = "[Discard] Discard changes"

	var selected string
	err := survey.AskOne(
		&survey.Select{
			Message:  "Select a column to edit or choose action:",
			Options:  options,
			PageSize: 15,
		},
		&selected,
	)
	if err != nil {
		return nil, err
	}

	if selected == "[Save] Save changes" {
		return []string{"SAVE"}, nil
	}
	if selected == "[Discard] Discard changes" {
		return []string{"DISCARD"}, nil
	}

	for _, column := range re.columns {
		if strings.HasPrefix(selected, column+":") {
			return []string{column}, nil
		}
	}

	return nil, fmt.Errorf("invalid selection")
}

// EditColumn edits a single column
func (re *RowEditor) EditColumn(column string, currentValue interface{}) (interface{}, error) {
	currentStr := formatValue(currentValue)

	var newValue string
	err := survey.AskOne(
		&survey.Input{
			Message: fmt.Sprintf("Edit %s (current: %s):", column, currentStr),
			Default: currentStr,
		},
		&newValue,
	)
	if err != nil {
		return nil, err
	}

	if newValue == "NULL" || newValue == "" {
		return nil, nil
	}
	return newValue, nil
}

// GetEditedRow returns the edited row
func (re *RowEditor) GetEditedRow() map[string]interface{} {
	return re.editedRow
}

// HasChanges checks if there are any changes
func (re *RowEditor) HasChanges() bool {
	for k, v := range re.originalRow {
		if re.editedRow[k] != v {
			return true
		}
	}
	return false
}

// Edit allows the user to edit row values
func (re *RowEditor) Edit() (map[string]interface{}, error) {
	editedRow := make(map[string]interface{})
	for k, v := range re.originalRow {
		editedRow[k] = v
	}

	for _, column := range re.columns {
		currentValue := formatValue(editedRow[column])

		var newValue string
		err := survey.AskOne(
			&survey.Input{
				Message: fmt.Sprintf("Edit %s (current: %s):", column, currentValue),
				Default: currentValue,
			},
			&newValue,
		)
		if err != nil {
			commons.DefaultLogger.Error("Failed to edit column %s: %v", column, err)
			return nil, err
		}

		if newValue == "NULL" || newValue == "" {
			editedRow[column] = nil
		} else {
			editedRow[column] = newValue
		}
	}

	return editedRow, nil
}

// formatValue formats a value for display
func formatValue(v interface{}) string {
	if v == nil {
		return "NULL"
	}
	switch val := v.(type) {
	case []byte:
		return string(val)
	default:
		return fmt.Sprintf("%v", val)
	}
}

// FormatValue formats a value for display (public)
func FormatValue(v interface{}) string {
	return formatValue(v)
}

// TruncateValue truncates a value to maxLength characters
func TruncateValue(v interface{}, maxLength int) string {
	str := formatValue(v)
	if len(str) > maxLength {
		return str[:maxLength-3] + "..."
	}
	return str
}

// PaginationSelector pagination selector
type PaginationSelector struct {
	currentPage  int
	totalPages   int
	pageSize     int
	totalRecords int
	columns      []string
	currentRows  []map[string]interface{}
}

// NewPaginationSelector creates a new pagination selector
func NewPaginationSelector(totalRecords, pageSize int) *PaginationSelector {
	totalPages := (totalRecords + pageSize - 1) / pageSize
	if totalPages < 1 {
		totalPages = 1
	}
	return &PaginationSelector{
		currentPage:  1,
		totalPages:   totalPages,
		pageSize:     pageSize,
		totalRecords: totalRecords,
	}
}

// SetCurrentPageData sets the current page data for display
func (ps *PaginationSelector) SetCurrentPageData(columns []string, rows []map[string]interface{}) {
	ps.columns = columns
	ps.currentRows = rows
}

// GetCurrentPage returns current page number
func (ps *PaginationSelector) GetCurrentPage() int {
	return ps.currentPage
}

// GetPageSize returns page size
func (ps *PaginationSelector) GetPageSize() int {
	return ps.pageSize
}

// GetTotalPages returns total pages
func (ps *PaginationSelector) GetTotalPages() int {
	return ps.totalPages
}

// GetTotalRecords returns total records
func (ps *PaginationSelector) GetTotalRecords() int {
	return ps.totalRecords
}

// SelectResult represents the result of a selection
type SelectResult struct {
	Action     string // "page", "row", or "exit"
	PageNumber int
	RowIndex   int
}

// SelectPage displays pagination options and row list, allows user to select a page or row
func (ps *PaginationSelector) SelectPage() (SelectResult, error) {
	result := SelectResult{
		Action:     "page",
		PageNumber: ps.currentPage,
		RowIndex:   -1,
	}

	if ps.totalPages <= 1 && len(ps.currentRows) == 0 {
		return result, nil
	}

	options := make([]string, 0)

	// Add backward navigation options (when not on first page)
	if ps.currentPage > 1 {
		options = append(options, "<< First Page")
		options = append(options, "< Previous Page")
	}

	// Add current page indicator
	options = append(options, fmt.Sprintf("[%d] Current Page", ps.currentPage))

	// Add forward navigation options (when not on last page)
	if ps.currentPage < ps.totalPages {
		options = append(options, "Next Page >")
		options = append(options, "Last Page >>")
	}

	// Add custom navigation options
	options = append(options, "Go to page...")
	options = append(options, "Go to row...")

	// Add back option
	options = append(options, "[Back] Back")

	// Add separator
	options = append(options, "---")

	// Add row options
	if len(ps.currentRows) > 0 {
		for i, row := range ps.currentRows {
			var values []string
			for _, col := range ps.columns {
				values = append(values, TruncateValue(row[col], 30))
			}
			rowNum := (ps.currentPage-1)*ps.pageSize + i + 1
			options = append(options, fmt.Sprintf("Row %d: %s", rowNum, strings.Join(values, ", ")))
		}
	}

	var selected string
	err := survey.AskOne(
		&survey.Select{
			Message: fmt.Sprintf("Select page or row (%d/%d, %d records per page, total %d records)",
				ps.currentPage, ps.totalPages, ps.pageSize, ps.totalRecords),
			Options:  options,
			PageSize: 20,
		},
		&selected,
	)
	if err != nil {
		return result, err
	}

	// Parse selection
	switch {
	case strings.HasPrefix(selected, "[") && strings.HasSuffix(selected, "] Current Page"):
		// Stay on current page, select a row
		result.Action = "page"
	case selected == "<< First Page":
		ps.currentPage = 1
		result.PageNumber = ps.currentPage
	case selected == "< Previous Page":
		ps.currentPage = max(1, ps.currentPage-1)
		result.PageNumber = ps.currentPage
	case selected == "Next Page >":
		ps.currentPage = min(ps.totalPages, ps.currentPage+1)
		result.PageNumber = ps.currentPage
	case selected == "Last Page >>":
		ps.currentPage = ps.totalPages
		result.PageNumber = ps.currentPage
	case selected == "Go to page...":
		var pageStr string
		err := survey.AskOne(
			&survey.Input{
				Message: fmt.Sprintf("Enter page number (1-%d):", ps.totalPages),
				Default: strconv.Itoa(ps.currentPage),
			},
			&pageStr,
		)
		if err == nil {
			if pageNum, err := strconv.Atoi(pageStr); err == nil && pageNum >= 1 && pageNum <= ps.totalPages {
				ps.currentPage = pageNum
				result.PageNumber = ps.currentPage
			}
		}
	case selected == "Go to row...":
		var rowStr string
		err := survey.AskOne(
			&survey.Input{
				Message: fmt.Sprintf("Enter row number (1-%d):", ps.totalRecords),
				Default: strconv.Itoa((ps.currentPage-1)*ps.pageSize + 1),
			},
			&rowStr,
		)
		if err == nil {
			if rowNum, err := strconv.Atoi(rowStr); err == nil && rowNum >= 1 && rowNum <= ps.totalRecords {
				ps.currentPage = (rowNum-1)/ps.pageSize + 1
				result.PageNumber = ps.currentPage
			}
		}
	case strings.HasPrefix(selected, "Row "):
		// Extract row number from selection
		rowNumStr := strings.TrimPrefix(selected, "Row ")
		rowNumStr = strings.Split(rowNumStr, ":")[0]
		if rowNum, err := strconv.Atoi(rowNumStr); err == nil {
			result.Action = "row"
			result.RowIndex = rowNum - 1 // Convert to 0-based index
		}
	case selected == "[Back] Back":
		// Exit to previous menu
		result.Action = "exit"
	}

	return result, nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// RowDetailViewer displays row details
type RowDetailViewer struct {
	row     map[string]interface{}
	columns []string
}

// NewRowDetailViewer creates a new row detail viewer
func NewRowDetailViewer(row map[string]interface{}, columns []string) *RowDetailViewer {
	return &RowDetailViewer{
		row:     row,
		columns: columns,
	}
}

// Show displays row details with edit menu
func (rdv *RowDetailViewer) Show() (string, error) {
	options := make([]string, len(rdv.columns)+3)

	// Add action options first
	options[0] = "[Edit] Edit this row"
	options[1] = "[Delete] Delete this row"
	options[2] = "[Back] Back to list"

	// Add row details with index
	for i, column := range rdv.columns {
		value := formatValue(rdv.row[column])
		options[i+3] = fmt.Sprintf("(%d) %s = %s", i+1, column, value)
	}

	var selected string
	err := survey.AskOne(
		&survey.Select{
			Message:  "Row Details",
			Options:  options,
			PageSize: 20,
		},
		&selected,
	)
	if err != nil {
		return "", err
	}

	switch selected {
	case "[Edit] Edit this row":
		return "EDIT", nil
	case "[Delete] Delete this row":
		return "DELETE", nil
	case "[Back] Back to list":
		return "BACK", nil
	default:
		return "BACK", nil
	}
}

// GetRow returns the row data
func (rdv *RowDetailViewer) GetRow() map[string]interface{} {
	return rdv.row
}
