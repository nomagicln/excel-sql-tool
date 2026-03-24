package repl

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/chzyer/readline"
	"github.com/nomagicln/excel-sql-tool/internal/engine"
)

// Run starts an interactive REPL session using the given engine and prompt string.
func Run(eng engine.Engine, prompt string) error {
	historyFile := filepath.Join(os.Getenv("HOME"), ".excel_sql_history")

	rl, err := readline.NewEx(&readline.Config{
		Prompt:          prompt,
		HistoryFile:     historyFile,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	})
	if err != nil {
		return err
	}
	defer rl.Close()

	tables, err := eng.ListTables()
	if err != nil {
		fmt.Fprintf(os.Stderr, "list tables: %v\n", err)
	} else {
		fmt.Printf("Loaded tables: [%s]\n", strings.Join(tables, ", "))
	}

	for {
		line, err := rl.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break
			}
			continue
		}
		if err == io.EOF {
			break
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if line == "exit" || line == "quit" {
			break
		}

		result, err := engine.PreprocessQuery(eng, line)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			continue
		}
		if result == nil {
			result, err = eng.Query(line)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				continue
			}
		}
		printResult(result)
	}

	return nil
}

func printResult(result *engine.QueryResult) {
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
