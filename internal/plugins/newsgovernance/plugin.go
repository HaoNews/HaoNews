package newsgovernance

import (
	"context"
	_ "embed"

	"aip2p.org/internal/apphost"
	newsplugin "aip2p.org/internal/plugins/news"
)

type Plugin struct{}

//go:embed aip2p.plugin.json
var pluginManifestJSON []byte

func (Plugin) Manifest() apphost.PluginManifest {
	return apphost.MustLoadPluginManifestJSON(pluginManifestJSON)
}

func (Plugin) Build(_ context.Context, cfg apphost.Config, theme apphost.WebTheme) (*apphost.Site, error) {
	return newsplugin.BuildScopedPluginSite(cfg, theme, Plugin{}.Manifest(), newsplugin.GovernanceOnlyAppOptions())
}
