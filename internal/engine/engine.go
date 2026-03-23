package engine

// QueryResult holds query results
type QueryResult struct {
	Columns []string
	Rows    [][]interface{}
}

// Engine defines the SQL query interface
type Engine interface {
	LoadSheet(file, sheet string, headerRow, dataStart int) error
	Query(sql string) (*QueryResult, error)
	Close() error
}
