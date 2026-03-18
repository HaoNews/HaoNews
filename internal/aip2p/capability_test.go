package aip2p

import (
	"testing"
	"time"
)

func TestCapabilityIndexUpdateAndGet(t *testing.T) {
	idx := NewCapabilityIndex()
	idx.Update(CapabilityEntry{
		Author: "agent://test",
		Tools:  []string{"translate", "summarize"},
		Models: []string{"gpt-4"},
	})

	entry, ok := idx.Get("agent://test")
	if !ok {
		t.Fatal("expected entry to be found")
	}
	if entry.Author != "agent://test" {
		t.Fatalf("got author=%s, want agent://test", entry.Author)
	}
	if len(entry.Tools) != 2 {
		t.Fatalf("got %d tools, want 2", len(entry.Tools))
	}
}

func TestCapabilityIndexCaseInsensitive(t *testing.T) {
	idx := NewCapabilityIndex()
	idx.Update(CapabilityEntry{Author: "Agent://Test", Tools: []string{"x"}})

	_, ok := idx.Get("agent://test")
	if !ok {
		t.Fatal("lookup should be case-insensitive")
	}
}

func TestCapabilityIndexAll(t *testing.T) {
	idx := NewCapabilityIndex()
	idx.Update(CapabilityEntry{Author: "agent://a", Tools: []string{"x"}})
	idx.Update(CapabilityEntry{Author: "agent://b", Tools: []string{"y"}})

	all := idx.All()
	if len(all) != 2 {
		t.Fatalf("got %d entries, want 2", len(all))
	}
}

func TestCapabilityIndexCount(t *testing.T) {
	idx := NewCapabilityIndex()
	if idx.Count() != 0 {
		t.Fatal("empty index should have count 0")
	}
	idx.Update(CapabilityEntry{Author: "agent://a"})
	idx.Update(CapabilityEntry{Author: "agent://b"})
	if idx.Count() != 2 {
		t.Fatalf("got count=%d, want 2", idx.Count())
	}
}

func TestCapabilityIndexFindByTool(t *testing.T) {
	idx := NewCapabilityIndex()
	idx.Update(CapabilityEntry{Author: "agent://a", Tools: []string{"translate", "summarize"}})
	idx.Update(CapabilityEntry{Author: "agent://b", Tools: []string{"code-review"}})
	idx.Update(CapabilityEntry{Author: "agent://c", Tools: []string{"Translate"}})

	results := idx.FindByTool("translate")
	if len(results) != 2 {
		t.Fatalf("got %d results, want 2", len(results))
	}
}

func TestCapabilityIndexFindByModel(t *testing.T) {
	idx := NewCapabilityIndex()
	idx.Update(CapabilityEntry{Author: "agent://a", Models: []string{"gpt-4", "claude-3"}})
	idx.Update(CapabilityEntry{Author: "agent://b", Models: []string{"gpt-4"}})
	idx.Update(CapabilityEntry{Author: "agent://c", Models: []string{"llama-3"}})

	results := idx.FindByModel("gpt-4")
	if len(results) != 2 {
		t.Fatalf("got %d results, want 2", len(results))
	}
}

func TestCapabilityIndexTTLExpiry(t *testing.T) {
	idx := NewCapabilityIndex()
	idx.ttl = 1 * time.Millisecond

	idx.Update(CapabilityEntry{Author: "agent://stale", Tools: []string{"x"}})
	time.Sleep(5 * time.Millisecond)

	_, ok := idx.Get("agent://stale")
	if ok {
		t.Fatal("stale entry should return ok=false")
	}

	all := idx.All()
	if len(all) != 0 {
		t.Fatalf("All() should return 0 for stale entries, got %d", len(all))
	}
}

func TestCapabilityIndexPrune(t *testing.T) {
	idx := NewCapabilityIndex()
	idx.ttl = 1 * time.Millisecond

	idx.Update(CapabilityEntry{Author: "agent://old"})
	time.Sleep(5 * time.Millisecond)
	idx.Update(CapabilityEntry{Author: "agent://new"})

	idx.Prune()

	if idx.Count() != 1 {
		t.Fatalf("after prune, got count=%d, want 1", idx.Count())
	}
}

func TestCapabilityIndexEmptyAuthor(t *testing.T) {
	idx := NewCapabilityIndex()
	idx.Update(CapabilityEntry{Author: "", Tools: []string{"x"}})
	if idx.Count() != 0 {
		t.Fatal("empty author should be ignored")
	}
}

func TestCapabilityIndexOverwrite(t *testing.T) {
	idx := NewCapabilityIndex()
	idx.Update(CapabilityEntry{Author: "agent://a", Tools: []string{"old"}})
	idx.Update(CapabilityEntry{Author: "agent://a", Tools: []string{"new"}})

	entry, ok := idx.Get("agent://a")
	if !ok {
		t.Fatal("entry should exist")
	}
	if len(entry.Tools) != 1 || entry.Tools[0] != "new" {
		t.Fatalf("expected tools=[new], got %v", entry.Tools)
	}
}
