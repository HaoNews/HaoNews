package scaffold

import (
	"strings"
	"testing"
)

func TestThemeFilesAreRunnableScaffold(t *testing.T) {
	files, err := ThemeFiles("Sample Theme")
	if err != nil {
		t.Fatalf("theme files: %v", err)
	}
	paths := make(map[string]string, len(files))
	for _, file := range files {
		paths[file.Path] = file.Content
	}
	required := []string{
		"aip2p.theme.json",
		"templates/home.html",
		"templates/post.html",
		"templates/directory.html",
		"templates/collection.html",
		"templates/network.html",
		"templates/archive_index.html",
		"templates/archive_day.html",
		"templates/archive_message.html",
		"templates/writer_policy.html",
		"templates/partials.html",
		"static/styles.css",
	}
	for _, path := range required {
		if _, ok := paths[path]; !ok {
			t.Fatalf("missing scaffold file %q", path)
		}
	}
	if got := paths["aip2p.theme.json"]; got == "" || !strings.Contains(got, `"required_plugins": []`) {
		t.Fatalf("theme manifest missing required_plugins: %q", got)
	}
}
