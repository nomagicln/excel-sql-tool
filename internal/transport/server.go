package transport

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/nomagicln/excel-sql-tool/internal/engine"
	"github.com/nomagicln/excel-sql-tool/internal/excel"
)

// Server holds the HTTP server state.
type Server struct {
	eng    engine.Engine
	parser excel.Parser
	mu     sync.RWMutex
	mux    *http.ServeMux
}

// NewServer creates a Server wrapping the given engine.
func NewServer(eng engine.Engine, parser excel.Parser) *Server {
	s := &Server{
		eng:    eng,
		parser: parser,
		mux:    http.NewServeMux(),
	}
	s.registerRoutes()
	return s
}

// ListenAndServe starts the HTTP server on addr (e.g. "0.0.0.0:8080").
func (s *Server) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, s.mux)
}

func (s *Server) registerRoutes() {
	s.mux.HandleFunc("GET /health", s.handleHealth)
	s.mux.HandleFunc("GET /tables", s.handleListTables)
	s.mux.HandleFunc("GET /columns/{table}", s.handleDescribeTable)
	s.mux.HandleFunc("POST /query", s.handleQuery)
	s.mux.HandleFunc("POST /load", s.handleLoad)
	s.mux.HandleFunc("DELETE /table/{name}", s.handleDropTable)
}

// writeJSON encodes v as JSON and writes it with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// GET /health
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// GET /tables
func (s *Server) handleListTables(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	tables, err := s.eng.ListTables()
	s.mu.RUnlock()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if tables == nil {
		tables = []string{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"tables": tables})
}

// GET /columns/{table}
func (s *Server) handleDescribeTable(w http.ResponseWriter, r *http.Request) {
	table := r.PathValue("table")
	s.mu.RLock()
	cols, err := s.eng.DescribeTable(table)
	s.mu.RUnlock()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	type colJSON struct {
		Name string `json:"name"`
		Type string `json:"type"`
	}
	out := make([]colJSON, len(cols))
	for i, c := range cols {
		out[i] = colJSON{Name: c.Name, Type: c.Type}
	}
	writeJSON(w, http.StatusOK, map[string]any{"columns": out})
}

// POST /query   body: {"sql": "..."}
func (s *Server) handleQuery(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SQL string `json:"sql"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if strings.TrimSpace(req.SQL) == "" {
		writeError(w, http.StatusBadRequest, "sql field is required")
		return
	}

	s.mu.RLock()
	result, err := engine.PreprocessQuery(s.eng, req.SQL)
	if err == nil && result == nil {
		result, err = s.eng.Query(req.SQL)
	}
	s.mu.RUnlock()

	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Serialize rows as [][]any (nil → null preserved)
	rows := make([][]any, len(result.Rows))
	for i, r := range result.Rows {
		rows[i] = make([]any, len(r))
		copy(rows[i], r)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"columns":  result.Columns,
		"rows":     rows,
		"rowCount": len(result.Rows),
	})
}

// POST /load   body: {"excel": "/path/to/file.xlsx"}
func (s *Server) handleLoad(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Excel string `json:"excel"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.Excel == "" {
		writeError(w, http.StatusBadRequest, "excel field is required")
		return
	}

	meta, err := s.parser.InspectAll(req.Excel)
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("inspect excel: %v", err))
		return
	}

	s.mu.Lock()
	loaded := make([]string, 0, len(meta.Sheets))
	for _, sheet := range meta.Sheets {
		if err := s.eng.LoadSheet(req.Excel, sheet.Name, sheet.HeaderRow, sheet.DataStart); err != nil {
			s.mu.Unlock()
			writeError(w, http.StatusInternalServerError, fmt.Sprintf("load sheet %q: %v", sheet.Name, err))
			return
		}
		loaded = append(loaded, sheet.Name)
	}
	s.mu.Unlock()

	writeJSON(w, http.StatusOK, map[string]any{"loaded": loaded})
}

// DELETE /table/{name}
func (s *Server) handleDropTable(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	s.mu.Lock()
	err := s.eng.DropTable(name)
	s.mu.Unlock()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"deleted": true})
}
