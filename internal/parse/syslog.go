package parse

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"go-log-aggregator/internal/ingest"
)

var syslogRegex = regexp.MustCompile(`^(?P<month>[A-Z][a-z]{2})\s+(?P<day>\d{1,2})\s+(?P<time>\d{2}:\d{2}:\d{2})\s+(?P<host>\S+)\s+(?P<tag>[^:]+):\s*(?P<msg>.*)$`)

func parseSyslog(event ingest.Event) (StructuredEvent, error) {
	matches := syslogRegex.FindStringSubmatch(event.Line)
	if matches == nil {
		return StructuredEvent{}, fmt.Errorf("syslog parse: no match")
	}

	values := make(map[string]string, len(matches))
	for i, name := range syslogRegex.SubexpNames() {
		if i == 0 || name == "" {
			continue
		}
		values[name] = matches[i]
	}

	timestamp := parseSyslogTimestamp(values["month"], values["day"], values["time"])
	message := values["msg"]
	severity := severityFromSyslog(message)

	fields := map[string]string{
		"host": values["host"],
		"tag":  values["tag"],
	}

	if pid := extractPID(values["tag"]); pid != "" {
		fields["pid"] = pid
	}

	return StructuredEvent{
		SourceName: event.SourceName,
		SourcePath: event.SourcePath,
		Format:     "syslog",
		Timestamp:  timestamp,
		Severity:   severity,
		Message:    message,
		Fields:     fields,
		Raw:        event.Line,
	}, nil
}

func parseSyslogTimestamp(month, day, clock string) time.Time {
	year := time.Now().Year()
	value := fmt.Sprintf("%s %s %s %d", month, day, clock, year)
	ts, err := time.Parse("Jan 2 15:04:05 2006", value)
	if err != nil {
		return time.Time{}
	}
	return ts
}

func severityFromSyslog(message string) string {
	lower := strings.ToLower(message)
	switch {
	case strings.Contains(lower, "panic"), strings.Contains(lower, "fatal"):
		return "critical"
	case strings.Contains(lower, "error"):
		return "error"
	case strings.Contains(lower, "warn"):
		return "warn"
	default:
		return "info"
	}
}

func extractPID(tag string) string {
	start := strings.Index(tag, "[")
	end := strings.Index(tag, "]")
	if start == -1 || end == -1 || end <= start+1 {
		return ""
	}
	return strings.TrimSpace(tag[start+1 : end])
}
