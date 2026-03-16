package newsplugin

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
