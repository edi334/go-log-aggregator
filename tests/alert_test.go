package tests

import (
	"testing"

	"go-log-aggregator/internal/alert"
	"go-log-aggregator/internal/config"
	"go-log-aggregator/internal/parse"
)

func TestEvaluatorMatches(t *testing.T) {
	rules := []config.AlertRule{
		{Name: "panic", Pattern: "panic|fatal", Severity: "critical"},
		{Name: "api-500", Pattern: "500", SourceName: "nginx"},
	}

	eval, err := alert.NewEvaluator(rules)
	if err != nil {
		t.Fatalf("new evaluator: %v", err)
	}

	event := parse.StructuredEvent{
		SourceName: "nginx",
		Severity:   "critical",
		Message:    "panic: GET /api/items 500",
		Raw:        "panic: GET /api/items 500",
	}

	matches := eval.Evaluate(event)
	if len(matches) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(matches))
	}
}

func TestEvaluatorInvalidPattern(t *testing.T) {
	_, err := alert.NewEvaluator([]config.AlertRule{{Name: "bad", Pattern: "("}})
	if err == nil {
		t.Fatalf("expected error for invalid pattern")
	}
}
