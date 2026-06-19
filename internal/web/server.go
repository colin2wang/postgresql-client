package web

import (
	"bufio"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/colin2wang/postgresql-client/internal/commons"
	"github.com/colin2wang/postgresql-client/internal/config"
	"github.com/colin2wang/postgresql-client/internal/database"
	"github.com/colin2wang/postgresql-client/internal/importer"
)

type Server struct {
	cfg      *config.Config
	db       *database.Database
	q        *database.Query
	password string
	port     int
	sessions map[string]time.Time
	mu       sync.RWMutex
}

type apiResp struct {
	Success   bool                     `json:"success"`
	Data      interface{}              `json:"data,omitempty"`
	Error     string                   `json:"error,omitempty"`
	Columns   []string                 `json:"columns,omitempty"`
	Rows      []map[string]interface{} `json:"rows,omitempty"`
	Total     int                      `json:"total,omitempty"`
	Page      int                      `json:"page,omitempty"`
	PageSize  int                      `json:"pageSize,omitempty"`
	TotalPage int                      `json:"totalPage,omitempty"`
}

func StartServer(cfg *config.Config, port int, password string) error {
	db, err := database.NewDatabase(cfg)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	s := &Server{
		cfg:      cfg,
		db:       db,
		q:        database.NewQuery(db),
		password: password,
		port:     port,
		sessions: make(map[string]time.Time),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/api/login", s.handleAPILogin)
	mux.HandleFunc("/api/tables", s.handleTables)
	mux.HandleFunc("/api/table/", s.handleTableData)
	mux.HandleFunc("/api/describe/", s.handleDescribe)
	mux.HandleFunc("/api/row/", s.handleRow)
	mux.HandleFunc("/api/import-csv", s.handleImportCSV)

	fmt.Printf("Starting web server on port %d...\n", port)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: s.wrapAuth(mux),
	}

	return server.ListenAndServe()
}

func (s *Server) wrapAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" || r.URL.Path == "/api/login" {
			next.ServeHTTP(w, r)
			return
		}

		cookie, err := r.Cookie("auth_token")
		if err != nil || cookie.Value != s.password {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "unauthorized"})
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) json(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func (s *Server) handleAPILogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.json(w, map[string]interface{}{"success": false, "error": "invalid request"})
		return
	}

	if req.Password != s.password {
		s.json(w, map[string]interface{}{"success": false, "error": "invalid password"})
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    req.Password,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   int(24 * time.Hour / time.Second),
	})

	s.json(w, map[string]interface{}{"success": true})
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, indexHTML)
}

func (s *Server) handleTables(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	tables, err := s.q.GetAllTables(ctx)
	if err != nil {
		s.json(w, map[string]interface{}{"success": false, "error": err.Error()})
		return
	}

	type tableJSON struct {
		Name     string `json:"name"`
		RowCount int    `json:"rowCount"`
	}

	result := make([]tableJSON, len(tables))
	for i, t := range tables {
		result[i] = tableJSON{Name: t.Name, RowCount: t.RowCount}
	}

	s.json(w, map[string]interface{}{"success": true, "data": result})
}

