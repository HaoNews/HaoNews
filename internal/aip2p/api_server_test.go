package aip2p

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestAPIServer(t *testing.T) *APIServer {
	t.Helper()
	store, err := OpenStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return NewAPIServer(store)
}

func TestAPIStatus(t *testing.T) {
	srv := newTestAPIServer(t)
	req := httptest.NewRequest("GET", "/api/v1/status", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("status code=%d, want 200", w.Code)
	}
	var resp APIResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if !resp.OK {
		t.Fatalf("expected ok=true, got message=%s", resp.Message)
	}
}

func TestAPIStatusMethodNotAllowed(t *testing.T) {
	srv := newTestAPIServer(t)
	req := httptest.NewRequest("POST", "/api/v1/status", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Code != 405 {
		t.Fatalf("status code=%d, want 405", w.Code)
	}
}

func TestAPIFeedEmpty(t *testing.T) {
	srv := newTestAPIServer(t)
	req := httptest.NewRequest("GET", "/api/v1/feed", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("status code=%d, want 200", w.Code)
	}
	var resp APIResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if !resp.OK {
		t.Fatal("expected ok=true")
	}
}

func TestAPICapabilityAnnounceAndQuery(t *testing.T) {
	srv := newTestAPIServer(t)

	// Announce
	body := `{"author":"agent://test","tools":["translate","summarize"],"models":["gpt-4"]}`
	req := httptest.NewRequest("POST", "/api/v1/capabilities/announce", strings.NewReader(body))
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("announce status=%d, want 200", w.Code)
	}

	// Query all
	req = httptest.NewRequest("GET", "/api/v1/capabilities", nil)
	w = httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("capabilities status=%d, want 200", w.Code)
	}
	var resp APIResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if !resp.OK {
		t.Fatal("expected ok=true")
	}
	caps, ok := resp.Data.([]any)
	if !ok || len(caps) != 1 {
		t.Fatalf("expected 1 capability, got %v", resp.Data)
	}

	// Query by tool
	req = httptest.NewRequest("GET", "/api/v1/capabilities?tool=translate", nil)
	w = httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	json.NewDecoder(w.Body).Decode(&resp)
	caps, _ = resp.Data.([]any)
	if len(caps) != 1 {
		t.Fatalf("expected 1 result for tool=translate, got %d", len(caps))
	}

	// Query by tool that doesn't exist
	req = httptest.NewRequest("GET", "/api/v1/capabilities?tool=nonexistent", nil)
	w = httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Data != nil {
		caps, _ = resp.Data.([]any)
		if len(caps) != 0 {
			t.Fatalf("expected 0 results, got %d", len(caps))
		}
	}
}

func TestAPICapabilityAnnounceNoAuthor(t *testing.T) {
	srv := newTestAPIServer(t)
	body := `{"tools":["x"]}`
	req := httptest.NewRequest("POST", "/api/v1/capabilities/announce", strings.NewReader(body))
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Code != 400 {
		t.Fatalf("status=%d, want 400", w.Code)
	}
}

func TestAPICapabilityAnnounceBadJSON(t *testing.T) {
	srv := newTestAPIServer(t)
	req := httptest.NewRequest("POST", "/api/v1/capabilities/announce", strings.NewReader("{bad"))
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Code != 400 {
		t.Fatalf("status=%d, want 400", w.Code)
	}
}

func TestAPIPublishBadJSON(t *testing.T) {
	srv := newTestAPIServer(t)
	req := httptest.NewRequest("POST", "/api/v1/publish", strings.NewReader("not json"))
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Code != 400 {
		t.Fatalf("status=%d, want 400", w.Code)
	}
}

func TestAPIPublishMethodNotAllowed(t *testing.T) {
	srv := newTestAPIServer(t)
	req := httptest.NewRequest("GET", "/api/v1/publish", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Code != 405 {
		t.Fatalf("status=%d, want 405", w.Code)
	}
}

func TestAPIPostNotFound(t *testing.T) {
	srv := newTestAPIServer(t)
	req := httptest.NewRequest("GET", "/api/v1/posts/nonexistent", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Code != 404 {
		t.Fatalf("status=%d, want 404", w.Code)
	}
}

func TestAPIPostNoInfohash(t *testing.T) {
	srv := newTestAPIServer(t)
	req := httptest.NewRequest("GET", "/api/v1/posts/", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Code != 400 {
		t.Fatalf("status=%d, want 400", w.Code)
	}
}

func TestAPIPeers(t *testing.T) {
	srv := newTestAPIServer(t)
	req := httptest.NewRequest("GET", "/api/v1/peers", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("status=%d, want 200", w.Code)
	}
}

func TestAPISubscribeAddAndGet(t *testing.T) {
	srv := newTestAPIServer(t)

	// Add subscriptions
	body := `{"action":"add","topics":["ai","p2p"],"channels":["general"]}`
	req := httptest.NewRequest("POST", "/api/v1/subscribe", strings.NewReader(body))
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("subscribe add status=%d, want 200", w.Code)
	}

	// Get subscriptions
	req = httptest.NewRequest("GET", "/api/v1/subscribe", nil)
	w = httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("subscribe get status=%d, want 200", w.Code)
	}
	var resp APIResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if !resp.OK {
		t.Fatal("expected ok=true")
	}
}

func TestAPISubscribeRemove(t *testing.T) {
	srv := newTestAPIServer(t)

	// Add first
	body := `{"action":"add","topics":["ai","p2p","ml"]}`
	req := httptest.NewRequest("POST", "/api/v1/subscribe", strings.NewReader(body))
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	// Remove one
	body = `{"action":"remove","topics":["p2p"]}`
	req = httptest.NewRequest("POST", "/api/v1/subscribe", strings.NewReader(body))
	w = httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("subscribe remove status=%d, want 200", w.Code)
	}
}

func TestAPISubscribeBadAction(t *testing.T) {
	srv := newTestAPIServer(t)
	body := `{"action":"invalid"}`
	req := httptest.NewRequest("POST", "/api/v1/subscribe", strings.NewReader(body))
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Code != 400 {
		t.Fatalf("status=%d, want 400", w.Code)
	}
}
