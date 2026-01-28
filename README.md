# go-log-aggregator

Log aggregation playground built in parts. The current milestone tails
configured log files, parses multiple formats, supports filters/regex,
evaluates alerts, and serves a live dashboard with historical backfill.

## Run

1. Ensure the log paths in `config/config.json` exist.
2. Start the tailer:
   - `go run ./cmd/go-log-aggregator -config config/config.json`
3. Append lines to any configured log file to see updates live.

## Filters

- Regex search: `-regex "panic|timeout"`
- Severity: `-severity error`
- Time bounds: `-since 2026-01-26T09:00:00Z -until 2026-01-26T10:00:00Z`
- Field match (repeatable): `-field service=api -field status=500`

All filters are applied to the live stream.

## Dashboard

The web dashboard streams live events over SSE and lets you:

- Select time windows (1m, 15m, 3h, 1d, 1w).
- Toggle which sources to display.
- See real-time updates in the same view.

Run it:
- `go run ./cmd/go-log-aggregator -config config/config.json -http-addr :8080`
- Open `http://localhost:8080` in your browser.

To disable the dashboard: `-http-addr ""`.

Backfill behavior:
- By default, existing log content is read once on startup.
- Control with `-backfill` and `-backfill-lines`.

## Alerts

Alert rules live in `config/config.json` under `alerts` and fire when the
pattern matches the raw line or parsed message. Matches are logged with
the `ALERT` prefix.

## Next

- Persist events to disk for longer history.
- Add server-side search endpoints for dashboard queries.