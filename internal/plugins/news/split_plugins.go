package newsplugin

import "aip2p.org/internal/apphost"

func ArchiveOnlyAppOptions() AppOptions {
	return AppOptions{
		ArchiveRoutes:    true,
		HistoryAPIRoutes: true,
	}
}

func GovernanceOnlyAppOptions() AppOptions {
	return AppOptions{
		WriterPolicyRoutes: true,
	}
}

func OpsOnlyAppOptions() AppOptions {
	return AppOptions{
		NetworkRoutes:    true,
		NetworkAPIRoutes: true,
	}
}

func BuildScopedPluginSite(cfg apphost.Config, theme apphost.WebTheme, manifest apphost.PluginManifest, options AppOptions) (*apphost.Site, error) {
	cfg = ApplyDefaultConfig(cfg)
	app, err := NewWithThemeAndOptions(
		cfg.StoreRoot,
		cfg.Project,
		cfg.Version,
		cfg.ArchiveRoot,
		cfg.RulesPath,
		cfg.WriterPolicyPath,
		cfg.NetPath,
		theme,
		options,
	)
	if err != nil {
		return nil, err
	}
	return &apphost.Site{
		Manifest: manifest,
		Theme:    theme.Manifest(),
		Handler:  app.Handler(),
	}, nil
}
