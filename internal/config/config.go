package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	Sources []Source    `json:"sources"`
	Alerts  []AlertRule `json:"alerts,omitempty"`
}

type Source struct {
	Name   string `json:"name"`
	Path   string `json:"path"`
	Format string `json:"format"`
}

type AlertRule struct {
	Name       string `json:"name"`
	Pattern    string `json:"pattern"`
	Severity   string `json:"severity,omitempty"`
	SourceName string `json:"sourceName,omitempty"`
}

func Load(path string) (Config, error) {
	if strings.TrimSpace(path) == "" {
		return Config{}, fmt.Errorf("config path is required")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}

	for i, src := range cfg.Sources {
		if strings.TrimSpace(src.Name) == "" {
			return Config{}, fmt.Errorf("source[%d] name is required", i)
		}
		if strings.TrimSpace(src.Path) == "" {
			return Config{}, fmt.Errorf("source[%d] path is required", i)
		}
		if strings.TrimSpace(src.Format) == "" {
			return Config{}, fmt.Errorf("source[%d] format is required", i)
		}
	}

	return cfg, nil
}
