package newsplugin

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
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

func TestGovernanceSummaryUsesLoadWriterResult(t *testing.T) {
	t.Parallel()

	app := newTestApp(t, fixtureIndex())
	app.loadWriter = func(string) (WriterPolicy, error) {
		return WriterPolicy{
			SyncMode:              WriterSyncModeWhitelist,
			AllowedAgentIDs:       []string{"agent://writer/1", "agent://writer/2"},
			BlockedAgentIDs:       []string{"agent://blocked/1"},
			SharedRegistries:      []string{"registry-1"},
			TrustedAuthorities:    map[string]string{"authority://main": "abcd"},
			DefaultCapability:     WriterCapabilityReadOnly,
			PublicKeyCapabilities: map[string]WriterCapability{"abcd": WriterCapabilityReadWrite},
		}, nil
	}

	summary := app.governanceSummary()
	if len(summary) == 0 {
		t.Fatal("expected governance summary")
	}
}

func TestGovernanceSummarySurfacesLoadWriterError(t *testing.T) {
	t.Parallel()

	app := newTestApp(t, fixtureIndex())
	app.loadWriter = func(string) (WriterPolicy, error) {
		return WriterPolicy{}, errors.New("load failed")
	}

	summary := app.governanceSummary()
	if len(summary) == 0 || summary[0].Value != "load error" {
		t.Fatalf("unexpected summary = %#v", summary)
	}
}
