package skillgen

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/nomagicln/excel-sql-tool/internal/config"
	"github.com/nomagicln/excel-sql-tool/internal/excel"
)

const skillTemplate = `---
name: excel-query-{{ slugify .Config.Name }}
description: {{ .Config.Description }}
---

# {{ .Config.Name }}

## 支持的查询类型
{{ range .Config.Examples }}
- {{ . }}
{{ end }}
## 数据 Schema
{{ range .Meta.Sheets }}
### Sheet: {{ .Name }}
| 列名 | 类型 | 说明 |
|------|------|------|
{{ range .ColumnsDef }}| {{ .Name }} | {{ .Type }} | {{ .Description }} |
{{ end }}{{ end }}
## Meta Commands

The following meta commands are available in addition to standard SQL:

| Command | Description |
|---|---|
` + "| `SHOW TABLES;` | List all loaded sheets |\n" +
	"| `DESC <table>;` | Show column names and types for a table |\n" +
	"| `DESCRIBE <table>;` | Same as DESC |\n" +
	"| `SHOW COLUMNS FROM <table>;` | Same as DESC |\n" + `
> When using ` + "`excel-sql-tool query <file> --interactive`" + `, these meta commands are available in the REPL.

## 领域知识
{{ range .Config.Domain }}
- {{ . }}
{{ end }}
## SQL 示例

` + "```" + `sql
{{ range .Meta.Sheets }}SELECT * FROM "{{ .Name }}" LIMIT 100;
SELECT {{ firstCol .ColumnsDef }}, COUNT(*) FROM "{{ .Name }}" GROUP BY {{ firstCol .ColumnsDef }};
{{ end }}` + "```" + `
`

type templateData struct {
	Config *config.Config
	Meta   *excel.Metadata
}

type generator struct{}

// NewGenerator creates a new skill generator
func NewGenerator() Generator {
	return &generator{}
}

func (g *generator) Generate(cfg *config.Config, meta *excel.Metadata) (string, error) {
	funcMap := template.FuncMap{
		"slugify": func(s string) string {
			s = strings.ToLower(s)
			s = strings.ReplaceAll(s, " ", "-")
			s = strings.ReplaceAll(s, "_", "-")
			return s
		},
		"firstCol": func(cols []excel.ColumnDef) string {
			if len(cols) == 0 {
				return "*"
			}
			return fmt.Sprintf(`"%s"`, cols[0].Name)
		},
	}

	tmpl, err := template.New("skill").Funcs(funcMap).Parse(skillTemplate)
	if err != nil {
		return "", fmt.Errorf("parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, templateData{Config: cfg, Meta: meta}); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}
	return buf.String(), nil
}
