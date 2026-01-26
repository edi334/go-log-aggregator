package alert

import (
	"fmt"
	"regexp"
	"strings"

	"go-log-aggregator/internal/config"
	"go-log-aggregator/internal/parse"
)

type Rule struct {
	Name       string
	Pattern    *regexp.Regexp
	Severity   string
	SourceName string
}

type Match struct {
	RuleName string
	Event    parse.StructuredEvent
}

type Evaluator struct {
	rules []Rule
}

func NewEvaluator(rules []config.AlertRule) (*Evaluator, error) {
	compiled := make([]Rule, 0, len(rules))
	for _, rule := range rules {
		if strings.TrimSpace(rule.Pattern) == "" {
			return nil, fmt.Errorf("alert %s: pattern is required", rule.Name)
		}
		re, err := regexp.Compile(rule.Pattern)
		if err != nil {
			return nil, fmt.Errorf("alert %s: invalid pattern: %w", rule.Name, err)
		}
		compiled = append(compiled, Rule{
			Name:       rule.Name,
			Pattern:    re,
			Severity:   strings.ToLower(strings.TrimSpace(rule.Severity)),
			SourceName: strings.TrimSpace(rule.SourceName),
		})
	}
	return &Evaluator{rules: compiled}, nil
}

func (e *Evaluator) Evaluate(event parse.StructuredEvent) []Match {
	if e == nil {
		return nil
	}

	matches := make([]Match, 0)
	for _, rule := range e.rules {
		if rule.Severity != "" && !strings.EqualFold(event.Severity, rule.Severity) {
			continue
		}
		if rule.SourceName != "" && !strings.EqualFold(event.SourceName, rule.SourceName) {
			continue
		}
		if !rule.Pattern.MatchString(event.Raw) && !rule.Pattern.MatchString(event.Message) {
			continue
		}
		matches = append(matches, Match{RuleName: rule.Name, Event: event})
	}

	return matches
}
