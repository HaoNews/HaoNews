package aip2p

import (
	"strings"
	"sync"
	"time"
)

// v0.3: per-author rate limiting to prevent spam/flooding

const (
	defaultRateLimitWindow     = 60 * time.Second
	defaultMaxMessagesPerWindow = 10
)

type RateLimitConfig struct {
	Window            time.Duration `json:"rate_limit_window"`
	MaxPerWindow      int           `json:"max_messages_per_window"`
}

type rateLimitEntry struct {
	timestamps []time.Time
}

type RateLimiter struct {
	mu      sync.Mutex
	entries map[string]*rateLimitEntry
	window  time.Duration
	max     int
}

func NewRateLimiter(cfg RateLimitConfig) *RateLimiter {
	window := cfg.Window
	if window <= 0 {
		window = defaultRateLimitWindow
	}
	max := cfg.MaxPerWindow
	if max <= 0 {
		max = defaultMaxMessagesPerWindow
	}
	return &RateLimiter{
		entries: make(map[string]*rateLimitEntry),
		window:  window,
		max:     max,
	}
}

// Allow checks if the author is within rate limits. Returns true if allowed.
func (r *RateLimiter) Allow(author string) bool {
	author = strings.ToLower(strings.TrimSpace(author))
	if author == "" {
		return true
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now().UTC()
	entry, ok := r.entries[author]
	if !ok {
		entry = &rateLimitEntry{}
		r.entries[author] = entry
	}

	// Prune old timestamps outside the window
	cutoff := now.Add(-r.window)
	pruned := entry.timestamps[:0]
	for _, ts := range entry.timestamps {
		if ts.After(cutoff) {
			pruned = append(pruned, ts)
		}
	}
	entry.timestamps = pruned

	if len(entry.timestamps) >= r.max {
		return false
	}

	entry.timestamps = append(entry.timestamps, now)
	return true
}

// Reset clears all rate limit state
func (r *RateLimiter) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.entries = make(map[string]*rateLimitEntry)
}

// Cleanup removes stale entries older than 2x the window
func (r *RateLimiter) Cleanup() {
	r.mu.Lock()
	defer r.mu.Unlock()

	cutoff := time.Now().UTC().Add(-2 * r.window)
	for author, entry := range r.entries {
		if len(entry.timestamps) == 0 {
			delete(r.entries, author)
			continue
		}
		latest := entry.timestamps[len(entry.timestamps)-1]
		if latest.Before(cutoff) {
			delete(r.entries, author)
		}
	}
}
