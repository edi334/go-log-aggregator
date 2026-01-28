package web

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

const dashboardHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <title>go-log-aggregator</title>
  <style>
    body { font-family: Arial, sans-serif; margin: 0; background: #0f1115; color: #e5e7eb; }
    header { padding: 16px; background: #111827; border-bottom: 1px solid #1f2937; }
    main { padding: 16px; }
    #status { color: #9ca3af; font-size: 12px; margin-left: 8px; }
    #log { white-space: pre; font-family: Consolas, monospace; font-size: 12px; }
    .row { display: flex; align-items: center; gap: 8px; }
    .pill { background: #1f2937; padding: 4px 8px; border-radius: 999px; font-size: 12px; }
    .controls { display: flex; gap: 16px; flex-wrap: wrap; margin: 12px 0; }
    .controls label { font-size: 12px; color: #9ca3af; }
    .sources { display: flex; gap: 8px; flex-wrap: wrap; }
    .sources label { font-size: 12px; background: #111827; border: 1px solid #1f2937; padding: 4px 8px; border-radius: 6px; }
  </style>
</head>
<body>
  <header>
    <div class="row">
      <strong>go-log-aggregator</strong>
      <span class="pill">live stream</span>
      <span id="status">connecting...</span>
    </div>
  </header>
  <main>
    <div class="controls">
      <label>
        Window
        <select id="window">
          <option value="1m">last 1 minute</option>
          <option value="15m">last 15 minutes</option>
          <option value="3h">last 3 hours</option>
          <option value="24h">last 1 day</option>
          <option value="168h">last 1 week</option>
        </select>
      </label>
      <div class="sources" id="sources"></div>
    </div>
    <div id="log"></div>
  </main>
  <script>
    const status = document.getElementById('status');
    const log = document.getElementById('log');
    const windowSelect = document.getElementById('window');
    const sourcesWrap = document.getElementById('sources');
    let selectedSources = new Set();
    let stream;

    function formatEvent(event) {
      const ts = event.timestamp || event.received_at || '';
      const sev = event.severity || 'unknown';
      const msg = event.message || event.raw || '';
      return '[' + ts + '] ' + event.source + ' ' + sev + ' ' + msg;
    }

    function appendLine(line) {
      log.textContent += line + "\n";
      log.scrollTop = log.scrollHeight;
    }

    function currentWindow() {
      return windowSelect.value;
    }

    function selectedSourcesList() {
      return Array.from(selectedSources.values());
    }

    function loadSources() {
      return fetch('/api/sources')
        .then(res => res.json())
        .then(sources => {
          sourcesWrap.innerHTML = '';
          sources.forEach(source => {
          const id = 'src-' + source;
            const label = document.createElement('label');
            const input = document.createElement('input');
            input.type = 'checkbox';
            input.value = source;
            input.checked = true;
            input.addEventListener('change', () => {
              if (input.checked) {
                selectedSources.add(source);
              } else {
                selectedSources.delete(source);
              }
              refreshHistory();
            });
            selectedSources.add(source);
            label.appendChild(input);
            label.append(' ' + source);
            sourcesWrap.appendChild(label);
          });
        });
    }

    function refreshHistory() {
      log.textContent = '';
      const params = new URLSearchParams();
      params.set('window', currentWindow());
      const sources = selectedSourcesList();
      if (sources.length > 0) {
        params.set('sources', sources.join(','));
      }
      fetch('/api/events?' + params.toString())
        .then(res => res.json())
        .then(events => {
          events.forEach(event => appendLine(formatEvent(event)));
        });
    }

    function openStream() {
      if (stream) {
        stream.close();
      }
      stream = new EventSource('/stream');
      stream.onopen = () => { status.textContent = 'connected'; };
      stream.onerror = () => { status.textContent = 'disconnected'; };
      stream.onmessage = (evt) => {
        try {
          const event = JSON.parse(evt.data);
          const windowMs = parseWindow(currentWindow());
          const ts = Date.parse(event.timestamp || event.received_at || 0);
          const now = Date.now();
          if (windowMs && ts && ts < now - windowMs) {
            return;
          }
          if (selectedSources.size > 0 && !selectedSources.has(event.source)) {
            return;
          }
          appendLine(formatEvent(event));
        } catch (err) {
          appendLine(evt.data);
        }
      };
    }

    function parseWindow(value) {
      if (!value) return 0;
      const match = value.match(/^(\d+)(m|h)$/);
      if (!match) return 0;
      const amount = parseInt(match[1], 10);
      const unit = match[2];
      if (unit === 'm') return amount * 60 * 1000;
      if (unit === 'h') return amount * 60 * 60 * 1000;
      return 0;
    }

    windowSelect.addEventListener('change', () => {
      refreshHistory();
    });

    loadSources().then(() => {
      refreshHistory();
      openStream();
    });
  </script>
</body>
</html>`

func StartServer(ctx context.Context, addr string, hub *Hub, store *Store, sources []string) error {
	if addr == "" {
		return fmt.Errorf("http address is required")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(dashboardHTML))
	})
	mux.HandleFunc("/api/sources", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, sources)
	})
	mux.HandleFunc("/api/events", func(w http.ResponseWriter, r *http.Request) {
		if store == nil {
			writeJSON(w, []Event{})
			return
		}
		windowValue := strings.TrimSpace(r.URL.Query().Get("window"))
		since := time.Time{}
		if windowValue != "" {
			if dur, err := time.ParseDuration(windowValue); err == nil {
				since = time.Now().Add(-dur)
			}
		}
		sourceList := parseSources(r.URL.Query().Get("sources"))
		events := store.Query(sourceList, since)
		writeJSON(w, events)
	})
	mux.HandleFunc("/stream", func(w http.ResponseWriter, r *http.Request) {
		stream(w, r, hub)
	})

	server := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("http shutdown: %v", err)
		}
	}()

	log.Printf("dashboard listening on http://%s", addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func stream(w http.ResponseWriter, r *http.Request, hub *Hub) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	client := make(chan []byte, 32)
	hub.Register(client)
	defer hub.Unregister(client)

	_, _ = w.Write([]byte(": connected\n\n"))
	flusher.Flush()

	for {
		select {
		case <-r.Context().Done():
			return
		case payload, ok := <-client:
			if !ok {
				return
			}
			_, _ = fmt.Fprintf(w, "data: %s\n\n", payload)
			flusher.Flush()
		}
	}
}

func parseSources(value string) []string {
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		out = append(out, part)
	}
	return out
}

func writeJSON(w http.ResponseWriter, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
