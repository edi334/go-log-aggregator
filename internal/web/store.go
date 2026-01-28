package web

import (
	"strings"
	"sync"
	"time"
)

type Store struct {
	mu        sync.RWMutex
	maxAge    time.Duration
	maxEvents int
	events    []Event
}

func NewStore(maxAge time.Duration, maxEvents int) *Store {
	if maxEvents <= 0 {
		maxEvents = 10000
	}
	return &Store{
		maxAge:    maxAge,
		maxEvents: maxEvents,
		events:    make([]Event, 0, maxEvents),
	}
}

func (s *Store) Add(event Event) {
	if s == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.events = append(s.events, event)
	s.pruneLocked()
}

func (s *Store) Query(sources []string, since time.Time) []Event {
	if s == nil {
		return nil
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	allowed := make(map[string]struct{}, len(sources))
	for _, source := range sources {
		if source == "" {
			continue
		}
		allowed[strings.ToLower(source)] = struct{}{}
	}

	out := make([]Event, 0, len(s.events))
	for _, event := range s.events {
		if !since.IsZero() && event.Timestamp.Before(since) {
			continue
		}
		if len(allowed) > 0 {
			if _, ok := allowed[strings.ToLower(event.Source)]; !ok {
				continue
			}
		}
		out = append(out, event)
	}
	return out
}

func (s *Store) pruneLocked() {
	if s.maxAge > 0 {
		cutoff := time.Now().Add(-s.maxAge)
		pruned := s.events[:0]
		for _, event := range s.events {
			if event.Timestamp.Before(cutoff) {
				continue
			}
			pruned = append(pruned, event)
		}
		s.events = pruned
	}

	if len(s.events) > s.maxEvents {
		s.events = append([]Event(nil), s.events[len(s.events)-s.maxEvents:]...)
	}
}
