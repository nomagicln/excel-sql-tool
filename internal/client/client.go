package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/nomagicln/excel-sql-tool/internal/engine"
)

// QueryRemote sends a SQL query to a remote excel-sql-tool server and returns the result.
func QueryRemote(serverURL, sql string) (*engine.QueryResult, error) {
	body, err := json.Marshal(map[string]string{"sql": sql})
	if err != nil {
		return nil, err
	}

	url := strings.TrimRight(serverURL, "/") + "/query"
	resp, err := http.Post(url, "application/json", bytes.NewReader(body)) //nolint:noctx
	if err != nil {
		return nil, fmt.Errorf("connect to server: %w", err)
	}
	defer resp.Body.Close()

	var raw struct {
		Columns  []string        `json:"columns"`
		Rows     [][]interface{} `json:"rows"`
		RowCount int             `json:"rowCount"`
		Error    string          `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	if raw.Error != "" {
		return nil, fmt.Errorf("server error: %s", raw.Error)
	}

	return &engine.QueryResult{
		Columns: raw.Columns,
		Rows:    raw.Rows,
	}, nil
}
