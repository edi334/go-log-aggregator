package parse

import "time"

type StructuredEvent struct {
	SourceName string
	SourcePath string
	Format     string
	Timestamp  time.Time
	Severity   string
	Message    string
	Fields     map[string]string
	Raw        string
}
