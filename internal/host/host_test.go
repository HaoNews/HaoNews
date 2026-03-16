package host

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestNewLoadsAppDirTheme(t *testing.T) {
	root := t.TempDir()
	writeHostFile(t, root, "aip2p.app.json", "{\n  \"id\": \"sample-news\",\n  \"name\": \"Sample News\",\n  \"plugins\": [\"news-content\"],\n  \"theme\": \"sample-theme\"\n}\n")
	writeHostFile(t, root, filepath.Join("themes", "sample-theme", "aip2p.theme.json"), "{\n  \"id\": \"sample-theme\",\n  \"name\": \"Sample Theme\",\n  \"supported_plugins\": [\"news-content\"],\n  \"required_plugins\": [\"news-content\"]\n}\n")
	writeHostFile(t, root, filepath.Join("themes", "sample-theme", "templates", "home.html"), "home\n")
	writeHostFile(t, root, filepath.Join("themes", "sample-theme", "templates", "post.html"), "post\n")
	writeHostFile(t, root, filepath.Join("themes", "sample-theme", "templates", "directory.html"), "directory\n")
	writeHostFile(t, root, filepath.Join("themes", "sample-theme", "templates", "collection.html"), "collection\n")
	writeHostFile(t, root, filepath.Join("themes", "sample-theme", "templates", "network.html"), "network\n")
	writeHostFile(t, root, filepath.Join("themes", "sample-theme", "templates", "archive_index.html"), "archive-index\n")
	writeHostFile(t, root, filepath.Join("themes", "sample-theme", "templates", "archive_day.html"), "archive-day\n")
	writeHostFile(t, root, filepath.Join("themes", "sample-theme", "templates", "archive_message.html"), "archive-message\n")
	writeHostFile(t, root, filepath.Join("themes", "sample-theme", "templates", "writer_policy.html"), "writer-policy\n")
	writeHostFile(t, root, filepath.Join("themes", "sample-theme", "templates", "partials.html"), "{{/* */}}\n")
	writeHostFile(t, root, filepath.Join("themes", "sample-theme", "static", "styles.css"), "body{}\n")

	instance, err := New(context.Background(), Config{
		AppDir: root,
	})
	if err != nil {
		t.Fatalf("new host: %v", err)
	}
	if instance.Site().Theme.ID != "sample-theme" {
		t.Fatalf("theme id = %q", instance.Site().Theme.ID)
	}
}

func writeHostFile(t *testing.T, root, rel, content string) {
	t.Helper()
	path := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", rel, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", rel, err)
	}
}