func (s *Server) handleTableData(w http.ResponseWriter, r *http.Request) {
	tableName := strings.TrimPrefix(r.URL.Path, "/api/table/")
	if tableName == "" {
		s.json(w, map[string]interface{}{"success": false, "error": "Table name required"})
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	sortCol := r.URL.Query().Get("sort")
	sortDir := r.URL.Query().Get("order")
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 1000 {
		pageSize = 20
	}
	if sortDir != "asc" && sortDir != "desc" {
		sortDir = "asc"
	}

	ctx := context.Background()
	offset := (page - 1) * pageSize

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s", quoteIdentifier(tableName))
	countRows, err := s.db.ExecuteQuery(ctx, countQuery)
	if err != nil {
		s.json(w, map[string]interface{}{"success": false, "error": err.Error()})
		return
	}
	var total int
	if countRows.Next() {
		countRows.Scan(&total)
	}
	countRows.Close()

	orderBy := ""
	if sortCol != "" {
		orderBy = fmt.Sprintf(" ORDER BY %s %s", quoteIdentifier(sortCol), sortDir)
	}
	query := fmt.Sprintf("SELECT * FROM %s%s LIMIT %d OFFSET %d", quoteIdentifier(tableName), orderBy, pageSize, offset)
	rows, err := s.db.ExecuteQuery(ctx, query)
	if err != nil {
		s.json(w, map[string]interface{}{"success": false, "error": err.Error()})
		return
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		s.json(w, map[string]interface{}{"success": false, "error": err.Error()})
		return
	}

	data, err := s.q.ParseTableData(rows, columns)
	if err != nil {
		s.json(w, map[string]interface{}{"success": false, "error": err.Error()})
		return
	}

	s.json(w, map[string]interface{}{
		"success":   true,
		"columns":   columns,
		"rows":      data,
		"total":     total,
		"page":      page,
		"pageSize":  pageSize,
		"totalPage": (total + pageSize - 1) / pageSize,
	})
}

func (s *Server) handleDescribe(w http.ResponseWriter, r *http.Request) {
	tableName := strings.TrimPrefix(r.URL.Path, "/api/describe/")
	if tableName == "" {
		s.json(w, map[string]interface{}{"success": false, "error": "Table name required"})
		return
	}

	ctx := context.Background()
	columns, err := s.q.DescribeTable(ctx, tableName)
	if err != nil {
		s.json(w, map[string]interface{}{"success": false, "error": err.Error()})
		return
	}

	type columnJSON struct {
		Name         string `json:"name"`
		DataType     string `json:"dataType"`
		IsNullable   bool   `json:"isNullable"`
		DefaultValue string `json:"defaultValue"`
	}

	result := make([]columnJSON, len(columns))
	for i, c := range columns {
		result[i] = columnJSON{
			Name:         c.Name,
			DataType:     c.DataType,
			IsNullable:   c.IsNullable,
			DefaultValue: c.DefaultValue,
		}
	}

	s.json(w, map[string]interface{}{"success": true, "data": result})
}

func (s *Server) handleRow(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/row/")
	parts := strings.SplitN(path, "/", 3)
	if len(parts) < 2 {
		s.json(w, map[string]interface{}{"success": false, "error": "Invalid path"})
		return
	}

	tableName := parts[0]
	action := parts[1]

	switch action {
	case "insert":
		s.handleInsertRow(w, r, tableName)
	case "update":
		s.handleUpdateRow(w, r, tableName)
	case "delete":
		s.handleDeleteRow(w, r, tableName)
	default:
		s.json(w, map[string]interface{}{"success": false, "error": "Unknown action"})
	}
}

func (s *Server) handleInsertRow(w http.ResponseWriter, r *http.Request, tableName string) {
	if r.Method != http.MethodPost {
		s.json(w, map[string]interface{}{"success": false, "error": "Method not allowed"})
		return
	}

	var row map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&row); err != nil {
		s.json(w, map[string]interface{}{"success": false, "error": "Invalid JSON"})
		return
	}

	ctx := context.Background()
	columns, err := s.q.DescribeTable(ctx, tableName)
	if err != nil {
		s.json(w, map[string]interface{}{"success": false, "error": err.Error()})
		return
	}

	colNames := make([]string, len(columns))
	for i, col := range columns {
		colNames[i] = col.Name
	}

	placeholders := make([]string, len(colNames))
	values := make([]interface{}, len(colNames))
	for i, col := range colNames {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		if val, ok := row[col]; ok {
			values[i] = val
		} else {
			values[i] = nil
		}
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		quoteIdentifier(tableName),
		strings.Join(quoteIdentifierList(colNames), ", "),
		strings.Join(placeholders, ", "))

	_, err = s.db.ExecuteNonQuery(ctx, query, values...)
	if err != nil {
		s.json(w, map[string]interface{}{"success": false, "error": err.Error()})
		return
	}

	s.json(w, map[string]interface{}{"success": true})
}

