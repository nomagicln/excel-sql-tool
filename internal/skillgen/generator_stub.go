package skillgen

import (
	"github.com/nomagicln/excel-sql-tool/internal/config"
	"github.com/nomagicln/excel-sql-tool/internal/excel"
)

type stubGenerator struct{}

// NewGenerator creates a new skill generator (stub - full implementation in feat/skillgen)
func NewGenerator() Generator {
	return &stubGenerator{}
}

func (g *stubGenerator) Generate(cfg *config.Config, meta *excel.Metadata) (string, error) {
	return "# Skill generation not yet implemented\n", nil
}
