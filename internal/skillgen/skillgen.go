package skillgen

import (
	"github.com/nomagicln/excel-sql-tool/internal/config"
	"github.com/nomagicln/excel-sql-tool/internal/excel"
)

// Generator defines the skill generation interface
type Generator interface {
	Generate(cfg *config.Config, meta *excel.Metadata) (string, error)
}
