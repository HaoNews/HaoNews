package workspace

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadAppBundle(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "aip2p.app.json", "{\n  \"id\": \"sample-app\",\n  \"name\": \"Sample App\",\n  \"plugins\": [\"news-content\"],\n  \"theme\": \"sample-theme\"\n}\n")
	writeFile(t, root, filepath.Join("themes", "sample-theme", "aip2p.theme.json"), "{\n  \"id\": \"sample-theme\",\n  \"name\": \"Sample Theme\",\n  \"supported_plugins\": [\"news-content\"],\n  \"required_plugins\": [\"news-content\"]\n}\n")
	writeFile(t, root, filepath.Join("themes", "sample-theme", "templates", "home.html"), "home\n")
	writeFile(t, root, filepath.Join("themes", "sample-theme", "static", "styles.css"), "body{}\n")
	writeFile(t, root, filepath.Join("plugins", "sample-plugin", "aip2p.plugin.json"), "{\n  \"id\": \"sample-plugin\",\n  \"name\": \"Sample Plugin\",\n  \"default_theme\": \"sample-theme\"\n}\n")

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
