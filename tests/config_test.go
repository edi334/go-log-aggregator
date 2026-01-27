package tests

import (
	"os"
	"path/filepath"
	"testing"

	"go-log-aggregator/internal/config"
)

func TestLoadConfigValid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	data := `{
  "sources": [
    {"name":"app","path":"logs/app.log","format":"json"}
  ],
  "alerts": [
    {"name":"panic","pattern":"panic","severity":"critical"}
  ]
}`
	if err := os.WriteFile(path, []byte(data), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if len(cfg.Sources) != 1 {
		t.Fatalf("expected 1 source, got %d", len(cfg.Sources))
	}
	if cfg.Sources[0].Name != "app" {
		t.Fatalf("unexpected source name: %s", cfg.Sources[0].Name)
	}
	if len(cfg.Alerts) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(cfg.Alerts))
	}
}

func TestLoadConfigInvalid(t *testing.T) {
	_, err := config.Load("")
	if err == nil {
		t.Fatalf("expected error for empty path")
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(path, []byte("{"), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if _, err := config.Load(path); err == nil {
		t.Fatalf("expected error for invalid JSON")
	}

	path = filepath.Join(dir, "missing-source.json")
	if err := os.WriteFile(path, []byte(`{"sources":[{"name":"","path":"","format":""}]}`), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if _, err := config.Load(path); err == nil {
		t.Fatalf("expected error for missing fields")
	}
}
