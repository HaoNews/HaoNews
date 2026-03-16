package workspace

import (
	"context"
	"html/template"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"aip2p.org/internal/apphost"
)

func TestLoadAppBundle(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "aip2p.app.json", "{\n  \"id\": \"sample-app\",\n  \"name\": \"Sample App\",\n  \"plugins\": [\"news-content\"],\n  \"theme\": \"sample-theme\"\n}\n")
	writeFile(t, root, filepath.Join("themes", "sample-theme", "aip2p.theme.json"), "{\n  \"id\": \"sample-theme\",\n  \"name\": \"Sample Theme\",\n  \"supported_plugins\": [\"news-content\"],\n  \"required_plugins\": [\"news-content\"]\n}\n")
	writeFile(t, root, filepath.Join("themes", "sample-theme", "templates", "home.html"), "home\n")
	writeFile(t, root, filepath.Join("themes", "sample-theme", "static", "styles.css"), "body{}\n")
	writeFile(t, root, filepath.Join("plugins", "sample-plugin", "aip2p.plugin.json"), "{\n  \"id\": \"sample-plugin\",\n  \"name\": \"Sample Plugin\",\n  \"base_plugin\": \"news-content\",\n  \"default_theme\": \"sample-theme\"\n}\n")

	bundle, err := LoadAppBundle(root)
	if err != nil {
		t.Fatalf("load app bundle: %v", err)
	}
	if bundle.App.ID != "sample-app" {
		t.Fatalf("app id = %q", bundle.App.ID)
	}
	if len(bundle.ThemeManifests) != 1 || bundle.ThemeManifests[0].ID != "sample-theme" {
		t.Fatalf("theme manifests = %#v", bundle.ThemeManifests)
	}
	if len(bundle.PluginManifests) != 1 || bundle.PluginManifests[0].ID != "sample-plugin" {
		t.Fatalf("plugin manifests = %#v", bundle.PluginManifests)
	}
}

func TestLoadPlugins(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, filepath.Join("plugins", "sample-plugin", "aip2p.plugin.json"), "{\n  \"id\": \"sample-plugin\",\n  \"name\": \"Sample Plugin\",\n  \"base_plugin\": \"news-content\",\n  \"default_theme\": \"default-news\"\n}\n")

	plugins, manifests, err := LoadPlugins(filepath.Join(root, "plugins"), stubResolver{})
	if err != nil {
		t.Fatalf("load plugins: %v", err)
	}
	if len(plugins) != 1 || len(manifests) != 1 {
		t.Fatalf("plugins/manifests = %d/%d", len(plugins), len(manifests))
	}
	if manifests[0].BasePlugin != "news-content" {
		t.Fatalf("base plugin = %q", manifests[0].BasePlugin)
	}
}

type stubResolver struct{}

func (stubResolver) ResolvePlugin(id string) (apphost.HTTPPlugin, apphost.PluginManifest, error) {
	plugin := stubPlugin{}
	return plugin, plugin.Manifest(), nil
}

type stubPlugin struct{}

func (stubPlugin) Manifest() apphost.PluginManifest {
	return apphost.PluginManifest{ID: "news-content", Name: "News Content", DefaultTheme: "default-news"}
}

func (stubPlugin) Build(context.Context, apphost.Config, apphost.WebTheme) (*apphost.Site, error) {
	return &apphost.Site{Handler: http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})}, nil
}

type stubTheme struct{}

func (stubTheme) Manifest() apphost.ThemeManifest {
	return apphost.ThemeManifest{ID: "default-news", Name: "Default News"}
}

func (stubTheme) ParseTemplates(template.FuncMap) (*template.Template, error) {
	return template.New("test"), nil
}

func (stubTheme) StaticFS() (fs.FS, error) {
	return nil, nil
}

func writeFile(t *testing.T, root, rel, content string) {
	t.Helper()
	path := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", rel, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", rel, err)
	}
}
