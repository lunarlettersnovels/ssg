package config

import (
	"fmt"
	"gopkg.in/ini.v1"
)

type Config struct {
	Database struct {
		DSN string `ini:"dsn"`
	} `ini:"database"`

	SSG struct {
		OutputDir   string `ini:"output_dir"`
		Concurrency int    `ini:"concurrency"`
		BaseURL     string `ini:"base_url"`
	} `ini:"ssg"`
}

func Load(path string) (*Config, error) {
	cfg := new(Config)
	err := ini.MapTo(cfg, path)
	if err != nil {
		return nil, fmt.Errorf("failed to map config: %w", err)
	}
	return cfg, nil
}
