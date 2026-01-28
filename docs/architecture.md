# Architecture

## Current pipeline (part 4)

- Config drives a set of log sources (name, path, format).
- A tailer watches each source file for write/create events.
- Lines are parsed into structured events (JSON/Nginx/Syslog).
- Filters and regex search apply to the live stream.
- Alert rules match patterns and emit alert notifications.
- Structured JSON is emitted to stdout for downstream consumers.
- Live events are broadcast to the web dashboard over SSE.
- Dashboard pulls recent events by source/window.
- Startup backfill seeds the in-memory store for recent history.

## Planned pipeline

- Add persistence, indexing, and historical queries.
- Add server-side search endpoints for dashboard queries.
