package newsplugin

import (
	"strings"

	"aip2p.org/internal/apphost"
)

const (
	contentPluginID    = "aip2p-public-content"
	archivePluginID    = "aip2p-public-archive"
	governancePluginID = "aip2p-public-governance"
	opsPluginID        = "aip2p-public-ops"
)

func FullAppOptions() AppOptions {
	return AppOptions{
		ContentRoutes:      true,
		ContentAPIRoutes:   true,
		ArchiveRoutes:      true,
		HistoryAPIRoutes:   true,
		NetworkRoutes:      true,
		NetworkAPIRoutes:   true,
		WriterPolicyRoutes: true,
	}
}

func ContentOnlyAppOptions() AppOptions {
	return AppOptions{
		ContentRoutes:    true,
		ContentAPIRoutes: true,
	}
}

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

func OptionsForPlugins(base AppOptions, cfg apphost.Config) AppOptions {
	out := base
	seen := make(map[string]struct{}, len(cfg.Plugins)+1)
	for _, pluginID := range cfg.Plugins {
		pluginID = strings.ToLower(strings.TrimSpace(pluginID))
		if pluginID != "" {
			seen[pluginID] = struct{}{}
		}
	}
	current := strings.ToLower(strings.TrimSpace(cfg.Plugin))
	if current != "" {
		seen[current] = struct{}{}
	}
	if _, ok := seen[contentPluginID]; ok {
		out.ContentRoutes = true
		out.ContentAPIRoutes = true
	}
	if _, ok := seen[archivePluginID]; ok {
		out.ArchiveRoutes = true
		out.HistoryAPIRoutes = true
	}
	if _, ok := seen[governancePluginID]; ok {
		out.WriterPolicyRoutes = true
	}
	if _, ok := seen[opsPluginID]; ok {
		out.NetworkRoutes = true
		out.NetworkAPIRoutes = true
	}
	return out
}
