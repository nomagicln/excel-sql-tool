# TASK_PLAN.md — excel-sql-tool

## Project Overview
A Go CLI tool that lets LLMs query Excel files via SQL without loading the full file.
Uses SQLite in-memory engine to load Excel sheets and run SQL queries against them.

## Tech Stack
- Go 1.23+
- github.com/xuri/excelize/v2 — Excel file parsing
- modernc.org/sqlite — pure Go SQLite (no CGo)
- gopkg.in/yaml.v3 — YAML config loading
- standard library flag — CLI argument parsing
- text/template — skill file generation

## Branch Strategy
Each feature is developed on its own git worktree branch, then merged to main.

| Branch               | Worktree Dir                   | Purpose                           |
|----------------------|--------------------------------|-----------------------------------|
| main                 | /home/claude/excel-sql-tool    | Integration & releases            |
| feat/core            | ../excel-sql-tool-core         | Interfaces & config types         |
| feat/excel-parser    | ../excel-sql-tool-excel        | Excel reading via excelize        |
| feat/sqlite-vtab     | ../excel-sql-tool-vtab         | SQLite engine & data loading      |
| feat/cli             | ../excel-sql-tool-cli          | CLI entry point & subcommands     |
| feat/skillgen        | ../excel-sql-tool-skillgen     | SKILL.md generation via templates |

## Task Checklist

### [x] STEP 1 — TASK_PLAN.md
- Create this file

### [ ] STEP 2 — go.mod + worktrees
- Initialize go.mod on main with `github.com/nomagicln/excel-sql-tool`
- Create git worktrees for all 5 feature branches

### [ ] STEP 3 — feat/core
- internal/config/config.go — Config struct + Load() function
- internal/excel/types.go — ColumnDef, SheetMeta, Metadata, Parser interface
- internal/engine/engine.go — QueryResult struct + Engine interface
- internal/skillgen/skillgen.go — Generator interface
- go mod tidy, go build, commit

### [ ] STEP 4 — feat/excel-parser
- Merge feat/core
- internal/excel/parser.go — Full excelize-based Parser implementation
  - Inspect(): read single sheet header, detect types, check merged cells
  - InspectAll(): iterate all sheets
  - ReadSheet(): load all rows into []map[string]interface{}
- go mod tidy, go build, commit

### [ ] STEP 5 — feat/sqlite-vtab
- Merge feat/excel-parser
- internal/vtab/vtab.go — placeholder for future virtual table support
- internal/engine/sqlite.go — SQLiteEngine implementation
  - LoadSheet(): create table + bulk insert rows in a transaction
  - Query(): run SQL and return QueryResult
  - Close(): close DB connection
- go mod tidy, go build, commit

### [ ] STEP 6 — feat/cli
- Merge feat/sqlite-vtab
- cmd/cli/main.go — Full CLI with subcommands:
  - inspect  <file> [--config] [--output]
  - generate <config.yaml> <metadata.json> [--output]
  - query    <file> "<sql>" [--sheet] [--header-row] [--data-start]
  - server   [--port] (stub)
- internal/transport/transport.go — placeholder
- go mod tidy, go build, commit

### [ ] STEP 7 — feat/skillgen
- Merge feat/core
- internal/skillgen/generator.go — Template-based SKILL.md generator
  - slugify + firstCol template functions
  - Renders config name/description/examples/domain
  - Renders schema table per sheet
  - Renders SQL examples per sheet
- go mod tidy, go build, commit

### [ ] STEP 8 — Merge all to main
- git merge feat/core feat/excel-parser feat/sqlite-vtab feat/cli feat/skillgen
- Resolve any conflicts

### [ ] STEP 9 — Final build verification
- go mod tidy
- go build ./...
- go build -o excel-sql-tool ./cmd/cli/
- ./excel-sql-tool (no args) should print usage

### [ ] STEP 10 — Smoke test
- Create test Excel file programmatically
- Run inspect, query commands
- Verify output format

## CLI Reference

```
excel-sql-tool <command> [options]

Commands:
  inspect  <excel-file> [--config config.yaml] [--output metadata.json]
  generate <config.yaml> <metadata.json> [--output SKILL.md]
  query    <excel-file> "<sql>" [--sheet name] [--header-row 1] [--data-start 2]
  server   [--port 8080]
```

## Data Flow

```
Excel File (.xlsx)
    │
    ▼
excel.Parser.ReadSheet()
    │   Reads rows via excelize
    ▼
engine.SQLiteEngine.LoadSheet()
    │   Creates SQLite table, bulk-inserts rows
    ▼
engine.SQLiteEngine.Query(sql)
    │   Runs SQL on in-memory SQLite
    ▼
engine.QueryResult{Columns, Rows}
    │
    ▼
tabwriter formatted output
```

## Config File Format (config.yaml)

```yaml
name: My Excel Dataset
description: Contains sales data for 2025
examples:
  - "Total sales by region"
  - "Top 10 products"
domain:
  - "Sales figures are in USD"
sheets:
  - name: Sales
    header_row: 1
    data_start_row: 2
  - name: Products
    header_row: 2
    data_start_row: 3
```

## Notes
- modernc.org/sqlite is pure Go, no CGo needed
- Merged cells in Excel are not supported (returns error)
- Column types are inferred from first data row: number, date, string
- Table and column names are double-quoted in SQL to handle spaces/special chars
- Virtual table (vtab) package reserved for future lazy-loading implementation
- Transport package reserved for future HTTP server mode
