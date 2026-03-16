package newsplugin

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func (a *App) handleAPIFeed(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/feed" {
		http.NotFound(w, r)
		return
	}
	index, err := a.index()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	opts := readFeedOptions(r)
	allPosts := index.FilterPosts(opts)
	posts, pagination := paginatePosts(allPosts, opts, "/api/feed")
	writeJSON(w, http.StatusOK, map[string]any{
		"project":    a.project,
		"scope":      "feed",
		"options":    apiOptions(opts),
		"summary":    buildSummaryStats(allPosts),
		"pagination": pagination,
		"posts":      apiPosts(posts),
		"facets": map[string]any{
			"channels": index.ChannelStats,
			"topics":   index.TopicStats,
			"sources":  index.SourceStats,
		},
	})
}

func (a *App) handleAPIPost(w http.ResponseWriter, r *http.Request) {
	infoHash := pathValue("/api/posts/", r.URL.Path)
	if infoHash == "" {
		http.NotFound(w, r)
		return
	}
	index, err := a.index()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	post, ok := index.PostByInfoHash[strings.ToLower(infoHash)]
	if !ok {
		http.NotFound(w, r)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"project":   a.project,
		"scope":     "post",
		"post":      apiPost(post, true),
		"replies":   apiReplies(index.RepliesByPost[strings.ToLower(infoHash)]),
		"reactions": apiReactions(index.ReactionsByPost[strings.ToLower(infoHash)]),
		"related":   apiPosts(index.RelatedPosts(infoHash, 4)),
	})
}

func (a *App) handleAPITorrent(w http.ResponseWriter, r *http.Request) {
	infoHash := pathValue("/api/torrents/", r.URL.Path)
	infoHash = strings.TrimSuffix(strings.ToLower(strings.TrimSpace(infoHash)), ".torrent")
	if infoHash == "" {
		http.NotFound(w, r)
		return
	}
	path := filepath.Join(a.storeRoot, "torrents", infoHash+".torrent")
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/x-bittorrent")
	http.ServeFile(w, r, path)
}

func (a *App) handleAPISources(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/sources" {
		http.NotFound(w, r)
		return
	}
	index, err := a.index()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"project": a.project,
		"scope":   "sources",
		"items":   buildSourceDirectory(index),
	})
}

func (a *App) handleAPISource(w http.ResponseWriter, r *http.Request) {
	name := pathValue("/api/sources/", r.URL.Path)
	if name == "" {
		http.NotFound(w, r)
		return
	}
	index, err := a.index()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !hasSource(index, name) {
		http.NotFound(w, r)
		return
	}
	opts := readFeedOptions(r)
	opts.Source = name
	posts := index.FilterPosts(opts)
	fullSet := index.FilterPosts(FeedOptions{Source: name, Now: opts.Now})
	writeJSON(w, http.StatusOK, map[string]any{
		"project": a.project,
		"scope":   "source",
		"name":    name,
		"options": apiOptions(opts),
		"summary": buildSummaryStats(posts),
		"posts":   apiPosts(posts),
		"facets": map[string]any{
			"channels": channelStatsForPosts(fullSet),
			"topics":   topicStatsForPosts(fullSet),
		},
		"source_url": sourceURLFromPosts(fullSet),
	})
}

func (a *App) handleAPITopics(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/topics" {
		http.NotFound(w, r)
		return
	}
	index, err := a.index()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"project": a.project,
		"scope":   "topics",
		"items":   buildTopicDirectory(index),
	})
}

func (a *App) handleAPITopic(w http.ResponseWriter, r *http.Request) {
	name := pathValue("/api/topics/", r.URL.Path)
	if name == "" {
		http.NotFound(w, r)
		return
	}
	index, err := a.index()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !hasTopic(index, name) {
		http.NotFound(w, r)
		return
	}
	opts := readFeedOptions(r)
	opts.Topic = name
	posts := index.FilterPosts(opts)
	fullSet := index.FilterPosts(FeedOptions{Topic: name, Now: opts.Now})
	writeJSON(w, http.StatusOK, map[string]any{
		"project": a.project,
		"scope":   "topic",
		"name":    name,
		"options": apiOptions(opts),
		"summary": buildSummaryStats(posts),
		"posts":   apiPosts(posts),
		"facets": map[string]any{
			"channels": channelStatsForPosts(fullSet),
			"sources":  sourceStatsForPosts(fullSet),
		},
	})
}

func readFeedOptions(r *http.Request) FeedOptions {
	return FeedOptions{
		Channel:  strings.TrimSpace(r.URL.Query().Get("channel")),
		Topic:    strings.TrimSpace(r.URL.Query().Get("topic")),
		Source:   strings.TrimSpace(r.URL.Query().Get("source")),
		Sort:     strings.TrimSpace(r.URL.Query().Get("sort")),
		Query:    strings.TrimSpace(r.URL.Query().Get("q")),
		Window:   canonicalWindow(r.URL.Query().Get("window")),
		Page:     parsePositiveInt(r.URL.Query().Get("page"), 1),
		PageSize: parseFeedPageSize(r.URL.Query().Get("page_size")),
		Now:      time.Now(),
	}
}

func parsePositiveInt(raw string, fallback int) int {
	value, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || value < 1 {
		return fallback
	}
	return value
}

func parseFeedPageSize(raw string) int {
	value := parsePositiveInt(raw, 20)
	if value < 1 {
		return 20
	}
	if value > 200 {
		return 200
	}
	return value
}
