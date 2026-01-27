package tests

import (
	"strings"
	"testing"
	"time"

	"go-log-aggregator/internal/ingest"
	"go-log-aggregator/internal/parse"
)

func TestParseJSON(t *testing.T) {
	line := `{"timestamp":"2026-01-26T09:00:00Z","level":"INFO","service":"api","msg":"startup","user_id":123}`
	event := ingest.Event{SourceName: "app", SourcePath: "/tmp/app.log", Line: line}

	parsed, err := parse.ParseLine("json", event)
	if err != nil {
		t.Fatalf("parse json: %v", err)
	}

	if parsed.Format != "json" {
		t.Fatalf("expected json format, got %s", parsed.Format)
	}
	if parsed.Severity != "info" {
		t.Fatalf("expected info severity, got %s", parsed.Severity)
	}
	if parsed.Message != "startup" {
		t.Fatalf("expected message startup, got %s", parsed.Message)
	}
	if parsed.Fields["service"] != "api" {
		t.Fatalf("expected service=api")
	}
	if parsed.Fields["user_id"] != "123" {
		t.Fatalf("expected user_id=123")
	}
	if parsed.Timestamp.IsZero() {
		t.Fatalf("expected timestamp parsed")
	}
}

func TestParseNginx(t *testing.T) {
	line := `10.0.0.2 - - [26/Jan/2026:09:02:00 +0000] "GET /api/items HTTP/1.1" 500 256 "-" "Go-http-client/1.1"`
	event := ingest.Event{SourceName: "nginx", SourcePath: "/tmp/nginx.log", Line: line}

	parsed, err := parse.ParseLine("nginx", event)
	if err != nil {
		t.Fatalf("parse nginx: %v", err)
	}

	if parsed.Severity != "error" {
		t.Fatalf("expected error severity, got %s", parsed.Severity)
	}
	if parsed.Fields["status"] != "500" {
		t.Fatalf("expected status 500")
	}
	if parsed.Fields["path"] != "/api/items" {
		t.Fatalf("expected path /api/items")
	}
	if !strings.Contains(parsed.Message, "GET") {
		t.Fatalf("expected message to contain method")
	}
	if parsed.Timestamp.IsZero() {
		t.Fatalf("expected timestamp parsed")
	}
}

func TestParseSyslog(t *testing.T) {
	line := "Jan 26 09:02:20 host1 myapp[4321]: panic: unexpected nil pointer"
	event := ingest.Event{SourceName: "syslog", SourcePath: "/tmp/syslog.log", Line: line}

	parsed, err := parse.ParseLine("syslog", event)
	if err != nil {
		t.Fatalf("parse syslog: %v", err)
	}

	if parsed.Severity != "critical" {
		t.Fatalf("expected critical severity, got %s", parsed.Severity)
	}
	if parsed.Fields["host"] != "host1" {
		t.Fatalf("expected host1")
	}
	if parsed.Fields["pid"] != "4321" {
		t.Fatalf("expected pid 4321")
	}
	if parsed.Timestamp.Year() != time.Now().Year() {
		t.Fatalf("expected current year in timestamp")
	}
}

func TestParseUnsupported(t *testing.T) {
	event := ingest.Event{SourceName: "unknown", SourcePath: "/tmp/unknown.log", Line: "hello"}
	if _, err := parse.ParseLine("weird", event); err == nil {
		t.Fatalf("expected error for unsupported format")
	}
}
