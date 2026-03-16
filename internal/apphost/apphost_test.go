package apphost

import (
	"context"
	"errors"
	"html/template"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"testing"
)

type testTheme struct{}

func (testTheme) Manifest() ThemeManifest {
	return ThemeManifest{ID: "test-theme", Name: "Test Theme", SupportedPlugins: []string{"test-plugin"}}
}

func (testTheme) ParseTemplates(template.FuncMap) (*template.Template, error) {
	return template.New("test"), nil
}

func (testTheme) StaticFS() (fs.FS, error) {
	return nil, nil
}

type testPlugin struct {
	build func(context.Context, Config, WebTheme) (*Site, error)
}

func (p testPlugin) Manifest() PluginManifest {
	return PluginManifest{ID: "test-plugin", Name: "Test Plugin", DefaultTheme: "test-theme"}
}

func (p testPlugin) Build(ctx context.Context, cfg Config, theme WebTheme) (*Site, error) {
	return p.build(ctx, cfg, theme)
}

func TestRegistryWrapsHandlerPanics(t *testing.T) {
	registry := NewRegistry()
	registry.MustRegisterTheme(testTheme{})
	registry.MustRegisterPlugin(testPlugin{
		build: func(context.Context, Config, WebTheme) (*Site, error) {
			return &Site{
				Handler: http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
					panic("boom")
				}),
			}, nil
		},
	})

	site, err := registry.Build(context.Background(), Config{Plugin: "test-plugin"})
	if err != nil {
		t.Fatalf("build site: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	site.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestRegistrySurfacesStartupErrors(t *testing.T) {
	registry := NewRegistry()
	registry.MustRegisterTheme(testTheme{})
	registry.MustRegisterPlugin(testPlugin{
		build: func(context.Context, Config, WebTheme) (*Site, error) {
			return nil, errors.New("init failed")
		},
	})

	_, err := registry.Build(context.Background(), Config{Plugin: "test-plugin"})
	if err == nil || err.Error() != "init failed" {
		t.Fatalf("err = %v, want init failed", err)
	}
}

func TestRegistryBuildsCompositeSite(t *testing.T) {
	registry := NewRegistry()
	registry.MustRegisterTheme(themeWithManifest{
		manifest: ThemeManifest{
			ID:               "test-theme",
			Name:             "Test Theme",
			SupportedPlugins: []string{"first-plugin", "second-plugin"},
		},
	})
	registry.MustRegisterPlugin(namedTestPlugin("first-plugin", func(context.Context, Config, WebTheme) (*Site, error) {
		return &Site{
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/" {
					http.NotFound(w, r)
					return
				}
				_, _ = w.Write([]byte("home"))
			}),
		}, nil
	}))
	registry.MustRegisterPlugin(namedTestPlugin("second-plugin", func(context.Context, Config, WebTheme) (*Site, error) {
		return &Site{
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/network" {
					http.NotFound(w, r)
					return
				}
				_, _ = w.Write([]byte("network"))
			}),
		}, nil
	}))

	site, err := registry.Build(context.Background(), Config{
		Plugins: []string{"first-plugin", "second-plugin"},
		Theme:   "test-theme",
	})
	if err != nil {
		t.Fatalf("build composite site: %v", err)
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/network", nil)
	site.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || rec.Body.String() != "network" {
		t.Fatalf("status/body = %d %q, want 200 %q", rec.Code, rec.Body.String(), "network")
	}
}

func TestRegistryRejectsIncompatibleTheme(t *testing.T) {
	registry := NewRegistry()
	registry.MustRegisterTheme(testTheme{})
	registry.MustRegisterPlugin(testPlugin{
		build: func(context.Context, Config, WebTheme) (*Site, error) {
			return &Site{Handler: http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})}, nil
		},
	})
	registry.MustRegisterTheme(themeWithManifest{
		ThemeManifest{ID: "other-theme", Name: "Other", SupportedPlugins: []string{"another-plugin"}},
	})

	_, err := registry.Build(context.Background(), Config{
		Plugin: "test-plugin",
		Theme:  "other-theme",
	})
	if err == nil || err.Error() != `theme "other-theme" does not support plugin "test-plugin"` {
		t.Fatalf("err = %v", err)
	}
}

func TestRegistryRejectsThemeWithMissingRequiredPlugin(t *testing.T) {
	registry := NewRegistry()
	registry.MustRegisterTheme(themeWithManifest{
		manifest: ThemeManifest{
			ID:               "stacked-theme",
			Name:             "Stacked Theme",
			SupportedPlugins: []string{"test-plugin"},
			RequiredPlugins:  []string{"test-plugin", "extra-plugin"},
		},
	})
	registry.MustRegisterPlugin(testPlugin{
		build: func(context.Context, Config, WebTheme) (*Site, error) {
			return &Site{Handler: http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})}, nil
		},
	})

	_, err := registry.Build(context.Background(), Config{
		Plugin: "test-plugin",
		Theme:  "stacked-theme",
	})
	if err == nil || err.Error() != `theme "stacked-theme" requires plugins: extra-plugin` {
		t.Fatalf("err = %v", err)
	}
}

func TestRegistryAcceptsThemeCompatibilityViaBasePlugin(t *testing.T) {
	registry := NewRegistry()
	registry.MustRegisterTheme(themeWithManifest{
		manifest: ThemeManifest{
			ID:               "default-news",
			Name:             "Default News",
			SupportedPlugins: []string{"news-content"},
			RequiredPlugins:  []string{"news-content"},
		},
	})
	registry.MustRegisterPlugin(pluginWithManifest{
		manifest: PluginManifest{
			ID:           "sample-content",
			Name:         "Sample Content",
			BasePlugin:   "news-content",
			DefaultTheme: "default-news",
		},
		build: func(context.Context, Config, WebTheme) (*Site, error) {
			return &Site{Handler: http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})}, nil
		},
	})

	if _, err := registry.Build(context.Background(), Config{
		Plugin: "sample-content",
		Theme:  "default-news",
	}); err != nil {
		t.Fatalf("build with base-plugin compatibility: %v", err)
	}
}

type themeWithManifest struct {
	manifest ThemeManifest
}

func (t themeWithManifest) Manifest() ThemeManifest {
	return t.manifest
}

func (themeWithManifest) ParseTemplates(template.FuncMap) (*template.Template, error) {
	return template.New("test"), nil
}

func (themeWithManifest) StaticFS() (fs.FS, error) {
	return nil, nil
}

type namedTestPluginValue struct {
	id    string
	build func(context.Context, Config, WebTheme) (*Site, error)
}

func namedTestPlugin(id string, build func(context.Context, Config, WebTheme) (*Site, error)) namedTestPluginValue {
	return namedTestPluginValue{id: id, build: build}
}

func (p namedTestPluginValue) Manifest() PluginManifest {
	return PluginManifest{ID: p.id, Name: p.id, DefaultTheme: "test-theme"}
}

func (p namedTestPluginValue) Build(ctx context.Context, cfg Config, theme WebTheme) (*Site, error) {
	return p.build(ctx, cfg, theme)
}

type pluginWithManifest struct {
	manifest PluginManifest
	build    func(context.Context, Config, WebTheme) (*Site, error)
}

func (p pluginWithManifest) Manifest() PluginManifest {
	return p.manifest
}

func (p pluginWithManifest) Build(ctx context.Context, cfg Config, theme WebTheme) (*Site, error) {
	return p.build(ctx, cfg, theme)
}
