package filter

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"go-log-aggregator/internal/parse"
)

type Criteria struct {
	Regex    *regexp.Regexp
	Severity string
	Since    time.Time
	Until    time.Time
	Fields   map[string]string
}

func (c Criteria) Matches(event parse.StructuredEvent) bool {
	if c.Regex != nil && !c.Regex.MatchString(event.Raw) && !c.Regex.MatchString(event.Message) {
		return false
	}
	if c.Severity != "" && !strings.EqualFold(event.Severity, c.Severity) {
		return false
	}
	if !c.Since.IsZero() && !event.Timestamp.IsZero() && event.Timestamp.Before(c.Since) {
		return false
	}
	if !c.Until.IsZero() && !event.Timestamp.IsZero() && event.Timestamp.After(c.Until) {
		return false
	}
	for key, value := range c.Fields {
		if !fieldMatches(event, key, value) {
			return false
		}
	}
	return true
}

func fieldMatches(event parse.StructuredEvent, key, value string) bool {
	switch strings.ToLower(key) {
	case "source":
		return strings.EqualFold(event.SourceName, value)
	case "format":
		return strings.EqualFold(event.Format, value)
	}

	if event.Fields == nil {
		return false
	}

	current, ok := event.Fields[key]
	if !ok {
		return false
	}

	return strings.EqualFold(current, value)
}

func ParseFieldAssignments(values []string) (map[string]string, error) {
	fields := make(map[string]string, len(values))
	for _, item := range values {
		parts := strings.SplitN(item, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("field filter must be key=value")
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if key == "" || value == "" {
			return nil, fmt.Errorf("field filter must be key=value")
		}
		fields[key] = value
	}
	return fields, nil
}
