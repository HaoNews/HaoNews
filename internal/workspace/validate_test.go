package workspace

import (
	"context"
	"html/template"
	"io/fs"
	"net/http"
	"testing"

	"aip2p.org/internal/apphost"
)

func TestValidatePluginManifest(t *testing.T) {
	manifest := apphost.PluginManifest{
		ID:           "sample-content",
		Name:         "Sample Content",
		BasePlugin:   "news-content",
		DefaultTheme: "default-news",
	}
	report, err := ValidatePluginManifest(manifest, stubResolver{})
	if err != nil {
		t.Fatalf("validate plugin manifest: %v", err)
	}
	if report.Base == nil || report.Base.ID != "news-content" {
		t.Fatalf("base manifest = %#v", report.Base)
	}
}

func TestValidateAppBundle(t *testing.T) {
	bundle := AppBundle{
		Root: t.TempDir(),
		App: apphost.AppManifest{
			ID:      "sample-app",
			Name:    "Sample App",
			Plugins: []string{"sample-content"},
			Theme:   "default-news",
		},
		Config: AppConfig{
			Project: "sample.project",
		},
		PluginConfigs: map[string]map[string]any{
			"sample-content": {
				"channel": "sample-world",
			},
		},
		PluginRoots: map[string]string{
			"sample-content": "/tmp/sample-content",
		},
	}
	registry := apphost.NewRegistry()
	registry.MustRegisterPlugin(pluginWithManifest(apphost.PluginManifest{
		ID:           "news-content",
		Name:         "News Content",
		DefaultTheme: "default-news",
	}))
	registry.MustRegisterPlugin(pluginWithManifest(apphost.PluginManifest{
		ID:           "sample-content",
		Name:         "Sample Content",
		BasePlugin:   "news-content",
		DefaultTheme: "default-news",
	}))
	registry.MustRegisterTheme(themeWithManifest(apphost.ThemeManifest{
		ID:               "default-news",
		Name:             "Default News",
		SupportedPlugins: []string{"news-content"},
		RequiredPlugins:  []string{"news-content"},
	}))

	report, err := ValidateAppBundle(bundle, registry, registry)
	if err != nil {
		t.Fatalf("validate app bundle: %v", err)
	}
	if !report.Valid {
		t.Fatalf("valid = false")
	}
	if len(report.Plugins) != 1 || report.Plugins[0].Base == nil || report.Plugins[0].Base.ID != "news-content" {
		t.Fatalf("plugins = %#v", report.Plugins)
	}
	if report.Config.Project != "sample.project" {
		t.Fatalf("project = %q", report.Config.Project)
	}
	if report.Plugins[0].Root != "/tmp/sample-content" {
		t.Fatalf("root = %q", report.Plugins[0].Root)
	}
	if got := report.Plugins[0].Config["channel"]; got != "sample-world" {
		t.Fatalf("config channel = %#v", got)
	}
}

type pluginWithManifest apphost.PluginManifest

func (p pluginWithManifest) Manifest() apphost.PluginManifest {
	return apphost.PluginManifest(p)
}

func (p pluginWithManifest) Build(context.Context, apphost.Config, apphost.WebTheme) (*apphost.Site, error) {
	return &apphost.Site{Handler: http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})}, nil
}

type themeWithManifest apphost.ThemeManifest

func (t themeWithManifest) Manifest() apphost.ThemeManifest {
	return apphost.ThemeManifest(t)
}

func (themeWithManifest) ParseTemplates(template.FuncMap) (*template.Template, error) {
	return template.New("test"), nil
}

func (themeWithManifest) StaticFS() (fs.FS, error) {
	return nil, nil
}
