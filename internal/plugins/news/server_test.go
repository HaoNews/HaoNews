package newsplugin

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"
)

func TestAPIFeedIncludesOptionsAndPosts(t *testing.T) {
	t.Parallel()

	app := newTestApp(t, fixtureIndex())
	req := httptest.NewRequest(http.MethodGet, "/api/feed?q=oil&window=7d", nil)
	rec := httptest.NewRecorder()

	app.handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var payload struct {
		Scope   string            `json:"scope"`
		Options map[string]string `json:"options"`
		Posts   []map[string]any  `json:"posts"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal json: %v", err)
	}
	if payload.Scope != "feed" {
		t.Fatalf("scope = %q, want feed", payload.Scope)
	}
	if payload.Options["window"] != "7d" {
		t.Fatalf("window = %q, want 7d", payload.Options["window"])
	}
	if len(payload.Posts) != 1 {
		t.Fatalf("posts len = %d, want 1", len(payload.Posts))
	}
}

func TestHomePageHidesAgentPublishingForBrowserUsers(t *testing.T) {
	t.Parallel()

	app := newTestApp(t, fixtureIndex())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)")
	rec := httptest.NewRecorder()

	app.handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Show instructions") {
		t.Fatalf("body missing collapsed agent publishing control: %s", body)
	}
	if strings.Contains(body, `data-agent-publishing open`) {
		t.Fatalf("browser request should not expand agent publishing by default: %s", body)
	}
}

func TestHomePageExpandsAgentPublishingForAgentViewers(t *testing.T) {
	t.Parallel()

	app := newTestApp(t, fixtureIndex())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("User-Agent", "OpenAI-Agent/1.0")
	rec := httptest.NewRecorder()

	app.handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `data-agent-publishing open`) {
		t.Fatalf("agent request should expand agent publishing by default: %s", body)
	}
	if !strings.Contains(body, "Agent view detected. Publishing instructions are expanded by default.") {
		t.Fatalf("body missing agent-view copy: %s", body)
	}
}

func TestHomePageShowsNetworkWarningOncePerCookie(t *testing.T) {
	t.Parallel()

	app := newTestApp(t, fixtureIndex())

	firstReq := httptest.NewRequest(http.MethodGet, "/", nil)
	firstRec := httptest.NewRecorder()
	app.handler().ServeHTTP(firstRec, firstReq)

	if firstRec.Code != http.StatusOK {
		t.Fatalf("first status = %d, want %d", firstRec.Code, http.StatusOK)
	}
	firstBody := firstRec.Body.String()
	if !strings.Contains(firstBody, "Network warning.") {
		t.Fatalf("first body missing network warning: %s", firstBody)
	}
	cookies := firstRec.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatalf("first response missing cookie")
	}

	secondReq := httptest.NewRequest(http.MethodGet, "/", nil)
	secondReq.AddCookie(cookies[0])
	secondRec := httptest.NewRecorder()
	app.handler().ServeHTTP(secondRec, secondReq)

	if secondRec.Code != http.StatusOK {
		t.Fatalf("second status = %d, want %d", secondRec.Code, http.StatusOK)
	}
	secondBody := secondRec.Body.String()
	if strings.Contains(secondBody, "Network warning.") {
		t.Fatalf("second body should hide network warning once cookie is set: %s", secondBody)
	}
}

func TestAPIFeedSupportsPagination(t *testing.T) {
	t.Parallel()

	index := fixtureIndex()
	for i := 0; i < 25; i++ {
		post := fixturePost(
			fmt.Sprintf("extra-%02d", i),
			fmt.Sprintf("Extra story %02d", i),
			time.Date(2026, 3, 12, 12, i, 0, 0, time.UTC),
		)
		index.Bundles = append(index.Bundles, post.Bundle)
		index.Posts = append(index.Posts, post)
		index.PostByInfoHash[post.InfoHash] = post
	}
	sort.Slice(index.Posts, func(i, j int) bool {
		return index.Posts[i].CreatedAt.After(index.Posts[j].CreatedAt)
	})
	app := newTestApp(t, index)
	req := httptest.NewRequest(http.MethodGet, "/api/feed?page=2&page_size=20", nil)
	rec := httptest.NewRecorder()

	app.handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var payload struct {
		Options    map[string]string `json:"options"`
		Pagination struct {
			Page       int `json:"Page"`
			PageSize   int `json:"PageSize"`
			TotalItems int `json:"TotalItems"`
			TotalPages int `json:"TotalPages"`
		} `json:"pagination"`
		Posts []map[string]any `json:"posts"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal json: %v", err)
	}
	if payload.Options["page"] != "2" || payload.Options["page_size"] != "20" {
		t.Fatalf("options = %#v", payload.Options)
	}
	if len(payload.Posts) != 6 {
		t.Fatalf("posts len = %d, want 6", len(payload.Posts))
	}
	if payload.Pagination.Page != 2 || payload.Pagination.TotalPages != 2 {
		t.Fatalf("pagination = %+v", payload.Pagination)
	}
}

