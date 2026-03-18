package aip2p

import (
	"strings"
	"sync"
	"time"
)

// v0.4: Capability discovery - agents announce their capabilities,
// other agents can query the index to find who can do what.

type CapabilityEntry struct {
	Author    string   `json:"author"`
	Tools     []string `json:"tools,omitempty"`
	Models    []string `json:"models,omitempty"`
	Languages []string `json:"languages,omitempty"`
	LatencyMs int64    `json:"latency_ms,omitempty"`
	MaxTokens int64    `json:"max_tokens,omitempty"`
	PubKey    string   `json:"pubkey,omitempty"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CapabilityIndex struct {
	mu      sync.RWMutex
	entries map[string]*CapabilityEntry // keyed by author
	ttl     time.Duration
}

func NewCapabilityIndex() *CapabilityIndex {
	return &CapabilityIndex{
		entries: make(map[string]*CapabilityEntry),
		ttl:     30 * time.Minute,
	}
}

// Update registers or refreshes a capability entry
func (idx *CapabilityIndex) Update(entry CapabilityEntry) {
	author := strings.ToLower(strings.TrimSpace(entry.Author))
	if author == "" {
		return
	}
	if entry.UpdatedAt.IsZero() {
		entry.UpdatedAt = time.Now().UTC()
	}
	idx.mu.Lock()
	defer idx.mu.Unlock()
	idx.entries[author] = &entry
}

// Get returns the capability entry for an author
func (idx *CapabilityIndex) Get(author string) (*CapabilityEntry, bool) {
	author = strings.ToLower(strings.TrimSpace(author))
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	e, ok := idx.entries[author]
	if !ok {
		return nil, false
	}
	if time.Since(e.UpdatedAt) > idx.ttl {
		return e, false // stale
	}
	return e, true
}

// All returns all non-stale capability entries
func (idx *CapabilityIndex) All() []CapabilityEntry {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	now := time.Now().UTC()
	result := make([]CapabilityEntry, 0, len(idx.entries))
	for _, e := range idx.entries {
		if now.Sub(e.UpdatedAt) <= idx.ttl {
			result = append(result, *e)
		}
	}
	return result
}

// Count returns the number of non-stale entries
func (idx *CapabilityIndex) Count() int {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	now := time.Now().UTC()
	count := 0
	for _, e := range idx.entries {
		if now.Sub(e.UpdatedAt) <= idx.ttl {
			count++
		}
	}
	return count
}

// Prune removes stale entries
func (idx *CapabilityIndex) Prune() {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	now := time.Now().UTC()
	for author, e := range idx.entries {
		if now.Sub(e.UpdatedAt) > idx.ttl {
			delete(idx.entries, author)
		}
	}
}

// FindByTool returns agents that have a specific tool
func (idx *CapabilityIndex) FindByTool(tool string) []CapabilityEntry {
	tool = strings.ToLower(strings.TrimSpace(tool))
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	now := time.Now().UTC()
	var result []CapabilityEntry
	for _, e := range idx.entries {
		if now.Sub(e.UpdatedAt) > idx.ttl {
			continue
		}
		for _, t := range e.Tools {
			if strings.ToLower(strings.TrimSpace(t)) == tool {
				result = append(result, *e)
				break
			}
		}
	}
	return result
}

// FindByModel returns agents that support a specific model
func (idx *CapabilityIndex) FindByModel(model string) []CapabilityEntry {
	model = strings.ToLower(strings.TrimSpace(model))
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	now := time.Now().UTC()
	var result []CapabilityEntry
	for _, e := range idx.entries {
		if now.Sub(e.UpdatedAt) > idx.ttl {
			continue
		}
		for _, m := range e.Models {
			if strings.ToLower(strings.TrimSpace(m)) == model {
				result = append(result, *e)
				break
			}
		}
	}
	return result
}
