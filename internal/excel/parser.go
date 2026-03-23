package excel

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

type excelParser struct{}

// NewParser creates a new Excel parser
func NewParser() Parser {
	return &excelParser{}
}

func (p *excelParser) Inspect(file, sheet string, headerRow int) (*SheetMeta, error) {
	f, err := excelize.OpenFile(file)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	// Check for merged cells
	merges, err := f.GetMergeCells(sheet)
	if err != nil {
		return nil, fmt.Errorf("get merge cells: %w", err)
	}
	if len(merges) > 0 {
		return nil, fmt.Errorf("sheet %q has merged cells, not supported", sheet)
	}

	rows, err := f.GetRows(sheet)
	if err != nil {
		return nil, fmt.Errorf("get rows: %w", err)
	}

	if len(rows) < headerRow {
		return nil, fmt.Errorf("sheet %q has fewer rows than header_row=%d", sheet, headerRow)
	}

	headerIdx := headerRow - 1
	headers := rows[headerIdx]

	// Infer types from first data row
	colTypes := make([]string, len(headers))
	for i := range colTypes {
		colTypes[i] = "string"
	}
	if len(rows) > headerRow {
		dataRow := rows[headerRow]
		for i, val := range dataRow {
			if i >= len(headers) {
				break
			}
			colTypes[i] = inferType(val)
		}
	}

	colsDef := make([]ColumnDef, len(headers))
	for i, h := range headers {
		colsDef[i] = ColumnDef{
			Name: h,
			Type: colTypes[i],
		}
	}

	dataRows := len(rows) - headerRow
	if dataRows < 0 {
		dataRows = 0
	}

	return &SheetMeta{
		Name:       sheet,
		Rows:       dataRows,
		Columns:    len(headers),
		HeaderRow:  headerRow,
		DataStart:  headerRow + 1,
		ColumnsDef: colsDef,
	}, nil
}

func (p *excelParser) InspectAll(file string) (*Metadata, error) {
	f, err := excelize.OpenFile(file)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	sheets := f.GetSheetList()
	meta := &Metadata{File: file, Sheets: make([]SheetMeta, 0, len(sheets))}

	for _, sheet := range sheets {
		sm, err := p.Inspect(file, sheet, 1)
		if err != nil {
			return nil, err
		}
		meta.Sheets = append(meta.Sheets, *sm)
	}
	return meta, nil
}

func (p *excelParser) ReadSheet(file, sheet string, headerRow, dataStart int) ([]map[string]interface{}, error) {
	f, err := excelize.OpenFile(file)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	rows, err := f.GetRows(sheet)
	if err != nil {
		return nil, fmt.Errorf("get rows: %w", err)
	}

	if len(rows) < headerRow {
		return nil, fmt.Errorf("no header row")
	}

	headers := rows[headerRow-1]
	result := make([]map[string]interface{}, 0)

	for i := dataStart - 1; i < len(rows); i++ {
		row := rows[i]
		record := make(map[string]interface{})
		for j, h := range headers {
			if j < len(row) {
				record[h] = parseValue(row[j])
			} else {
				record[h] = nil
			}
		}
		result = append(result, record)
	}
	return result, nil
}

func inferType(val string) string {
	if val == "" {
		return "string"
	}
	if _, err := strconv.ParseFloat(val, 64); err == nil {
		return "number"
	}
	for _, layout := range []string{"2006-01-02", "2006/01/02", "01/02/2006", "2006-01-02 15:04:05"} {
		if _, err := time.Parse(layout, strings.TrimSpace(val)); err == nil {
			return "date"
		}
	}
	return "string"
}

func parseValue(val string) interface{} {
	if val == "" {
		return nil
	}
	if f, err := strconv.ParseFloat(val, 64); err == nil {
		return f
	}
	return val
}
