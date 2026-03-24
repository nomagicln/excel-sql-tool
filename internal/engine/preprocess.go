package engine

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	reShowTables    = regexp.MustCompile(`(?i)^\s*SHOW\s+TABLES\s*;?\s*$`)
	reDesc          = regexp.MustCompile(`(?i)^\s*(?:DESC|DESCRIBE)\s+(.+?)\s*;?\s*$`)
	reShowColumns   = regexp.MustCompile(`(?i)^\s*SHOW\s+COLUMNS\s+FROM\s+(.+?)\s*;?\s*$`)
)

// PreprocessQuery intercepts MySQL-style meta-commands before they reach SQLite.
// If the query matches a known meta-command, it calls the appropriate engine
// method and returns a synthetic *QueryResult. Otherwise it returns nil, nil
// and the caller should fall through to eng.Query().
func PreprocessQuery(eng Engine, sql string) (*QueryResult, error) {
	sql = strings.TrimSpace(sql)

	if reShowTables.MatchString(sql) {
		tables, err := eng.ListTables()
		if err != nil {
			return nil, err
		}
		result := &QueryResult{
			Columns: []string{"Tables"},
			Rows:    make([][]interface{}, len(tables)),
		}
		for i, t := range tables {
			result.Rows[i] = []interface{}{t}
		}
		return result, nil
	}

	if m := reDesc.FindStringSubmatch(sql); m != nil {
		return describeResult(eng, m[1])
	}

	if m := reShowColumns.FindStringSubmatch(sql); m != nil {
		return describeResult(eng, m[1])
	}

	return nil, nil
}

func describeResult(eng Engine, tableName string) (*QueryResult, error) {
	tableName = strings.Trim(strings.TrimSpace(tableName), `"` + "`")
	cols, err := eng.DescribeTable(tableName)
	if err != nil {
		return nil, err
	}
	if len(cols) == 0 {
		return nil, fmt.Errorf("table %q not found or has no columns", tableName)
	}
	result := &QueryResult{
		Columns: []string{"Field", "Type"},
		Rows:    make([][]interface{}, len(cols)),
	}
	for i, c := range cols {
		result.Rows[i] = []interface{}{c.Name, c.Type}
	}
	return result, nil
}
