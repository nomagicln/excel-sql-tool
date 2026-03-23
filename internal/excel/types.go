package excel

// ColumnDef describes a single column
type ColumnDef struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

// SheetMeta describes a sheet's metadata
type SheetMeta struct {
	Name       string      `json:"name"`
	Rows       int         `json:"rows"`
	Columns    int         `json:"columns"`
	HeaderRow  int         `json:"header_row"`
	DataStart  int         `json:"data_start"`
	ColumnsDef []ColumnDef `json:"columns_def"`
}

// Metadata is the full metadata for an Excel file
type Metadata struct {
	File   string      `json:"file"`
	Sheets []SheetMeta `json:"sheets"`
}

// Parser defines the interface for Excel parsing
type Parser interface {
	Inspect(file, sheet string, headerRow int) (*SheetMeta, error)
	InspectAll(file string) (*Metadata, error)
	ReadSheet(file, sheet string, headerRow, dataStart int) ([]map[string]interface{}, error)
}
