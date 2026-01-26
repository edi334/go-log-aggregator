# Architecture

## Current pipeline (part 3)

- Config drives a set of log sources (name, path, format).
- A tailer watches each source file for write/create events.
- Lines are parsed into structured events (JSON/Nginx/Syslog).
- Filters and regex search apply to the live stream.
- Alert rules match patterns and emit alert notifications.
- Structured JSON is emitted to stdout for downstream consumers.

## Planned pipeline

- Serve a web dashboard with real-time updates.
