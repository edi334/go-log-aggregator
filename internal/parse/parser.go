package parse

import (
	"fmt"
	"strings"

	"go-log-aggregator/internal/ingest"
)

func ParseLine(format string, event ingest.Event) (StructuredEvent, error) {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "json":
		return parseJSON(event)
	case "nginx", "apache":
		return parseNginx(event)
	case "syslog":
		return parseSyslog(event)
	default:
		return StructuredEvent{}, fmt.Errorf("unsupported format: %s", format)
	}
}

func normalizeSeverity(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return "unknown"
	}

	switch value {
	case "panic", "fatal", "critical", "crit":
		return "critical"
	case "err", "error":
		return "error"
	case "warn", "warning":
		return "warn"
	case "info", "information":
		return "info"
	case "debug", "trace":
		return "debug"
	default:
		return value
	}
}
