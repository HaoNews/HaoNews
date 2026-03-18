package aip2p

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/anacrolix/torrent/bencode"
)

// v0.4: Local HTTP API for Agent integration
// All endpoints are under /api/v1/ and bound to localhost only.

type APIServer struct {
	store        *Store
	mux          *http.ServeMux
	capabilities *CapabilityIndex
}

type APIPublishRequest struct {
	Author       string         `json:"author"`
	Kind         string         `json:"kind"`
	Channel      string         `json:"channel"`
	Title        string         `json:"title"`
	Body         string         `json:"body"`
	ContentType  string         `json:"content_type"`
	BodyEncoding string         `json:"body_encoding"`
	Tags         []string       `json:"tags"`
	IdentityFile string         `json:"identity_file"`
	ReplyTo      *MessageLink   `json:"reply_to,omitempty"`
	Nonce        string         `json:"nonce"`
	Extensions   map[string]any `json:"extensions,omitempty"`
}

type APIResponse struct {
	OK      bool   `json:"ok"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

func NewAPIServer(store *Store) *APIServer {
	s := &APIServer{
		store:        store,
		mux:          http.NewServeMux(),
		capabilities: NewCapabilityIndex(),
	}
	s.mux.HandleFunc("/api/v1/publish", s.handlePublish)
	s.mux.HandleFunc("/api/v1/feed", s.handleFeed)
	s.mux.HandleFunc("/api/v1/posts/", s.handlePost)
	s.mux.HandleFunc("/api/v1/status", s.handleStatus)
	s.mux.HandleFunc("/api/v1/peers", s.handlePeers)
	s.mux.HandleFunc("/api/v1/capabilities", s.handleCapabilities)
	s.mux.HandleFunc("/api/v1/capabilities/announce", s.handleCapabilityAnnounce)
	s.mux.HandleFunc("/api/v1/subscribe", s.handleSubscribe)
	return s
}

func (s *APIServer) Handler() http.Handler {
	return s.mux
}

func (s *APIServer) Capabilities() *CapabilityIndex {
	return s.capabilities
}

func (s *APIServer) writeJSON(w http.ResponseWriter, status int, resp APIResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp)
}

// POST /api/v1/publish
func (s *APIServer) handlePublish(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.writeJSON(w, 405, APIResponse{OK: false, Message: "method not allowed"})
		return
	}
	var req APIPublishRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeJSON(w, 400, APIResponse{OK: false, Message: "invalid JSON: " + err.Error()})
		return
	}

	input := MessageInput{
		Kind:         req.Kind,
		Author:       req.Author,
		Channel:      req.Channel,
		Title:        req.Title,
		Body:         req.Body,
		ContentType:  req.ContentType,
		BodyEncoding: req.BodyEncoding,
		Tags:         req.Tags,
		Nonce:        req.Nonce,
		ReplyTo:      req.ReplyTo,
		Extensions:   req.Extensions,
	}

	// Load identity if provided
	if req.IdentityFile != "" {
		identity, err := LoadAgentIdentity(req.IdentityFile)
		if err != nil {
			s.writeJSON(w, 400, APIResponse{OK: false, Message: "identity error: " + err.Error()})
			return
		}
		input.Identity = &identity
		if input.Author == "" {
			input.Author = identity.Author
		}
	}

	// Verify PoW if present in extensions
	if req.Extensions != nil {
		if err := VerifyMessagePoW(req.Extensions, req.Author, req.Title, req.Body); err != nil {
			s.writeJSON(w, 400, APIResponse{OK: false, Message: "pow verification failed: " + err.Error()})
			return
		}
	}

	result, err := PublishMessage(s.store, input)
	if err != nil {
		s.writeJSON(w, 500, APIResponse{OK: false, Message: err.Error()})
		return
	}
	s.writeJSON(w, 200, APIResponse{OK: true, Data: result})
}

// GET /api/v1/feed
func (s *APIServer) handleFeed(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.writeJSON(w, 405, APIResponse{OK: false, Message: "method not allowed"})
		return
	}
	limit := 50
	if v := r.URL.Query().Get("limit"); v != "" {
		fmt.Sscanf(v, "%d", &limit)
		if limit <= 0 || limit > 200 {
			limit = 50
		}
	}

	entries, err := os.ReadDir(s.store.DataDir)
	if err != nil {
		s.writeJSON(w, 500, APIResponse{OK: false, Message: err.Error()})
		return
	}

	type feedItem struct {
		InfoHash  string         `json:"infohash,omitempty"`
		Kind      string         `json:"kind"`
		Author    string         `json:"author"`
		Title     string         `json:"title"`
		Channel   string         `json:"channel"`
		CreatedAt string         `json:"created_at"`
		BodyFile  string         `json:"body_file"`
		Body      string         `json:"body"`
		Message   *Message       `json:"message,omitempty"`
		ReplyTo   *MessageLink   `json:"reply_to,omitempty"`
	}

	// Build infohash lookup from torrent dir
	torrentMap := s.buildTorrentMap()

	var items []feedItem
	for i := len(entries) - 1; i >= 0 && len(items) < limit; i-- {
		e := entries[i]
		if !e.IsDir() {
			continue
		}
		dir := filepath.Join(s.store.DataDir, e.Name())
		msg, body, err := LoadMessage(dir)
		if err != nil {
			continue
		}
		item := feedItem{
			InfoHash:  torrentMap[e.Name()],
			Kind:      msg.Kind,
			Author:    msg.Author,
			Title:     msg.Title,
			Channel:   msg.Channel,
			CreatedAt: msg.CreatedAt,
			BodyFile:  msg.BodyFile,
			Body:      body,
			ReplyTo:   msg.ReplyTo,
		}
		if msg.Kind == "task-result" || msg.Kind == "task-assign" {
			item.Message = &msg
		}
		items = append(items, item)
	}
	s.writeJSON(w, 200, APIResponse{OK: true, Data: items})
}

// GET /api/v1/posts/{infohash}
func (s *APIServer) handlePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.writeJSON(w, 405, APIResponse{OK: false, Message: "method not allowed"})
		return
	}
	infohash := strings.TrimPrefix(r.URL.Path, "/api/v1/posts/")
	infohash = strings.TrimSpace(infohash)
	if infohash == "" {
		s.writeJSON(w, 400, APIResponse{OK: false, Message: "infohash required"})
		return
	}

	entries, err := os.ReadDir(s.store.DataDir)
	if err != nil {
		s.writeJSON(w, 500, APIResponse{OK: false, Message: err.Error()})
		return
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		dir := filepath.Join(s.store.DataDir, e.Name())
		msg, body, err := LoadMessage(dir)
		if err != nil {
			continue
		}
		// Match by directory name containing infohash or by checking torrent
		if strings.Contains(strings.ToLower(e.Name()), strings.ToLower(infohash)) {
			s.writeJSON(w, 200, APIResponse{OK: true, Data: map[string]any{
				"message": msg,
				"body":    body,
				"dir":     e.Name(),
			}})
			return
		}
		_ = msg
	}
	s.writeJSON(w, 404, APIResponse{OK: false, Message: "post not found"})
}

// GET /api/v1/status
func (s *APIServer) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.writeJSON(w, 405, APIResponse{OK: false, Message: "method not allowed"})
		return
	}

	// Count bundles
	bundleCount := 0
	var totalSize int64
	if entries, err := os.ReadDir(s.store.DataDir); err == nil {
		for _, e := range entries {
			if e.IsDir() {
				bundleCount++
				dirPath := filepath.Join(s.store.DataDir, e.Name())
				if files, err := os.ReadDir(dirPath); err == nil {
					for _, f := range files {
						if info, err := f.Info(); err == nil {
							totalSize += info.Size()
						}
					}
				}
			}
		}
	}

	torrentCount := 0
	if entries, err := os.ReadDir(s.store.TorrentDir); err == nil {
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".torrent") {
				torrentCount++
			}
		}
	}

	// Load sync status if available
	var syncStatus map[string]any
	statusPath := filepath.Join(s.store.Root, "sync", "status.json")
	if data, err := os.ReadFile(statusPath); err == nil {
		json.Unmarshal(data, &syncStatus)
	}

	status := map[string]any{
		"version":       ProtocolVersion,
		"store_root":    s.store.Root,
		"bundles":       bundleCount,
		"torrents":      torrentCount,
		"total_size_mb": float64(totalSize) / (1024 * 1024),
		"capabilities":  s.capabilities.Count(),
		"timestamp":     time.Now().UTC().Format(time.RFC3339),
	}
	if syncStatus != nil {
		status["sync"] = syncStatus
	}
	s.writeJSON(w, 200, APIResponse{OK: true, Data: status})
}

// GET /api/v1/peers
func (s *APIServer) handlePeers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.writeJSON(w, 405, APIResponse{OK: false, Message: "method not allowed"})
		return
	}
	statusPath := filepath.Join(s.store.Root, "sync", "status.json")
	data, err := os.ReadFile(statusPath)
	if err != nil {
		s.writeJSON(w, 200, APIResponse{OK: true, Data: map[string]any{"peers": []any{}, "message": "sync daemon not running"}})
		return
	}
	var status map[string]any
	if err := json.Unmarshal(data, &status); err != nil {
		s.writeJSON(w, 500, APIResponse{OK: false, Message: err.Error()})
		return
	}
	s.writeJSON(w, 200, APIResponse{OK: true, Data: status["libp2p"]})
}

// GET /api/v1/capabilities
func (s *APIServer) handleCapabilities(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.writeJSON(w, 405, APIResponse{OK: false, Message: "method not allowed"})
		return
	}
	tool := r.URL.Query().Get("tool")
	model := r.URL.Query().Get("model")

	caps := s.capabilities.All()
	if tool != "" || model != "" {
		var filtered []CapabilityEntry
		for _, c := range caps {
			if tool != "" && !containsString(c.Tools, tool) {
				continue
			}
			if model != "" && !containsString(c.Models, model) {
				continue
			}
			filtered = append(filtered, c)
		}
		caps = filtered
	}
	s.writeJSON(w, 200, APIResponse{OK: true, Data: caps})
}

// POST /api/v1/capabilities/announce
func (s *APIServer) handleCapabilityAnnounce(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.writeJSON(w, 405, APIResponse{OK: false, Message: "method not allowed"})
		return
	}
	var entry CapabilityEntry
	if err := json.NewDecoder(r.Body).Decode(&entry); err != nil {
		s.writeJSON(w, 400, APIResponse{OK: false, Message: "invalid JSON: " + err.Error()})
		return
	}
	if entry.Author == "" {
		s.writeJSON(w, 400, APIResponse{OK: false, Message: "author required"})
		return
	}
	entry.UpdatedAt = time.Now().UTC()
	s.capabilities.Update(entry)
	s.writeJSON(w, 200, APIResponse{OK: true, Message: "capability registered"})
}

// POST /api/v1/subscribe — add/remove/list subscription topics
func (s *APIServer) handleSubscribe(w http.ResponseWriter, r *http.Request) {
	rulesPath := filepath.Join(s.store.Root, "subscriptions.json")

	if r.Method == http.MethodGet {
		rules, _ := LoadSyncSubscriptions(rulesPath)
		s.writeJSON(w, 200, APIResponse{OK: true, Data: rules})
		return
	}
	if r.Method != http.MethodPost {
		s.writeJSON(w, 405, APIResponse{OK: false, Message: "method not allowed"})
		return
	}

	var req struct {
		Action   string   `json:"action"`   // "add" or "remove"
		Topics   []string `json:"topics"`
		Channels []string `json:"channels"`
		Tags     []string `json:"tags"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeJSON(w, 400, APIResponse{OK: false, Message: "invalid JSON: " + err.Error()})
		return
	}

	rules, _ := LoadSyncSubscriptions(rulesPath)

	switch strings.ToLower(req.Action) {
	case "add", "":
		rules.Topics = uniqueFold(append(rules.Topics, req.Topics...))
		rules.Channels = uniqueFold(append(rules.Channels, req.Channels...))
		rules.Tags = uniqueFold(append(rules.Tags, req.Tags...))
	case "remove":
		rules.Topics = removeFold(rules.Topics, req.Topics)
		rules.Channels = removeFold(rules.Channels, req.Channels)
		rules.Tags = removeFold(rules.Tags, req.Tags)
	default:
		s.writeJSON(w, 400, APIResponse{OK: false, Message: "action must be 'add' or 'remove'"})
		return
	}

	data, _ := json.MarshalIndent(rules, "", "  ")
	if err := os.WriteFile(rulesPath, data, 0o644); err != nil {
		s.writeJSON(w, 500, APIResponse{OK: false, Message: err.Error()})
		return
	}
	s.writeJSON(w, 200, APIResponse{OK: true, Data: rules})
}

