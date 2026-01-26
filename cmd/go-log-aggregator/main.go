package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"

	"go-log-aggregator/internal/alert"
	"go-log-aggregator/internal/config"
	"go-log-aggregator/internal/filter"
	"go-log-aggregator/internal/ingest"
	"go-log-aggregator/internal/parse"
)

func main() {
	var configPath string
	var regexFilter string
	var severityFilter string
	var sinceFilter string
	var untilFilter string
	var fieldFilters multiValue
	flag.StringVar(&configPath, "config", "config/config.json", "path to config file")
	flag.StringVar(&regexFilter, "regex", "", "regex filter applied to raw/message")
	flag.StringVar(&severityFilter, "severity", "", "severity filter (info, warn, error, critical)")
	flag.StringVar(&sinceFilter, "since", "", "only include logs since RFC3339 timestamp")
	flag.StringVar(&untilFilter, "until", "", "only include logs until RFC3339 timestamp")
	flag.Var(&fieldFilters, "field", "field filter key=value (repeatable)")
	flag.Parse()

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	if len(cfg.Sources) == 0 {
		fmt.Fprintln(os.Stdout, "no sources configured")
		return
	}

	fmt.Fprintln(os.Stdout, "configured sources:")
	for _, src := range cfg.Sources {
		fmt.Fprintf(os.Stdout, "- %s (%s) format=%s\n", src.Name, src.Path, src.Format)
	}

	criteria, err := buildCriteria(regexFilter, severityFilter, sinceFilter, untilFilter, fieldFilters)
	if err != nil {
		log.Fatalf("filters: %v", err)
	}

	alerts, err := alert.NewEvaluator(cfg.Alerts)
	if err != nil {
		log.Fatalf("alerts: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	events := make(chan ingest.Event, 128)
	errs := make(chan error, 16)

	for _, src := range cfg.Sources {
		if err := ingest.StartTailer(ctx, src, events, errs); err != nil {
			log.Printf("start tailer %s: %v", src.Name, err)
		}
	}

	log.Println("tailing configured sources (ctrl+c to stop)")
	for {
		select {
		case <-ctx.Done():
			return
		case err := <-errs:
			if err != nil {
				log.Printf("tailer error: %v", err)
			}
		case event := <-events:
			parsed, err := parse.ParseLine(sourceFormat(cfg.Sources, event.SourceName), event)
			if err != nil {
				parsed = parse.StructuredEvent{
					SourceName: event.SourceName,
					SourcePath: event.SourcePath,
					Format:     "unknown",
					Severity:   "unknown",
					Message:    event.Line,
					Raw:        event.Line,
				}
			}

			if !criteria.Matches(parsed) {
				continue
			}

			printEvent(parsed)
			for _, match := range alerts.Evaluate(parsed) {
				log.Printf("ALERT %s source=%s message=%s", match.RuleName, match.Event.SourceName, match.Event.Message)
			}
		}
	}
}

type multiValue []string

func (m *multiValue) String() string {
	if m == nil {
		return ""
	}
	return strings.Join(*m, ",")
}

func (m *multiValue) Set(value string) error {
	*m = append(*m, value)
	return nil
}

func buildCriteria(regexFilter, severityFilter, sinceFilter, untilFilter string, fieldFilters []string) (filter.Criteria, error) {
	var criteria filter.Criteria

	if strings.TrimSpace(regexFilter) != "" {
		re, err := regexp.Compile(regexFilter)
		if err != nil {
			return filter.Criteria{}, fmt.Errorf("regex filter: %w", err)
		}
		criteria.Regex = re
	}

	if strings.TrimSpace(severityFilter) != "" {
		criteria.Severity = strings.ToLower(strings.TrimSpace(severityFilter))
	}

	if strings.TrimSpace(sinceFilter) != "" {
		ts, err := time.Parse(time.RFC3339, sinceFilter)
		if err != nil {
			return filter.Criteria{}, fmt.Errorf("since filter: %w", err)
		}
		criteria.Since = ts
	}

	if strings.TrimSpace(untilFilter) != "" {
		ts, err := time.Parse(time.RFC3339, untilFilter)
		if err != nil {
			return filter.Criteria{}, fmt.Errorf("until filter: %w", err)
		}
		criteria.Until = ts
	}

	if len(fieldFilters) > 0 {
		fields, err := filter.ParseFieldAssignments(fieldFilters)
		if err != nil {
			return filter.Criteria{}, err
		}
		criteria.Fields = fields
	}

	return criteria, nil
}

func sourceFormat(sources []config.Source, name string) string {
	for _, src := range sources {
		if src.Name == name {
			return src.Format
		}
	}
	return ""
}

type outputEvent struct {
	Timestamp string            `json:"timestamp,omitempty"`
	Severity  string            `json:"severity,omitempty"`
	Message   string            `json:"message,omitempty"`
	Source    string            `json:"source"`
	Format    string            `json:"format,omitempty"`
	Fields    map[string]string `json:"fields,omitempty"`
	Raw       string            `json:"raw,omitempty"`
}

func printEvent(event parse.StructuredEvent) {
	out := outputEvent{
		Severity: event.Severity,
		Message:  event.Message,
		Source:   event.SourceName,
		Format:   event.Format,
		Fields:   event.Fields,
		Raw:      event.Raw,
	}
	if !event.Timestamp.IsZero() {
		out.Timestamp = event.Timestamp.Format(time.RFC3339)
	}

	data, err := json.Marshal(out)
	if err != nil {
		log.Printf("marshal output: %v", err)
		return
	}

	fmt.Fprintln(os.Stdout, string(data))
}
