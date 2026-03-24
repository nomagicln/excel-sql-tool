package engine

// QueryResult holds query results
type QueryResult struct {
	Columns []string
	Rows    [][]interface{}
}

// ColumnInfo describes a single column in a table
type ColumnInfo struct {
	Name string
	Type string // TEXT | REAL
}

// Engine defines the SQL query interface
type Engine interface {
	LoadSheet(file, sheet string, headerRow, dataStart int) error
	Query(sql string) (*QueryResult, error)
	ListTables() ([]string, error)
	DescribeTable(name string) ([]ColumnInfo, error)
	DropTable(name string) error
	Close() error
}
