package tests

import (
	"os"
	"path/filepath"
	"testing"

	"go-log-aggregator/internal/ingest"
)

func TestReadAvailableHandlesPartialLine(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.log")

	content := "line1\nline2\npartial"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("open file: %v", err)
	}
	defer file.Close()

	var lines []string
	partial := ""
	if err := ingest.ReadAvailableForTest(file, &partial, func(line string) {
		lines = append(lines, line)
	}); err != nil {
		t.Fatalf("read available: %v", err)
	}

	if len(lines) != 2 {
		t.Fatalf("expected 2 complete lines, got %d", len(lines))
	}
	if partial != "partial" {
		t.Fatalf("expected partial line to remain, got %q", partial)
	}
}
