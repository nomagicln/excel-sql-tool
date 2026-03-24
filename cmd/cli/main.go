package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/nomagicln/excel-sql-tool/internal/config"
	"github.com/nomagicln/excel-sql-tool/internal/engine"
	"github.com/nomagicln/excel-sql-tool/internal/excel"
	"github.com/nomagicln/excel-sql-tool/internal/skillgen"
	"github.com/nomagicln/excel-sql-tool/internal/transport"
)

// splitArgs separates positional arguments from flag arguments so flags can
// appear either before or after positional args.
func splitArgs(args []string) (positional, flags []string) {
	for i := 0; i < len(args); i++ {
		a := args[i]
		if len(a) > 0 && a[0] == '-' {
			flags = append(flags, a)
			// If this flag looks like --key (not --key=val), consume next arg as value
			if len(a) > 1 && a[1] == '-' && !strings.Contains(a, "=") && i+1 < len(args) && args[i+1][0] != '-' {
				i++
				flags = append(flags, args[i])
			} else if len(a) == 2 && i+1 < len(args) && args[i+1][0] != '-' {
				// short flag like -p value
				i++
				flags = append(flags, args[i])
			}
		} else {
			positional = append(positional, a)
		}
	}
	return
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "inspect":
		cmdInspect(os.Args[2:])
	case "generate":
		cmdGenerate(os.Args[2:])
	case "query":
		cmdQuery(os.Args[2:])
	case "server":
		cmdServer(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Println("Usage: excel-sql-tool <command> [options]")
	fmt.Println("Commands:")
	fmt.Println("  inspect  <excel-file> [--config config.yaml] [--output metadata.json]")
	fmt.Println("  generate <config.yaml> <metadata.json> [--output SKILL.md]")
	fmt.Println("  query    <excel-file> \"<sql>\" [--sheet name] [--header-row 1] [--data-start 2]")
	fmt.Println("  server   [--port 8080] [--host 0.0.0.0] [--excel file.xlsx ...]")
}

func cmdInspect(args []string) {
	fs := flag.NewFlagSet("inspect", flag.ExitOnError)
	cfgPath := fs.String("config", "", "config file path")
	output := fs.String("output", "", "output metadata JSON file")
	positional, flagArgs := splitArgs(args)
	fs.Parse(flagArgs)

	if len(positional) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: inspect <excel-file>")
		os.Exit(1)
	}
	file := positional[0]

	parser := excel.NewParser()

	var meta *excel.Metadata
	var err error

	if *cfgPath != "" {
		cfg, err := config.Load(*cfgPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "load config: %v\n", err)
			os.Exit(1)
		}
		meta = &excel.Metadata{File: file, Sheets: make([]excel.SheetMeta, 0)}
		for _, sc := range cfg.Sheets {
			sm, err := parser.Inspect(file, sc.Name, sc.HeaderRow)
			if err != nil {
				fmt.Fprintf(os.Stderr, "inspect sheet %q: %v\n", sc.Name, err)
				os.Exit(1)
			}
			sm.DataStart = sc.DataStartRow
			meta.Sheets = append(meta.Sheets, *sm)
		}
	} else {
		meta, err = parser.InspectAll(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "inspect: %v\n", err)
			os.Exit(1)
		}
	}

	data, _ := json.MarshalIndent(meta, "", "  ")

	if *output != "" {
		if err := os.WriteFile(*output, data, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "write output: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Metadata written to %s\n", *output)
	} else {
		fmt.Println(string(data))
	}
}