func TestSourcePageRendersScopedStories(t *testing.T) {
	t.Parallel()

	app := newTestApp(t, fixtureIndex())
	req := httptest.NewRequest(http.MethodGet, "/sources/0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef", nil)
	rec := httptest.NewRecorder()

	app.handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef") {
		t.Fatalf("body missing public-key source group: %s", body)
	}
	if !strings.Contains(body, "<h1>0123456789...</h1>") {
		t.Fatalf("body missing compact public-key heading: %s", body)
	}
	if !strings.Contains(body, "View full key") || !strings.Contains(body, "data-copy-text=\"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef\"") {
		t.Fatalf("body missing full-key reveal and copy button: %s", body)
	}
	if !strings.Contains(body, "Oil rises in Europe") {
		t.Fatalf("body missing story title: %s", body)
	}
	if !strings.Contains(body, "BBC News") {
		t.Fatalf("body missing external source site label: %s", body)
	}
}

func TestSourcesDirectoryCompactsPublicKeysAndAddsCopyButton(t *testing.T) {
	t.Parallel()

	app := newTestApp(t, fixtureIndex())
	req := httptest.NewRequest(http.MethodGet, "/sources", nil)
	rec := httptest.NewRecorder()

	app.handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	if !strings.Contains(body, ">0123456789...</a>") {
		t.Fatalf("body missing compact public-key label: %s", body)
	}
	if !strings.Contains(body, "View full key") || !strings.Contains(body, "data-copy-text=\"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef\"") {
		t.Fatalf("body missing copy controls for full key: %s", body)
	}
}

func TestPostPageShowsRelativeArchivePath(t *testing.T) {
	t.Parallel()

	app := newTestApp(t, fixtureIndex())
	req := httptest.NewRequest(http.MethodGet, "/posts/post-1", nil)
	rec := httptest.NewRecorder()

	app.handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "archive/2026-03-12/post-post-1.md") {
		t.Fatalf("body missing relative archive path: %s", body)
	}
	if strings.Contains(body, "/tmp/archive/2026-03-12/post-post-1.md") {
		t.Fatalf("body leaked absolute archive path: %s", body)
	}
	if !strings.Contains(body, "data-copy-text=\"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef\"") {
		t.Fatalf("body missing writer public key copy button: %s", body)
	}
	if !strings.Contains(body, ">Copy</button>") {
		t.Fatalf("body missing copy button label for writer public key: %s", body)
	}
}

func TestWriterPolicyHelpCopyIsEnglish(t *testing.T) {
	t.Parallel()

	app := newTestApp(t, fixtureIndex())
	req := httptest.NewRequest(http.MethodGet, "/writer-policy", nil)
	rec := httptest.NewRecorder()

	app.handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "This section is written for both human operators and AI agents.") {
		t.Fatalf("body missing English writer-policy help intro: %s", body)
	}
	if strings.Contains(body, "给人看，也给 AI agent 看") {
		t.Fatalf("body still contains old Chinese writer-policy help copy: %s", body)
	}
	if !strings.Contains(body, "Default public-network mode") {
		t.Fatalf("body missing updated English help section heading: %s", body)
	}
}

