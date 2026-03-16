package newsplugin

func (a *App) LatestHistoryListPayload() (HistoryManifestAPIResponse, error) {
	return a.latestHistoryListPayload()
}

func BuildArchiveDays(index Index) []ArchiveDay {
	return buildArchiveDays(index)
}

func MarkArchiveDayActive(days []ArchiveDay, active string) []ArchiveDay {
	return markArchiveDayActive(days, active)
}

func HasArchiveDay(days []ArchiveDay, target string) bool {
	return hasArchiveDay(days, target)
}

func BuildArchiveSummaryStats(days []ArchiveDay, bundles int) []SummaryStat {
	return buildArchiveSummaryStats(days, bundles)
}

func BuildArchiveDayStats(entries []ArchiveEntry) []SummaryStat {
	return buildArchiveDayStats(entries)
}

func BuildArchiveEntries(index Index, day string) []ArchiveEntry {
	return buildArchiveEntries(index, day)
}

func FindArchiveEntry(index Index, infoHash string) (ArchiveEntry, bool) {
	return findArchiveEntry(index, infoHash)
}
