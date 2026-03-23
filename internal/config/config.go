package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type SheetConfig struct {
	Name         string `yaml:"name"`
	HeaderRow    int    `yaml:"header_row"`
	DataStartRow int    `yaml:"data_start_row"`
}

type Config struct {
	Name        string        `yaml:"name"`
	Description string        `yaml:"description"`
	Examples    []string      `yaml:"examples"`
	Domain      []string      `yaml:"domain"`
	Sheets      []SheetConfig `yaml:"sheets"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	// defaults
	for i := range cfg.Sheets {
		if cfg.Sheets[i].HeaderRow == 0 {
			cfg.Sheets[i].HeaderRow = 1
		}
		if cfg.Sheets[i].DataStartRow == 0 {
			cfg.Sheets[i].DataStartRow = 2
		}
	}
	return &cfg, nil
}
