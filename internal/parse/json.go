package parse

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"go-log-aggregator/internal/ingest"
)

func parseJSON(event ingest.Event) (StructuredEvent, error) {
	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(event.Line), &payload); err != nil {
		return StructuredEvent{}, fmt.Errorf("json parse: %w", err)
	}

	timestamp := extractTimestamp(payload)
	message := extractString(payload, "msg", "message")
	severity := normalizeSeverity(extractString(payload, "level", "severity"))

	fields := make(map[string]string, len(payload))
	for key, value := range payload {
		if key == "timestamp" || key == "time" || key == "ts" || key == "msg" || key == "message" || key == "level" || key == "severity" {
			continue
		}
		fields[key] = fmt.Sprint(value)
	}

	if message == "" {
		message = event.Line
	}

	return StructuredEvent{
		SourceName: event.SourceName,
		SourcePath: event.SourcePath,
		Format:     "json",
		Timestamp:  timestamp,
		ReceivedAt: event.ReceivedAt,
		Severity:   severity,
		Message:    message,
		Fields:     fields,
		Raw:        event.Line,
	}, nil
}

func extractTimestamp(payload map[string]interface{}) time.Time {
	for _, key := range []string{"timestamp", "time", "ts"} {
		if value, ok := payload[key]; ok {
			if ts, ok := parseTimestamp(value); ok {
				return ts
			}
		}
	}
	return time.Time{}
}

func parseTimestamp(value interface{}) (time.Time, bool) {
	switch v := value.(type) {
	case string:
		for _, layout := range []string{time.RFC3339Nano, time.RFC3339} {
			if ts, err := time.Parse(layout, v); err == nil {
				return ts, true
			}
		}
	case float64:
		sec := int64(v)
		return time.Unix(sec, 0), true
	case int64:
		return time.Unix(v, 0), true
	case int:
		return time.Unix(int64(v), 0), true
	case json.Number:
		if v == "" {
			return time.Time{}, false
		}
		if i, err := v.Int64(); err == nil {
			return time.Unix(i, 0), true
		}
		if f, err := v.Float64(); err == nil {
			return time.Unix(int64(f), 0), true
		}
	}

	return time.Time{}, false
}

func extractString(payload map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		if value, ok := payload[key]; ok {
			switch v := value.(type) {
			case string:
				return v
			case json.Number:
				return v.String()
			case float64:
				return strconv.FormatFloat(v, 'f', -1, 64)
			default:
				return fmt.Sprint(v)
			}
		}
	}
	return ""
}