func (s *Server) handleUpdateRow(w http.ResponseWriter, r *http.Request, tableName string) {
	if r.Method != http.MethodPut {
		s.json(w, map[string]interface{}{"success": false, "error": "Method not allowed"})
		return
	}

	var req struct {
		OldValues map[string]interface{} `json:"oldValues"`
		NewValues map[string]interface{} `json:"newValues"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.json(w, map[string]interface{}{"success": false, "error": "Invalid JSON"})
		return
	}

	ctx := context.Background()
	columns, err := s.q.DescribeTable(ctx, tableName)
	if err != nil {
		s.json(w, map[string]interface{}{"success": false, "error": err.Error()})
		return
	}

	colNames := make([]string, len(columns))
	for i, col := range columns {
		colNames[i] = col.Name
	}

	args := make([]interface{}, 0)
	setClauses := make([]string, 0)
	for _, col := range colNames {
		if val, ok := req.NewValues[col]; ok {
			setClauses = append(setClauses, fmt.Sprintf("%s = $%d", quoteIdentifier(col), len(args)+1))
			args = append(args, val)
		}
	}

	whereClauses := make([]string, 0)
	for _, col := range colNames {
		whereClauses = append(whereClauses, fmt.Sprintf("%s = $%d", quoteIdentifier(col), len(args)+1))
		if val, ok := req.OldValues[col]; ok {
			args = append(args, val)
		} else {
			args = append(args, nil)
		}
	}

	query := fmt.Sprintf("UPDATE %s SET %s WHERE %s",
		quoteIdentifier(tableName),
		strings.Join(setClauses, ", "),
		strings.Join(whereClauses, " AND "))

	_, err = s.db.ExecuteNonQuery(ctx, query, args...)
	if err != nil {
		s.json(w, map[string]interface{}{"success": false, "error": err.Error()})
		return
	}

	s.json(w, map[string]interface{}{"success": true})
}

func (s *Server) handleDeleteRow(w http.ResponseWriter, r *http.Request, tableName string) {
	if r.Method != http.MethodDelete {
		s.json(w, map[string]interface{}{"success": false, "error": "Method not allowed"})
		return
	}

	var row map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&row); err != nil {
		s.json(w, map[string]interface{}{"success": false, "error": "Invalid JSON"})
		return
	}

	ctx := context.Background()
	columns, err := s.q.DescribeTable(ctx, tableName)
	if err != nil {
		s.json(w, map[string]interface{}{"success": false, "error": err.Error()})
		return
	}

	colNames := make([]string, len(columns))
	for i, col := range columns {
		colNames[i] = col.Name
	}

	whereClauses := make([]string, 0)
	args := make([]interface{}, 0)
	for _, col := range colNames {
		whereClauses = append(whereClauses, fmt.Sprintf("%s = $%d", quoteIdentifier(col), len(args)+1))
		if val, ok := row[col]; ok {
			args = append(args, val)
		} else {
			args = append(args, nil)
		}
	}

	query := fmt.Sprintf("DELETE FROM %s WHERE %s",
		quoteIdentifier(tableName),
		strings.Join(whereClauses, " AND "))

	_, err = s.db.ExecuteNonQuery(ctx, query, args...)
	if err != nil {
		s.json(w, map[string]interface{}{"success": false, "error": err.Error()})
		return
	}

	s.json(w, map[string]interface{}{"success": true})
}

func (s *Server) handleImportCSV(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.json(w, map[string]interface{}{"success": false, "error": "Method not allowed"})
		return
	}

	if err := r.ParseMultipartForm(64 << 20); err != nil {
		s.json(w, map[string]interface{}{"success": false, "error": "Failed to parse form: " + err.Error()})
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		s.json(w, map[string]interface{}{"success": false, "error": "No file uploaded"})
		return
	}
	defer file.Close()

	mode := r.FormValue("mode")
	tableName := r.FormValue("tableName")
	if tableName == "" {
		tableName = strings.TrimSuffix(header.Filename, ".csv")
	}

	var customHeaders []string
	if cn := r.FormValue("columnNames"); cn != "" {
		json.Unmarshal([]byte(cn), &customHeaders)
	}

	commons.DefaultLogger.Info("CSV import started: file=%s, table=%s, mode=%s, size=%d", header.Filename, tableName, mode, header.Size)

	ctx := context.Background()

	reader := csv.NewReader(bufio.NewReader(file))
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true

	headers, err := reader.Read()
	if err != nil {
		s.json(w, map[string]interface{}{"success": false, "error": "Failed to read CSV headers: " + err.Error()})
		return
	}

	if mode == "import" {
		exists, err := s.tableExists(ctx, tableName)
		if err != nil {
			s.json(w, map[string]interface{}{"success": false, "error": "Failed to check table: " + err.Error()})
			return
		}
		if !exists {
			s.json(w, map[string]interface{}{"success": false, "error": "Table '" + tableName + "' does not exist"})
			return
		}

		tableColumns, err := s.getTableColumns(ctx, tableName)
		if err != nil {
			s.json(w, map[string]interface{}{"success": false, "error": "Failed to get table columns: " + err.Error()})
			return
		}

		if len(headers) != len(tableColumns) {
			s.json(w, map[string]interface{}{"success": false, "error": fmt.Sprintf("Column count mismatch: CSV has %d columns, table has %d columns", len(headers), len(tableColumns))})
			return
		}
		for i, h := range headers {
			if !strings.EqualFold(h, tableColumns[i]) {
				s.json(w, map[string]interface{}{"success": false, "error": fmt.Sprintf("Column mismatch at position %d: CSV '%s' vs table '%s'", i+1, h, tableColumns[i])})
				return
			}
		}
	} else {
		if len(customHeaders) == len(headers) {
			headers = customHeaders
		}

		exists, err := s.tableExists(ctx, tableName)
		if err != nil {
			s.json(w, map[string]interface{}{"success": false, "error": "Failed to check table: " + err.Error()})
			return
		}

		if !exists {
			createStmt := buildCreateTableSQL(tableName, headers)
			if _, err := s.db.ExecuteNonQuery(ctx, createStmt); err != nil {
				s.json(w, map[string]interface{}{"success": false, "error": "Failed to create table: " + err.Error()})
				return
			}
			commons.DefaultLogger.Info("Auto-created table: %s", tableName)
		}
	}

	tmpFile, err := os.CreateTemp("", "csv-import-*.csv")
	if err != nil {
		s.json(w, map[string]interface{}{"success": false, "error": "Failed to create temp file"})
		return
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	file.Seek(0, 0)
	if _, err := io.Copy(tmpFile, file); err != nil {
		tmpFile.Close()
		s.json(w, map[string]interface{}{"success": false, "error": "Failed to save temp file"})
		return
	}
	tmpFile.Close()

	imp := importer.NewImporter(s.db, s.q, &s.cfg.Import)
	if err := imp.ImportCSV(ctx, tmpPath); err != nil {
		s.json(w, map[string]interface{}{"success": false, "error": err.Error()})
		return
	}

	s.json(w, map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"table":   tableName,
			"message": "CSV imported successfully",
		},
	})
}

func (s *Server) tableExists(ctx context.Context, tableName string) (bool, error) {
	query := "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name = $1"
	rows, err := s.db.ExecuteQuery(ctx, query, tableName)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	if rows.Next() {
		var count int
		if err := rows.Scan(&count); err != nil {
			return false, err
		}
		return count > 0, nil
	}
	return false, nil
}

func (s *Server) getTableColumns(ctx context.Context, tableName string) ([]string, error) {
	query := "SELECT column_name FROM information_schema.columns WHERE table_schema = 'public' AND table_name = $1 ORDER BY ordinal_position"
	rows, err := s.db.ExecuteQuery(ctx, query, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var columns []string
	for rows.Next() {
		var col string
		if err := rows.Scan(&col); err != nil {
			return nil, err
		}
		columns = append(columns, col)
	}
	return columns, nil
}

func buildCreateTableSQL(tableName string, columns []string) string {
	colDefs := make([]string, len(columns))
	for i, col := range columns {
		colDefs[i] = fmt.Sprintf("%s TEXT", quoteIdentifier(col))
	}
	return fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s)", quoteIdentifier(tableName), strings.Join(colDefs, ", "))
}

func quoteIdentifier(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}

func quoteIdentifierList(names []string) []string {
	quoted := make([]string, len(names))
	for i, name := range names {
		quoted[i] = quoteIdentifier(name)
	}
	return quoted
}
