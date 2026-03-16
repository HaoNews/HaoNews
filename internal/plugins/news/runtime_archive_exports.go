package newsplugin

import (
	"os"
	"strings"
	"time"
)

func (a *App) LatestHistoryListPayload() (HistoryManifestAPIResponse, error) {
	index, err := a.index()
	if err != nil {
		return HistoryManifestAPIResponse{}, err
	}
	if len(index.Bundles) == 0 {
		return HistoryManifestAPIResponse{}, os.ErrNotExist
	}
	var networkID string
	if syncStatus, err := a.syncRuntimeStatus(); err == nil {
		networkID = strings.TrimSpace(syncStatus.NetworkID)
	}
	entries := make([]HistoryManifestEntry, 0, len(index.Bundles))
	for _, bundle := range index.Bundles {
		originAuthor, originAgentID, originKeyType, originPublicKey, originSigned := originSummary(bundle.Message.Origin)
		delegated, parentAgentID, parentKeyType, parentPublicKey := delegationSummary(bundle.Delegation)
		entries = append(entries, HistoryManifestEntry{
			Protocol:          "aip2p-sync/0.1",
			InfoHash:          strings.ToLower(strings.TrimSpace(bundle.InfoHash)),
			Magnet:            strings.TrimSpace(bundle.Magnet),
			SizeBytes:         bundle.SizeBytes,
			Kind:              strings.TrimSpace(bundle.Message.Kind),
			Channel:           strings.TrimSpace(bundle.Message.Channel),
			Title:             strings.TrimSpace(bundle.Message.Title),
			Author:            strings.TrimSpace(bundle.Message.Author),
			CreatedAt:         strings.TrimSpace(bundle.Message.CreatedAt),
			Project:           a.project,
			NetworkID:         networkID,
			Topics:            stringSlice(bundle.Message.Extensions["topics"]),
			Tags:              append([]string(nil), bundle.Message.Tags...),
			OriginAuthor:      originAuthor,
			OriginAgentID:     originAgentID,
			OriginKeyType:     originKeyType,
			OriginPublicKey:   originPublicKey,
			OriginSigned:      originSigned,
			Delegated:         delegated,
			ParentAgentID:     parentAgentID,
			ParentKeyType:     parentKeyType,
			ParentPublicKey:   parentPublicKey,
			SharedByLocalNode: bundle.SharedByLocalNode,
		})
	}
	return HistoryManifestAPIResponse{
		Project:     a.project,
		Version:     a.version,
		NetworkID:   networkID,
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		EntryCount:  len(entries),
		Entries:     entries,
	}, nil
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
