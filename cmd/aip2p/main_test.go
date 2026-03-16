package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveCreateTargetUsesPathAsOutput(t *testing.T) {
	name, out, err := resolveCreateTarget("/tmp/demo-app", "")
	if err != nil {
		t.Fatalf("resolveCreateTarget() error = %v", err)
	}
	if name != "demo-app" {
		t.Fatalf("name = %q", name)
	}
	if out != "/tmp/demo-app" {
		t.Fatalf("out = %q", out)
	}
}

func TestResolveCreateTargetUsesExplicitOut(t *testing.T) {
	name, out, err := resolveCreateTarget("/tmp/demo-app", "custom-output")
	if err != nil {
		t.Fatalf("resolveCreateTarget() error = %v", err)
	}
	if name != "demo-app" {
		t.Fatalf("name = %q", name)
	}
	if out != "custom-output" {
		t.Fatalf("out = %q", out)
	}
}

func TestInspectAppDir(t *testing.T) {
	root := t.TempDir()
	writeMainTestFile(t, root, "aip2p.app.json", "{\n  \"id\": \"sample-app\",\n  \"name\": \"Sample App\",\n  \"plugins\": [\"sample-content\"],\n  \"theme\": \"sample-theme\"\n}\n")
	writeMainTestFile(t, root, "aip2p.app.config.json", "{\n  \"project\": \"sample.project\",\n  \"runtime_root\": \"runtime-data\"\n}\n")
	writeMainTestFile(t, root, filepath.Join("plugins", "sample-content", "aip2p.plugin.json"), "{\n  \"id\": \"sample-content\",\n  \"name\": \"Sample Content\",\n  \"base_plugin\": \"news-content\",\n  \"default_theme\": \"sample-theme\"\n}\n")
	writeMainTestFile(t, root, filepath.Join("plugins", "sample-content", "aip2p.plugin.config.json"), "{\n  \"channel\": \"sample-world\"\n}\n")
	writeMainTestFile(t, root, filepath.Join("themes", "sample-theme", "aip2p.theme.json"), "{\n  \"id\": \"sample-theme\",\n  \"name\": \"Sample Theme\",\n  \"supported_plugins\": [\"sample-content\"],\n  \"required_plugins\": [\"sample-content\"]\n}\n")
	writeMainTestFile(t, root, filepath.Join("themes", "sample-theme", "templates", "home.html"), "home\n")
	writeMainTestFile(t, root, filepath.Join("themes", "sample-theme", "templates", "post.html"), "post\n")
	writeMainTestFile(t, root, filepath.Join("themes", "sample-theme", "templates", "directory.html"), "directory\n")
	writeMainTestFile(t, root, filepath.Join("themes", "sample-theme", "templates", "collection.html"), "collection\n")
	writeMainTestFile(t, root, filepath.Join("themes", "sample-theme", "templates", "network.html"), "network\n")
	writeMainTestFile(t, root, filepath.Join("themes", "sample-theme", "templates", "archive_index.html"), "archive-index\n")
	writeMainTestFile(t, root, filepath.Join("themes", "sample-theme", "templates", "archive_day.html"), "archive-day\n")
	writeMainTestFile(t, root, filepath.Join("themes", "sample-theme", "templates", "archive_message.html"), "archive-message\n")
	writeMainTestFile(t, root, filepath.Join("themes", "sample-theme", "templates", "writer_policy.html"), "writer-policy\n")
	writeMainTestFile(t, root, filepath.Join("themes", "sample-theme", "templates", "partials.html"), "{{/* */}}\n")
	writeMainTestFile(t, root, filepath.Join("themes", "sample-theme", "static", "styles.css"), "body{}\n")

	bundle, report, err := inspectAppDir(root, "")
	if err != nil {
		t.Fatalf("inspect app dir: %v", err)
	}
	if bundle.App.ID != "sample-app" {
		t.Fatalf("app id = %q", bundle.App.ID)
	}
	if !report.Valid {
		t.Fatalf("report valid = false")
	}
	if report.Config.Project != "sample.project" {
		t.Fatalf("project = %q", report.Config.Project)
	}
	if len(report.Plugins) != 1 || report.Plugins[0].Base == nil || report.Plugins[0].Base.ID != "news-content" {
		t.Fatalf("plugins = %#v", report.Plugins)
	}
	if got := report.Plugins[0].Config["channel"]; got != "sample-world" {
		t.Fatalf("plugin config = %#v", report.Plugins[0].Config)
	}
}

func writeMainTestFile(t *testing.T, root, rel, content string) {
	t.Helper()
	path := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", rel, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", rel, err)
	}
}
