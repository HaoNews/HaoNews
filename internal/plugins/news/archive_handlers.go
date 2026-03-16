package newsplugin

import (
	"net/http"
	"os"
	"strings"
	"time"
)

func (a *App) handleArchiveIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/archive" {
		http.NotFound(w, r)
		return
	}
	index, err := a.index()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	rules, err := a.subscriptionRules()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	days := buildArchiveDays(index)
	data := ArchiveIndexPageData{
		Project:       displayProjectName(a.project),
		Version:       a.version,
		PageNav:       a.pageNav("/archive"),
		Now:           time.Now(),
		Days:          days,
		SummaryStats:  buildArchiveSummaryStats(days, len(index.Bundles)),
		Subscriptions: rules,
		NodeStatus:    a.nodeStatus(index),
	}
	if err := a.templates.ExecuteTemplate(w, "archive_index.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (a *App) handleArchiveSubtree(w http.ResponseWriter, r *http.Request) {
	switch {
	case strings.HasPrefix(r.URL.Path, "/archive/messages/"):
		a.handleArchiveMessage(w, r)
	case strings.HasPrefix(r.URL.Path, "/archive/raw/"):
		a.handleArchiveRaw(w, r)
	default:
		a.handleArchiveDay(w, r)
	}
}

func (a *App) handleArchiveDay(w http.ResponseWriter, r *http.Request) {
	day := pathValue("/archive/", r.URL.Path)
	if day == "" {
		http.NotFound(w, r)
		return
	}
	index, err := a.index()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	rules, err := a.subscriptionRules()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	days := buildArchiveDays(index)
	if !hasArchiveDay(days, day) {
		http.NotFound(w, r)
		return
	}
	entries := buildArchiveEntries(index, day)
	data := ArchiveDayPageData{
		Project:       displayProjectName(a.project),
		Version:       a.version,
		PageNav:       a.pageNav("/archive"),
		Now:           time.Now(),
		Day:           day,
		Days:          markArchiveDayActive(days, day),
		Entries:       entries,
		SummaryStats:  buildArchiveDayStats(entries),
		Subscriptions: rules,
		NodeStatus:    a.nodeStatus(index),
	}
	if err := a.templates.ExecuteTemplate(w, "archive_day.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (a *App) handleArchiveMessage(w http.ResponseWriter, r *http.Request) {
	infoHash := pathValue("/archive/messages/", r.URL.Path)
	if infoHash == "" {
		http.NotFound(w, r)
		return
	}
	index, err := a.index()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	entry, ok := findArchiveEntry(index, infoHash)
	if !ok {
		http.NotFound(w, r)
		return
	}
	content, err := os.ReadFile(entry.ArchiveMD)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data := ArchiveMessagePageData{
		Project:    displayProjectName(a.project),
		Version:    a.version,
		PageNav:    a.pageNav("/archive"),
		Now:        time.Now(),
		Entry:      entry,
		Content:    string(content),
		Thread:     entry.ThreadURL,
		RawURL:     entry.RawURL,
		DayURL:     "/archive/" + entry.Day,
		Archive:    entry.ArchiveMD,
		NodeStatus: a.nodeStatus(index),
	}
	if err := a.templates.ExecuteTemplate(w, "archive_message.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (a *App) handleArchiveRaw(w http.ResponseWriter, r *http.Request) {
	infoHash := pathValue("/archive/raw/", r.URL.Path)
	if infoHash == "" {
		http.NotFound(w, r)
		return
	}
	index, err := a.index()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	entry, ok := findArchiveEntry(index, infoHash)
	if !ok {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	http.ServeFile(w, r, entry.ArchiveMD)
}

func (a *App) handleAPIHistoryList(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/history/list" && r.URL.Path != "/api/history/manifest" {
		http.NotFound(w, r)
		return
	}
	payload, err := a.latestHistoryListPayload()
	if err != nil {
		if os.IsNotExist(err) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, payload)
}

func (a *App) latestHistoryListPayload() (HistoryManifestAPIResponse, error) {
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