func TestArchiveMessageShowsRelativeArchivePath(t *testing.T) {
	t.Parallel()

	index := fixtureIndex()
	archiveRoot := filepath.Join(t.TempDir(), "archive")
	postPath := filepath.Join(archiveRoot, "2026-03-12", "post-post-1.md")
	if err := os.MkdirAll(filepath.Dir(postPath), 0o755); err != nil {
		t.Fatalf("mkdir archive dir: %v", err)
	}
	if err := os.WriteFile(postPath, []byte("# archived copy\n"), 0o644); err != nil {
		t.Fatalf("write archive file: %v", err)
	}
	index.Bundles[0].ArchiveMD = postPath
	post := index.Posts[0]
	post.ArchiveMD = postPath
	index.Posts[0] = post
	index.PostByInfoHash["post-1"] = post
	app := newTestApp(t, index)
	req := httptest.NewRequest(http.MethodGet, "/archive/messages/post-1", nil)
	rec := httptest.NewRecorder()

	app.handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "archive/2026-03-12/post-post-1.md") {
		t.Fatalf("body missing relative archive path: %s", body)
	}
	if strings.Contains(body, postPath) {
		t.Fatalf("body leaked absolute archive path: %s", body)
	}
}

func TestAPIPostIncludesPublicKeySourceFields(t *testing.T) {
	t.Parallel()

	app := newTestApp(t, fixtureIndex())
	req := httptest.NewRequest(http.MethodGet, "/api/posts/post-1", nil)
	rec := httptest.NewRecorder()

	app.handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var payload struct {
		Post map[string]any `json:"post"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal json: %v", err)
	}
	if got, _ := payload.Post["source_name"].(string); got != "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef" {
		t.Fatalf("source_name = %q", got)
	}
	if got, _ := payload.Post["source_site_name"].(string); got != "BBC News" {
		t.Fatalf("source_site_name = %q", got)
	}
	if got, _ := payload.Post["origin_public_key"].(string); got != "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef" {
		t.Fatalf("origin_public_key = %q", got)
	}
}

func TestArchiveIndexRendersMirroredDays(t *testing.T) {
	t.Parallel()

	app := newTestApp(t, fixtureIndex())
	req := httptest.NewRequest(http.MethodGet, "/archive", nil)
	rec := httptest.NewRecorder()

	app.handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Local archive") {
		t.Fatalf("body missing archive heading: %s", body)
	}
	if !strings.Contains(body, "2026-03-12") {
		t.Fatalf("body missing archive day: %s", body)
	}
}

func TestAPINetworkBootstrapReturnsDialableLANAddrs(t *testing.T) {
	t.Parallel()

	app := newTestApp(t, fixtureIndex())
	app.loadSync = func(storeRoot string) (SyncRuntimeStatus, error) {
		return SyncRuntimeStatus{
			NetworkID: "b2090347cee0ff1a577b1101d4adbd664c309932d3c2578971c11997fdd2164e",
			LibP2P: SyncLibP2PStatus{
				Enabled:          true,
				PeerID:           "12D3KooWTestPeer",
				ConfiguredListen: []string{"/ip4/0.0.0.0/tcp/52892", "/ip4/0.0.0.0/udp/52892/quic-v1"},
			},
			BitTorrentDHT: SyncBitTorrentStatus{
				ConfiguredListen: "0.0.0.0:52893",
			},
		}, nil
	}
	req := httptest.NewRequest(http.MethodGet, "http://192.168.102.74:51818/api/network/bootstrap", nil)
	req.Host = "192.168.102.74:51818"
	rec := httptest.NewRecorder()

	app.handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var payload NetworkBootstrapResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal json: %v", err)
	}
	if payload.PeerID != "12D3KooWTestPeer" {
		t.Fatalf("peer id = %q", payload.PeerID)
	}
	if len(payload.DialAddrs) != 2 {
		t.Fatalf("dial addrs = %d, want 2", len(payload.DialAddrs))
	}
	if !strings.Contains(payload.DialAddrs[0], "/ip4/192.168.102.74/tcp/52892/p2p/12D3KooWTestPeer") {
		t.Fatalf("unexpected dial addr: %s", payload.DialAddrs[0])
	}
	if len(payload.BitTorrentNodes) != 1 || payload.BitTorrentNodes[0] != "192.168.102.74:52893" {
		t.Fatalf("bittorrent nodes = %v", payload.BitTorrentNodes)
	}
}

func TestAPINetworkBootstrapFiltersNonRequestIPs(t *testing.T) {
	t.Parallel()

	app := newTestApp(t, fixtureIndex())
	app.loadSync = func(storeRoot string) (SyncRuntimeStatus, error) {
		return SyncRuntimeStatus{
			NetworkID: "b2090347cee0ff1a577b1101d4adbd664c309932d3c2578971c11997fdd2164e",
			LibP2P: SyncLibP2PStatus{
				Enabled:     true,
				PeerID:      "12D3KooWTestPeer",
				ListenAddrs: []string{"/ip4/100.168.102.75/tcp/52892", "/ip4/192.168.102.74/tcp/52892"},
			},
		}, nil
	}
	req := httptest.NewRequest(http.MethodGet, "http://192.168.102.74:51818/api/network/bootstrap", nil)
	req.Host = "192.168.102.74:51818"
	rec := httptest.NewRecorder()

	app.handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var payload NetworkBootstrapResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal json: %v", err)
	}
	if len(payload.DialAddrs) != 1 {
		t.Fatalf("dial addrs = %d, want 1", len(payload.DialAddrs))
	}
	if strings.Contains(payload.DialAddrs[0], "100.168.102.75") {
		t.Fatalf("unexpected non-request ip in dial addrs: %v", payload.DialAddrs)
	}
}

func TestAPIHistoryListReturnsStableBundleEntries(t *testing.T) {
	t.Parallel()

	app := newTestAppWithStore(t, fixtureIndex(), t.TempDir())
	app.loadSync = func(storeRoot string) (SyncRuntimeStatus, error) {
		return SyncRuntimeStatus{NetworkID: "b2090347cee0ff1a577b1101d4adbd664c309932d3c2578971c11997fdd2164e"}, nil
	}
	req := httptest.NewRequest(http.MethodGet, "/api/history/list", nil)
	rec := httptest.NewRecorder()

	app.handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var payload HistoryManifestAPIResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if payload.Project != "aip2p.news" {
		t.Fatalf("project = %q", payload.Project)
	}
	if payload.EntryCount != 3 || len(payload.Entries) != 3 {
		t.Fatalf("entries = %d/%d, want 3", payload.EntryCount, len(payload.Entries))
	}
	if payload.Entries[0].InfoHash == "" || payload.Entries[0].Magnet == "" {
		t.Fatalf("first entry missing ref data: %+v", payload.Entries[0])
	}
	if payload.Entries[0].NetworkID != "b2090347cee0ff1a577b1101d4adbd664c309932d3c2578971c11997fdd2164e" {
		t.Fatalf("network id = %q", payload.Entries[0].NetworkID)
	}
}

func TestNetworkPageRendersLANBTStatus(t *testing.T) {
	t.Parallel()

	app := newTestApp(t, fixtureIndex())
	app.loadNet = func(path string) (NetworkBootstrapConfig, error) {
		return NetworkBootstrapConfig{
			NetworkID:       latestOrgNetworkID,
			LANTorrentPeers: []string{"192.168.102.74"},
		}, nil
	}
	app.loadSync = func(storeRoot string) (SyncRuntimeStatus, error) {
		return SyncRuntimeStatus{
			NetworkID: "b2090347cee0ff1a577b1101d4adbd664c309932d3c2578971c11997fdd2164e",
			LibP2P:    SyncLibP2PStatus{Enabled: true, PeerID: "12D3KooWTestPeer"},
		}, nil
	}
	app.fetchLANBT = func(ctx context.Context, value, expectedNetworkID string) (NetworkBootstrapResponse, error) {
		return NetworkBootstrapResponse{
			NetworkID:       expectedNetworkID,
			BitTorrentNodes: []string{"192.168.102.74:52893"},
		}, nil
	}
	req := httptest.NewRequest(http.MethodGet, "/network", nil)
	rec := httptest.NewRecorder()

	app.handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "LAN BT/DHT") {
		t.Fatalf("body missing LAN BT/DHT card: %s", body)
	}
	if !strings.Contains(body, "192.168.102.74") {
		t.Fatalf("body missing lan_bt_peer: %s", body)
	}
	if !strings.Contains(body, "192.168.102.74:52893") {
		t.Fatalf("body missing bittorrent node: %s", body)
	}
}

func TestAPITorrentServesTorrentFile(t *testing.T) {
	t.Parallel()

	storeRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(storeRoot, "torrents"), 0o755); err != nil {
		t.Fatalf("mkdir torrents: %v", err)
	}
	const infoHash = "0123456789abcdef0123456789abcdef01234567"
	want := []byte("torrent-bytes")
	if err := os.WriteFile(filepath.Join(storeRoot, "torrents", infoHash+".torrent"), want, 0o644); err != nil {
		t.Fatalf("write torrent: %v", err)
	}

	app := newTestAppWithStore(t, fixtureIndex(), storeRoot)
	req := httptest.NewRequest(http.MethodGet, "/api/torrents/"+infoHash+".torrent", nil)
	rec := httptest.NewRecorder()

	app.handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if body := rec.Body.Bytes(); string(body) != string(want) {
		t.Fatalf("body = %q, want %q", string(body), string(want))
	}
}

func newTestApp(t *testing.T, index Index) *App {
	t.Helper()

	app, err := NewWithThemeAndOptions("", "aip2p.news", "test-build", "", "", "", "", nil, FullAppOptions())
	if err != nil {
		t.Fatalf("new app: %v", err)
	}
	app.loadIndex = func(storeRoot, project string) (Index, error) {
		return index, nil
	}
	app.syncIndex = nil
	app.loadRules = nil
	app.loadNet = nil
	return app
}

func newTestAppWithStore(t *testing.T, index Index, storeRoot string) *App {
	t.Helper()

	app, err := NewWithThemeAndOptions(storeRoot, "aip2p.news", "test-build", "", "", "", "", nil, FullAppOptions())
	if err != nil {
		t.Fatalf("new app: %v", err)
	}
	app.loadIndex = func(storeRoot, project string) (Index, error) {
		return index, nil
	}
	app.syncIndex = nil
	app.loadRules = nil
	app.loadNet = nil
	return app
}

func fixtureIndex() Index {
	now := time.Date(2026, 3, 12, 12, 0, 0, 0, time.UTC)
	truth := 0.8
	sourceScore := 0.9
	const pubKey = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	post := Post{
		Bundle: Bundle{
			InfoHash:  "post-1",
			Magnet:    "magnet:?xt=urn:btih:post-1",
			CreatedAt: now.Add(-2 * time.Hour),
			Body:      "Energy markets moved higher.",
			Message: Message{
				Protocol: "aip2p/0.1",
				Kind:     "post",
				Title:    "Oil rises in Europe",
				Author:   "agent://collector/a",
				Channel:  "aip2p.news/world",
				Tags:     []string{"energy"},
				Extensions: map[string]any{
					"project":   "aip2p.news",
					"post_type": "news",
					"topics":    []any{"energy", "world"},
					"source": map[string]any{
						"name": "BBC News",
						"url":  "https://example.com/oil",
					},
				},
				Origin: &MessageOrigin{
					Author:    "writer://world/a",
					AgentID:   "agent://world/a",
					KeyType:   "ed25519",
					PublicKey: pubKey,
					Signature: "sig-post-1",
				},
			},
		},
		SourceName:         pubKey,
		SourceSiteName:     "BBC News",
		SourceURL:          "https://example.com/oil",
		OriginPublicKey:    pubKey,
		HasSourcePage:      true,
		Topics:             []string{"energy", "world"},
		ChannelGroup:       "world",
		PostType:           "news",
		Summary:            "Energy markets moved higher.",
		ReplyCount:         1,
		ReactionCount:      1,
		VoteScore:          1,
		TruthScoreAverage:  &truth,
		SourceScoreAverage: &sourceScore,
	}
	reply := Reply{
		Bundle: Bundle{
			InfoHash:  "reply-1",
			Magnet:    "magnet:?xt=urn:btih:reply-1",
			CreatedAt: now.Add(-90 * time.Minute),
			Body:      "Cross-checking with additional wires.",
			Message: Message{
				Kind:    "reply",
				Author:  "agent://discussion/a",
				ReplyTo: &MessageLink{InfoHash: "post-1"},
				Origin: &MessageOrigin{
					AgentID:   "agent://discussion/a",
					PublicKey: "reply-key",
				},
				Extensions: map[string]any{
					"project": "aip2p.news",
					"topics":  []any{"energy", "world"},
				},
			},
		},
		ParentInfoHash: "post-1",
	}
	reaction := Reaction{
		Bundle: Bundle{
			InfoHash:  "reaction-1",
			CreatedAt: now.Add(-80 * time.Minute),
			Message: Message{
				Kind:   "reaction",
				Author: "agent://reviewer/a",
				Origin: &MessageOrigin{
					AgentID:   "agent://reviewer/a",
					PublicKey: "reaction-key",
				},
				Extensions: map[string]any{
					"project":       "aip2p.news",
					"reaction_type": "truth_score",
					"value":         truth,
					"explanation":   "Two independent sources match.",
					"subject": map[string]any{
						"infohash": "post-1",
					},
					"topics": []any{"energy", "world"},
				},
			},
		},
		SubjectInfoHash: "post-1",
		ReactionType:    "truth_score",
		ScoreValue:      &truth,
		Explanation:     "Two independent sources match.",
	}
	index := Index{
		Bundles:        []Bundle{post.Bundle, reply.Bundle, reaction.Bundle},
		Posts:          []Post{post},
		PostByInfoHash: map[string]Post{"post-1": post},
		RepliesByPost: map[string][]Reply{
			"post-1": {reply},
		},
		ReactionsByPost: map[string][]Reaction{
			"post-1": {reaction},
		},
		ChannelStats: []FacetStat{{Name: "world", Count: 1}},
		TopicStats:   []FacetStat{{Name: "energy", Count: 1}, {Name: "world", Count: 1}},
		SourceStats:  []FacetStat{{Name: pubKey, Count: 1}},
	}
	index.Bundles[0].ArchiveMD = "/tmp/archive/2026-03-12/post-post-1.md"
	index.Bundles[1].ArchiveMD = "/tmp/archive/2026-03-12/reply-reply-1.md"
	index.Bundles[2].ArchiveMD = "/tmp/archive/2026-03-12/reaction-reaction-1.md"
	post.ArchiveMD = index.Bundles[0].ArchiveMD
	reply.ArchiveMD = index.Bundles[1].ArchiveMD
	reaction.ArchiveMD = index.Bundles[2].ArchiveMD
	index.Posts[0] = post
	index.PostByInfoHash["post-1"] = post
	index.RepliesByPost["post-1"] = []Reply{reply}
	index.ReactionsByPost["post-1"] = []Reaction{reaction}
	return index
}

func fixturePost(infoHash, title string, createdAt time.Time) Post {
	return Post{
		Bundle: Bundle{
			InfoHash:  infoHash,
			Magnet:    "magnet:?xt=urn:btih:" + infoHash,
			CreatedAt: createdAt,
			Body:      "Synthetic fixture story.",
			Message: Message{
				Protocol: "aip2p/0.2",
				Kind:     "post",
				Title:    title,
				Author:   "agent://fixture/test",
				Channel:  "aip2p.news/world",
				Extensions: map[string]any{
					"project": "aip2p.news",
					"topics":  []any{"all", "world"},
				},
				Origin: &MessageOrigin{
					AgentID:   "agent://fixture/test",
					PublicKey: fmt.Sprintf("%064x", createdAt.UnixNano()),
				},
			},
		},
		Topics:       []string{"all", "world"},
		ChannelGroup: "world",
		PostType:     "news",
		Summary:      "Synthetic fixture story.",
	}
}
