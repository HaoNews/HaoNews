package aip2p

import (
	"testing"
	"time"
)

func TestBuildAndLoadMessage(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	msg, body, err := BuildMessage(MessageInput{
		Kind:    "post",
		Author:  "agent://demo/alice",
		Channel: "general",
		Title:   "hello",
		Body:    "hello world",
		Tags:    []string{"demo", "demo", "test"},
		Extensions: map[string]any{
			"project": "latest.org",
		},
		CreatedAt: time.Date(2026, 3, 12, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("BuildMessage error = %v", err)
	}
	if err := WriteMessage(dir, msg, body); err != nil {
		t.Fatalf("WriteMessage error = %v", err)
	}

	gotMsg, gotBody, err := LoadMessage(dir)
	if err != nil {
		t.Fatalf("LoadMessage error = %v", err)
	}

	if gotBody != "hello world" {
		t.Fatalf("body = %q, want %q", gotBody, "hello world")
	}
	if gotMsg.Protocol != ProtocolVersion {
		t.Fatalf("protocol = %q, want %q", gotMsg.Protocol, ProtocolVersion)
	}
	if len(gotMsg.Tags) != 2 {
		t.Fatalf("tags len = %d, want 2", len(gotMsg.Tags))
	}
	if gotMsg.Extensions["project"] != "latest.org" {
		t.Fatalf("extensions project = %v, want latest.org", gotMsg.Extensions["project"])
	}
}
