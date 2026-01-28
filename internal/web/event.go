package web

import "time"

type Event struct {
	Timestamp  time.Time         `json:"timestamp"`
	ReceivedAt time.Time         `json:"received_at"`
	Severity   string            `json:"severity,omitempty"`
	Message    string            `json:"message,omitempty"`
	Source     string            `json:"source"`
	Format     string            `json:"format,omitempty"`
	Fields     map[string]string `json:"fields,omitempty"`
	Raw        string            `json:"raw,omitempty"`
}
