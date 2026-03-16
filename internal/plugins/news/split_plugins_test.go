package newsplugin

import (
	"html/template"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"testing"

	"aip2p.org/internal/apphost"
)

func TestSplitPluginOptionFactories(t *testing.T) {
	t.Parallel()

	content := ContentOnlyAppOptions()
	if !content.ContentRoutes || !content.ContentAPIRoutes {
		t.Fatalf("content options = %+v", content)
	}
	if content.ArchiveRoutes || content.NetworkRoutes || content.WriterPolicyRoutes {
		t.Fatalf("content options leaked non-content routes = %+v", content)
	}

	archive := ArchiveOnlyAppOptions()
	if !archive.ArchiveRoutes || !archive.HistoryAPIRoutes {
		t.Fatalf("archive options = %+v", archive)
	}
	if archive.ContentRoutes || archive.NetworkRoutes || archive.WriterPolicyRoutes {
		t.Fatalf("archive options leaked unrelated routes = %+v", archive)
	}

	governance := GovernanceOnlyAppOptions()
	if !governance.WriterPolicyRoutes {
		t.Fatalf("governance options = %+v", governance)
	}
	if governance.ContentRoutes || governance.ArchiveRoutes || governance.NetworkRoutes {
		t.Fatalf("governance options leaked unrelated routes = %+v", governance)
	}

	ops := OpsOnlyAppOptions()
	if !ops.NetworkRoutes || !ops.NetworkAPIRoutes {
		t.Fatalf("ops options = %+v", ops)
	}
	if ops.ContentRoutes || ops.ArchiveRoutes || ops.WriterPolicyRoutes {
		t.Fatalf("ops options leaked unrelated routes = %+v", ops)
	}
}

func TestBuildScopedPluginSitePreservesManifestAndTheme(t *testing.T) {
	t.Parallel()

	cfg := ApplyDefaultConfig(apphost.Config{})
	site, err := BuildScopedPluginSite(
		cfg,
		splitTestTheme{},
		apphost.PluginManifest{ID: "scoped-plugin", Name: "Scoped Plugin", DefaultTheme: "test-theme"},
		ContentOnlyAppOptions(),
	)
	if err != nil {
		t.Fatalf("build scoped plugin site: %v", err)
	}
	if site.Manifest.ID != "scoped-plugin" {
		t.Fatalf("manifest id = %q", site.Manifest.ID)
	}
	if site.Theme.ID != "test-theme" {
		t.Fatalf("theme id = %q", site.Theme.ID)
	}
	if site.Handler == nil {
		t.Fatalf("handler is nil")
	}
}

func TestScopedRoutesRegisterOnlySelectedDomains(t *testing.T) {
	t.Parallel()

	app := newTestApp(t, fixtureIndex())
	app.options = ArchiveOnlyAppOptions()
	archiveHandler := app.handler()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/archive", nil)
	archiveHandler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("archive status = %d, want %d", rec.Code, http.StatusOK)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	archiveHandler.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("archive-only home status = %d, want %d", rec.Code, http.StatusNotFound)
	}

	app = newTestApp(t, fixtureIndex())
	app.options = GovernanceOnlyAppOptions()
	governanceHandler := app.handler()

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/writer-policy", nil)
	governanceHandler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("governance status = %d, want %d", rec.Code, http.StatusOK)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/network", nil)
	governanceHandler.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("governance-only network status = %d, want %d", rec.Code, http.StatusNotFound)
	}

	app = newTestApp(t, fixtureIndex())
	app.options = OpsOnlyAppOptions()
	opsHandler := app.handler()

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/network", nil)
	opsHandler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("ops status = %d, want %d", rec.Code, http.StatusOK)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/topics", nil)
	opsHandler.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("ops-only topics status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

type splitTestTheme struct{}

func (splitTestTheme) Manifest() apphost.ThemeManifest {
	return apphost.ThemeManifest{ID: "test-theme", Name: "Test Theme"}
}

func (splitTestTheme) ParseTemplates(template.FuncMap) (*template.Template, error) {
	return template.New("").Parse(`{{define "partials.html"}}{{end}}{{define "home.html"}}home{{end}}{{define "post.html"}}post{{end}}{{define "directory.html"}}directory{{end}}{{define "collection.html"}}collection{{end}}{{define "network.html"}}network{{end}}{{define "archive_index.html"}}archive{{end}}{{define "archive_day.html"}}archive-day{{end}}{{define "archive_message.html"}}archive-message{{end}}{{define "writer_policy.html"}}writer-policy{{end}}`)
}

func (splitTestTheme) StaticFS() (fs.FS, error) {
	return nil, nil
}
