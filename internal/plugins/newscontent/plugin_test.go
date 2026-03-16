package newscontent

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"aip2p.org/internal/apphost"
	"aip2p.org/internal/themes/defaultnews"
)

func TestPluginBuildServesHomePage(t *testing.T) {
	t.Parallel()

	site := buildContentSite(t)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	site.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "AiP2P News Public") {
		t.Fatalf("expected home page content, got %q", rec.Body.String())
	}
}

func TestPluginBuildServesFeedAPI(t *testing.T) {
	t.Parallel()

	site := buildContentSite(t)
	req := httptest.NewRequest(http.MethodGet, "/api/feed", nil)
	rec := httptest.NewRecorder()
	site.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "\"scope\": \"feed\"") {
		t.Fatalf("expected feed payload, got %q", rec.Body.String())
	}
}

func buildContentSite(t *testing.T) *apphost.Site {
	t.Helper()

	root := t.TempDir()
	cfg := apphost.Config{
		RuntimeRoot:      filepath.Join(root, "runtime"),
		StoreRoot:        filepath.Join(root, "store"),
		ArchiveRoot:      filepath.Join(root, "archive"),
		RulesPath:        filepath.Join(root, "config", "subscriptions.json"),
		WriterPolicyPath: filepath.Join(root, "config", "writer_policy.json"),
		NetPath:          filepath.Join(root, "config", "aip2p_net.inf"),
		Project:          "aip2p.news",
		Version:          "test",
	}
	site, err := Plugin{}.Build(context.Background(), cfg, defaultnews.Theme{})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	return site
}