func cmdGenerate(args []string) {
	fs := flag.NewFlagSet("generate", flag.ExitOnError)
	output := fs.String("output", "SKILL.md", "output SKILL.md file")
	positional, flagArgs := splitArgs(args)
	fs.Parse(flagArgs)

	if len(positional) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: generate <config.yaml> <metadata.json>")
		os.Exit(1)
	}

	cfg, err := config.Load(positional[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config: %v\n", err)
		os.Exit(1)
	}

	metaData, err := os.ReadFile(positional[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "read metadata: %v\n", err)
		os.Exit(1)
	}
	var meta excel.Metadata
	if err := json.Unmarshal(metaData, &meta); err != nil {
		fmt.Fprintf(os.Stderr, "parse metadata: %v\n", err)
		os.Exit(1)
	}

	gen := skillgen.NewGenerator()
	content, err := gen.Generate(cfg, &meta)
	if err != nil {
		fmt.Fprintf(os.Stderr, "generate: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(*output, []byte(content), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "write output: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("SKILL.md written to %s\n", *output)
}

func cmdQuery(args []string) {
	fs := flag.NewFlagSet("query", flag.ExitOnError)
	sheet := fs.String("sheet", "", "sheet name")
	headerRow := fs.Int("header-row", 1, "header row number")
	dataStart := fs.Int("data-start", 2, "data start row number")
	positional, flagArgs := splitArgs(args)
	fs.Parse(flagArgs)

	if len(positional) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: query <excel-file> \"<sql>\" [--sheet name]")
		os.Exit(1)
	}
	file := positional[0]
	sqlStr := positional[1]

	parser := excel.NewParser()
	eng, err := engine.NewSQLiteEngine(parser)
	if err != nil {
		fmt.Fprintf(os.Stderr, "create engine: %v\n", err)
		os.Exit(1)
	}
	defer eng.Close()

	// Determine which sheets to load
	sheetsToLoad := []string{}
	if *sheet != "" {
		sheetsToLoad = append(sheetsToLoad, *sheet)
	} else {
		// Load all sheets from the file
		meta, err := parser.InspectAll(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "inspect: %v\n", err)
			os.Exit(1)
		}
		for _, s := range meta.Sheets {
			sheetsToLoad = append(sheetsToLoad, s.Name)
		}
	}

	for _, s := range sheetsToLoad {
		if err := eng.LoadSheet(file, s, *headerRow, *dataStart); err != nil {
			fmt.Fprintf(os.Stderr, "load sheet %q: %v\n", s, err)
			os.Exit(1)
		}
	}

	result, err := engine.PreprocessQuery(eng, sqlStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "query error: %v\n", err)
		os.Exit(1)
	}
	if result == nil {
		result, err = eng.Query(sqlStr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "query error: %v\n", err)
			os.Exit(1)
		}
	}

	// Print as aligned table
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, strings.Join(result.Columns, "\t"))
	sep := make([]string, len(result.Columns))
	for i := range sep {
		sep[i] = strings.Repeat("-", 10)
	}
	fmt.Fprintln(w, strings.Join(sep, "\t"))
	for _, row := range result.Rows {
		cells := make([]string, len(row))
		for i, v := range row {
			if v == nil {
				cells[i] = "NULL"
			} else {
				cells[i] = fmt.Sprintf("%v", v)
			}
		}
		fmt.Fprintln(w, strings.Join(cells, "\t"))
	}
	w.Flush()
	fmt.Printf("\n%d rows\n", len(result.Rows))
}

// excelFlags is a flag.Value that accumulates repeated --excel flags.
type excelFlags []string

func (e *excelFlags) String() string { return strings.Join(*e, ",") }
func (e *excelFlags) Set(v string) error {
	*e = append(*e, v)
	return nil
}

func cmdServer(args []string) {
	fs := flag.NewFlagSet("server", flag.ExitOnError)
	port := fs.Int("port", 8080, "listen port")
	host := fs.String("host", "0.0.0.0", "listen host")
	var excels excelFlags
	fs.Var(&excels, "excel", "excel file to load at startup (repeatable)")
	fs.Parse(args)

	parser := excel.NewParser()
	eng, err := engine.NewSQLiteEngine(parser)
	if err != nil {
		fmt.Fprintf(os.Stderr, "create engine: %v\n", err)
		os.Exit(1)
	}
	defer eng.Close()

	// Pre-load any excel files specified via --excel
	for _, file := range excels {
		meta, err := parser.InspectAll(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "inspect %q: %v\n", file, err)
			os.Exit(1)
		}
		for _, sheet := range meta.Sheets {
			if err := eng.LoadSheet(file, sheet.Name, sheet.HeaderRow, sheet.DataStart); err != nil {
				fmt.Fprintf(os.Stderr, "load sheet %q: %v\n", sheet.Name, err)
				os.Exit(1)
			}
		}
		fmt.Printf("Loaded %q (%d sheets)\n", file, len(meta.Sheets))
	}

	addr := fmt.Sprintf("%s:%d", *host, *port)
	fmt.Printf("Listening on http://%s\n", addr)

	srv := transport.NewServer(eng, parser)
	if err := srv.ListenAndServe(addr); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}
