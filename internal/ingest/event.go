package ingest

import "time"

type Event struct {
	SourceName string
	SourcePath string
	Line       string
	ReceivedAt time.Time
}
