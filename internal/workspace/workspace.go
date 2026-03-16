package workspace

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"aip2p.org/internal/apphost"
	"aip2p.org/internal/themes/directorytheme"
)

type AppBundle struct {
	Root            string
	App             apphost.AppManifest
	ThemeManifests  []apphost.ThemeManifest
	PluginManifests []apphost.PluginManifest
	Themes          []directorytheme.Theme
}

func LoadAppBundle(root string) (AppBundle, error) {
	root = strings.TrimSpace(root)
	if root == "" {
		return AppBundle{}, fmt.Errorf("app directory is required")
	}
	root, err := filepath.Abs(root)
	if err != nil {
		return AppBundle{}, err
	}
	data, err := os.ReadFile(filepath.Join(root, "aip2p.app.json"))
	if err != nil {
		return AppBundle{}, err
	}
	app, err := apphost.LoadAppManifestJSON(data)
	if err != nil {
		return AppBundle{}, err
	}
	themes, themeManifests, err := loadThemes(filepath.Join(root, "themes"))
	if err != nil {
		return AppBundle{}, err
	}
	pluginManifests, err := loadPluginManifests(filepath.Join(root, "plugins"))
	if err != nil {
		return AppBundle{}, err
	}
	return AppBundle{
		Root:            root,
		App:             app,
		ThemeManifests:  themeManifests,
		PluginManifests: pluginManifests,
		Themes:          themes,
	}, nil
}

func LoadPluginManifestDir(root string) (apphost.PluginManifest, error) {
	root = strings.TrimSpace(root)
	if root == "" {
		return apphost.PluginManifest{}, fmt.Errorf("plugin directory is required")
	}
	root, err := filepath.Abs(root)
	if err != nil {
		return apphost.PluginManifest{}, err
	}
	data, err := os.ReadFile(filepath.Join(root, "aip2p.plugin.json"))
	if err != nil {
		return apphost.PluginManifest{}, err
	}
	return apphost.LoadPluginManifestJSON(data)
}

func loadThemes(root string) ([]directorytheme.Theme, []apphost.ThemeManifest, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, nil
		}
		return nil, nil, err
	}
	themes := make([]directorytheme.Theme, 0, len(entries))
	manifests := make([]apphost.ThemeManifest, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		theme, err := directorytheme.Load(filepath.Join(root, entry.Name()))
		if err != nil {
			return nil, nil, err
		}
		themes = append(themes, theme)
		manifests = append(manifests, theme.Manifest())
	}
	sort.Slice(manifests, func(i, j int) bool {
		return manifests[i].ID < manifests[j].ID
	})
	sort.Slice(themes, func(i, j int) bool {
		return themes[i].Manifest().ID < themes[j].Manifest().ID
	})
	return themes, manifests, nil
}

func loadPluginManifests(root string) ([]apphost.PluginManifest, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	manifests := make([]apphost.PluginManifest, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		manifest, err := LoadPluginManifestDir(filepath.Join(root, entry.Name()))
		if err != nil {
			return nil, err
		}
		manifests = append(manifests, manifest)
	}
	sort.Slice(manifests, func(i, j int) bool {
		return manifests[i].ID < manifests[j].ID
	})
	return manifests, nil
}
