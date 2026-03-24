package engine

import (
	"database/sql"
	"fmt"
	"strings"

	_ "modernc.org/sqlite"

	"github.com/nomagicln/excel-sql-tool/internal/excel"
)

type sqliteEngine struct {
	db     *sql.DB
	parser excel.Parser
}

// NewSQLiteEngine creates a new SQLite-based query engine
func NewSQLiteEngine(parser excel.Parser) (Engine, error) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	return &sqliteEngine{db: db, parser: parser}, nil
}

func (e *sqliteEngine) LoadSheet(file, sheet string, headerRow, dataStart int) error {
	rows, err := e.parser.ReadSheet(file, sheet, headerRow, dataStart)
	if err != nil {
		return fmt.Errorf("read sheet: %w", err)
	}

	// Get column metadata
	meta, err := e.parser.Inspect(file, sheet, headerRow)
	if err != nil {
		return err
	}

	// Build CREATE TABLE statement
	cols := make([]string, len(meta.ColumnsDef))
	for i, c := range meta.ColumnsDef {
		switch c.Type {
		case "number":
			cols[i] = fmt.Sprintf(`"%s" REAL`, c.Name)
		default:
			cols[i] = fmt.Sprintf(`"%s" TEXT`, c.Name)
		}
	}

	createSQL := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS "%s" (%s)`, sheet, strings.Join(cols, ", "))
	if _, err := e.db.Exec(createSQL); err != nil {
		return fmt.Errorf("create table: %w", err)
	}

	if len(rows) == 0 {
		return nil
	}

	// Prepare INSERT statement
	colNames := make([]string, len(meta.ColumnsDef))
	placeholders := make([]string, len(meta.ColumnsDef))
	for i, c := range meta.ColumnsDef {
		colNames[i] = fmt.Sprintf(`"%s"`, c.Name)
		placeholders[i] = "?"
	}
	insertSQL := fmt.Sprintf(`INSERT INTO "%s" (%s) VALUES (%s)`,
		sheet,
		strings.Join(colNames, ", "),
		strings.Join(placeholders, ", "))

	// Bulk insert in a transaction
	tx, err := e.db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(insertSQL)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	for _, row := range rows {
		vals := make([]interface{}, len(meta.ColumnsDef))
		for i, c := range meta.ColumnsDef {
			vals[i] = row[c.Name]
		}
		if _, err := stmt.Exec(vals...); err != nil {
			tx.Rollback()
			return fmt.Errorf("insert row: %w", err)
		}
	}
	return tx.Commit()
}

func (e *sqliteEngine) Query(sqlStr string) (*QueryResult, error) {
	rows, err := e.db.Query(sqlStr)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var result QueryResult
	result.Columns = cols

	for rows.Next() {
		vals := make([]interface{}, len(cols))
		ptrs := make([]interface{}, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, err
		}
		result.Rows = append(result.Rows, vals)
	}
	return &result, rows.Err()
}

func (e *sqliteEngine) ListTables() ([]string, error) {
	rows, err := e.db.Query(`SELECT name FROM sqlite_master WHERE type='table' ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("list tables: %w", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		tables = append(tables, name)
	}
	return tables, rows.Err()
}

func (e *sqliteEngine) DescribeTable(name string) ([]ColumnInfo, error) {
	rows, err := e.db.Query(fmt.Sprintf(`PRAGMA table_info("%s")`, name))
	if err != nil {
		return nil, fmt.Errorf("describe table: %w", err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var result []ColumnInfo
	for rows.Next() {
		vals := make([]interface{}, len(cols))
		ptrs := make([]interface{}, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, err
		}
		// PRAGMA table_info columns: cid, name, type, notnull, dflt_value, pk
		colName := fmt.Sprintf("%s", vals[1])
		colType := fmt.Sprintf("%s", vals[2])
		result = append(result, ColumnInfo{Name: colName, Type: colType})
	}
	return result, rows.Err()
}

func (e *sqliteEngine) DropTable(name string) error {
	_, err := e.db.Exec(fmt.Sprintf(`DROP TABLE IF EXISTS "%s"`, name))
	return err
}

func (e *sqliteEngine) Close() error {
	return e.db.Close()
}