// removeFold removes items from slice (case-insensitive)
func removeFold(slice []string, remove []string) []string {
	removeSet := make(map[string]struct{}, len(remove))
	for _, r := range remove {
		removeSet[strings.ToLower(strings.TrimSpace(r))] = struct{}{}
	}
	var out []string
	for _, s := range slice {
		if _, ok := removeSet[strings.ToLower(strings.TrimSpace(s))]; !ok {
			out = append(out, s)
		}
	}
	return out
}

func containsString(slice []string, target string) bool {
	target = strings.ToLower(strings.TrimSpace(target))
	for _, s := range slice {
		if strings.ToLower(strings.TrimSpace(s)) == target {
			return true
		}
	}
	return false
}

// buildTorrentMap scans the torrent directory and builds a map from
// bundle directory name → infohash by reading each .torrent file's Name field.
func (s *APIServer) buildTorrentMap() map[string]string {
	result := make(map[string]string)
	entries, err := os.ReadDir(s.store.TorrentDir)
	if err != nil {
		return result
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".torrent") {
			continue
		}
		infohash := strings.TrimSuffix(e.Name(), ".torrent")
		torrentPath := filepath.Join(s.store.TorrentDir, e.Name())
		data, err := os.ReadFile(torrentPath)
		if err != nil {
			continue
		}
		// Extract the Name field from torrent info — it matches the bundle dir name
		var mi struct {
			Info struct {
				Name string `bencode:"name"`
			} `bencode:"info"`
		}
		if err := bencode.Unmarshal(data, &mi); err != nil {
			continue
		}
		if mi.Info.Name != "" {
			result[mi.Info.Name] = infohash
		}
	}
	return result
}
