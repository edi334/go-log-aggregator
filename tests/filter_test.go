package tests

import (
	"regexp"
	"testing"
	"time"

	"go-log-aggregator/internal/filter"
	"go-log-aggregator/internal/parse"
)

func TestCriteriaMatches(t *testing.T) {
	event := parse.StructuredEvent{
		SourceName: "app",
		Format:     "json",
		Severity:   "error",
		Message:    "db timeout",
		Raw:        "db timeout",
		Timestamp:  time.Date(2026, 1, 26, 9, 0, 0, 0, time.UTC),
		Fields: map[string]string{
			"service": "api",
			"status":  "500",
		},
	}

	criteria := filter.Criteria{
		Regex:    regexp.MustCompile("timeout"),
		Severity: "error",
		Since:    time.Date(2026, 1, 26, 8, 0, 0, 0, time.UTC),
		Until:    time.Date(2026, 1, 26, 10, 0, 0, 0, time.UTC),
		Fields: map[string]string{
			"service": "api",
			"status":  "500",
			"source":  "app",
			"format":  "json",
		},
	}

	if !criteria.Matches(event) {
		t.Fatalf("expected event to match criteria")
	}

	criteria.Severity = "warn"
	if criteria.Matches(event) {
		t.Fatalf("expected severity mismatch")
	}
}

func TestParseFieldAssignments(t *testing.T) {
	fields, err := filter.ParseFieldAssignments([]string{"service=api", "status=500"})
	if err != nil {
		t.Fatalf("parse fields: %v", err)
	}
	if fields["service"] != "api" {
		t.Fatalf("expected service=api")
	}

	if _, err := filter.ParseFieldAssignments([]string{"badfield"}); err == nil {
		t.Fatalf("expected error for invalid field")
	}
}
