package newsplugin

import (
	"context"
	_ "embed"

	"aip2p.org/internal/apphost"
)

type Plugin struct{}

//go:embed aip2p.plugin.json
var pluginManifestJSON []byte

func (Plugin) Manifest() apphost.PluginManifest {
	return apphost.MustLoadPluginManifestJSON(pluginManifestJSON)
}

func (Plugin) Build(ctx context.Context, cfg apphost.Config, theme apphost.WebTheme) (*apphost.Site, error) {
	cfg = ApplyDefaultConfig(cfg)
	runtime := RuntimePathsFromRoot(cfg.RuntimeRoot)

	app, err := NewWithTheme(cfg.StoreRoot, cfg.Project, cfg.Version, cfg.ArchiveRoot, cfg.RulesPath, cfg.WriterPolicyPath, cfg.NetPath, theme)
	if err != nil {
		return nil, err
	}

	var supervisor *ManagedSyncSupervisor
	syncMode, err := ParseSyncMode(cfg.SyncMode)
	if err != nil {
		return nil, err
	}
	if syncMode == SyncModeManaged {
		supervisor, err = StartManagedSyncSupervisor(ctx, ManagedSyncConfig{
			Runtime:          runtime,
			BinaryPath:       cfg.SyncBinaryPath,
			StoreRoot:        cfg.StoreRoot,
			NetPath:          cfg.NetPath,
			RulesPath:        cfg.RulesPath,
			WriterPolicyPath: cfg.WriterPolicyPath,
			Trackers:         cfg.TrackerPath,
			StaleAfter:       cfg.SyncStaleAfter,
			Logf:             cfg.Logf,
		})
		if err != nil {
			return nil, err
		}
	}

	return &apphost.Site{
		Manifest: Plugin{}.Manifest(),
		Theme:    theme.Manifest(),
		Handler:  app.Handler(),
		Close: func(context.Context) error {
			if supervisor != nil {
				supervisor.Stop()
			}
			return nil
		},
	}, nil
}
