package parse

import (
	"fmt"
	"regexp"
	"strconv"
	"time"

	"go-log-aggregator/internal/ingest"
)

var nginxRegex = regexp.MustCompile(`^(?P<remote>\S+) \S+ \S+ \[(?P<time>[^\]]+)\] "(?P<method>\S+) (?P<path>[^"]+) (?P<proto>[^"]+)" (?P<status>\d{3}) (?P<body>\d+|-) "(?P<referer>[^"]*)" "(?P<agent>[^"]*)"`)

func parseNginx(event ingest.Event) (StructuredEvent, error) {
	matches := nginxRegex.FindStringSubmatch(event.Line)
	if matches == nil {
		return StructuredEvent{}, fmt.Errorf("nginx parse: no match")
	}

	values := make(map[string]string, len(matches))
	for i, name := range nginxRegex.SubexpNames() {
		if i == 0 || name == "" {
			continue
		}
		values[name] = matches[i]
	}

	timestamp := parseNginxTimestamp(values["time"])
	status := values["status"]
	severity := severityFromStatus(status)
	message := fmt.Sprintf("%s %s %s", values["method"], values["path"], status)

	fields := map[string]string{
		"remote_addr": values["remote"],
		"method":      values["method"],
		"path":        values["path"],
		"protocol":    values["proto"],
		"status":      status,
		"bytes":       values["body"],
		"referer":     values["referer"],
		"user_agent":  values["agent"],
	}

	return StructuredEvent{
		SourceName: event.SourceName,
		SourcePath: event.SourcePath,
		Format:     "nginx",
		Timestamp:  timestamp,
		ReceivedAt: event.ReceivedAt,
		Severity:   severity,
		Message:    message,
		Fields:     fields,
		Raw:        event.Line,
	}, nil
}

func parseNginxTimestamp(value string) time.Time {
	if value == "" {
		return time.Time{}
	}
	ts, err := time.Parse("02/Jan/2006:15:04:05 -0700", value)
	if err != nil {
		return time.Time{}
	}
	return ts
}

func severityFromStatus(value string) string {
	status, err := strconv.Atoi(value)
	if err != nil {
		return "unknown"
	}
	switch {
	case status >= 500:
		return "error"
	case status >= 400:
		return "warn"
	case status >= 300:
		return "info"
	default:
		return "info"
	}
}
