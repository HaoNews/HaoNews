package newsplugin

import (
	"net/http"
	"strings"
	"time"
)

func (a *App) handleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
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
	opts := readFeedOptions(r)
	allPosts := index.FilterPosts(opts)
	posts, pagination := paginatePosts(allPosts, opts, "/")
	showNetworkWarn := shouldShowNetworkWarning(r)
	if showNetworkWarn {
		http.SetCookie(w, &http.Cookie{
			Name:     "aip2p_news_network_warning_seen",
			Value:    "1",
			Path:     "/",
			MaxAge:   180 * 24 * 60 * 60,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})
	}
	data := HomePageData{
		Project:         displayProjectName(a.project),
		Version:         a.version,
		Posts:           posts,
		Now:             time.Now(),
		ListenAddr:      a.httpListenAddr(),
		AgentView:       isAgentViewer(r),
		ShowNetworkWarn: showNetworkWarn,
		Options:         opts,
		PageNav:         a.pageNav("/"),
		TopicFacets:     buildFeedFacets(index.TopicStats, opts, "/", "topic"),
		SourceFacets:    buildFeedFacets(index.SourceStats, opts, "/", "source"),
		SortOptions:     buildSortOptions(opts, "/"),
		WindowOptions:   buildWindowOptions(opts, "/"),
		PageSizeOptions: buildPageSizeOptions(opts, "/"),
		ActiveFilters:   buildActiveFilters(opts, "/"),
		SummaryStats:    buildSummaryStats(allPosts),
		TotalPostCount:  len(index.Posts),
		Pagination:      pagination,
		Subscriptions:   rules,
		NodeStatus:      a.nodeStatus(index),
	}
	if err := a.templates.ExecuteTemplate(w, "home.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func shouldShowNetworkWarning(r *http.Request) bool {
	if r == nil {
		return true
	}
	cookie, err := r.Cookie("aip2p_news_network_warning_seen")
	if err != nil {
		return true
	}
	return strings.TrimSpace(cookie.Value) == ""
}

func compactIdentity(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if isPublicKeyish(value) {
		if len(value) <= 10 {
			return value
		}
		return value[:10] + "..."
	}
	if len(value) <= 24 {
		return value
	}
	return value[:24] + "..."
}

func isPublicKeyish(value string) bool {
	value = strings.TrimSpace(value)
	if len(value) < 32 {
		return false
	}
	for _, r := range value {
		if (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F') {
			continue
		}
		return false
	}
	return true
}

func isAgentViewer(r *http.Request) bool {
	if r == nil {
		return false
	}
	if value := strings.TrimSpace(r.URL.Query().Get("agent")); value != "" {
		switch strings.ToLower(value) {
		case "1", "true", "yes", "on":
			return true
		case "0", "false", "no", "off":
			return false
		}
	}
	ua := strings.ToLower(strings.TrimSpace(r.UserAgent()))
	if ua == "" {
		return false
	}
	if strings.Contains(ua, "mozilla/") && !strings.Contains(ua, "bot") && !strings.Contains(ua, "agent") {
		return false
	}
	markers := []string{
		"agent",
		"bot",
		"crawler",
		"python",
		"curl",
		"wget",
		"httpie",
		"go-http-client",
		"openai",
		"anthropic",
		"claude",
		"gpt",
		"llm",
	}
	for _, marker := range markers {
		if strings.Contains(ua, marker) {
			return true
		}
	}
	return false
}

func (a *App) handlePost(w http.ResponseWriter, r *http.Request) {
	infoHash := pathValue("/posts/", r.URL.Path)
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
	data := PostPageData{
		Project:    displayProjectName(a.project),
		Version:    a.version,
		PageNav:    a.pageNav("/"),
		Post:       post,
		Replies:    index.RepliesByPost[strings.ToLower(infoHash)],
		Reactions:  index.ReactionsByPost[strings.ToLower(infoHash)],
		Related:    index.RelatedPosts(infoHash, 4),
		NodeStatus: a.nodeStatus(index),
	}
	if err := a.templates.ExecuteTemplate(w, "post.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (a *App) handleSources(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/sources" {
		http.NotFound(w, r)
		return
	}
	index, err := a.index()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data := DirectoryPageData{
		Project:      displayProjectName(a.project),
		Version:      a.version,
		Kind:         "Sources",
		Path:         "/sources",
		APIPath:      "/api/sources",
		Now:          time.Now(),
		PageNav:      a.pageNav("/sources"),
		Items:        buildSourceDirectory(index),
		SummaryStats: buildDirectorySummaryStats(index.SourceStats, index.Posts),
		NodeStatus:   a.nodeStatus(index),
	}
	if err := a.templates.ExecuteTemplate(w, "directory.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (a *App) handleSource(w http.ResponseWriter, r *http.Request) {
	name := pathValue("/sources/", r.URL.Path)
	if name == "" {
		http.NotFound(w, r)
		return
	}
	index, err := a.index()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	opts := readFeedOptions(r)
	opts.Source = name
	allPosts := index.FilterPosts(opts)
	posts, pagination := paginatePosts(allPosts, opts, sourcePath(name))
	if !hasSource(index, name) {
		http.NotFound(w, r)
		return
	}
	fullSet := index.FilterPosts(FeedOptions{Source: name, Now: opts.Now})
	data := CollectionPageData{
		Project:         displayProjectName(a.project),
		Version:         a.version,
		Kind:            "Source",
		Name:            name,
		Path:            sourcePath(name),
		DirectoryURL:    "/sources",
		APIPath:         "/api" + sourcePath(name),
		Now:             time.Now(),
		Posts:           posts,
		Options:         opts,
		PageNav:         a.pageNav("/sources"),
		SortOptions:     buildSortOptions(opts, sourcePath(name), "source"),
		WindowOptions:   buildWindowOptions(opts, sourcePath(name), "source"),
		PageSizeOptions: buildPageSizeOptions(opts, sourcePath(name), "source"),
		SideLabel:       "Topics from this source",
		SideFacets:      buildFacetLinks(topicStatsForPosts(fullSet), opts, sourcePath(name), "topic", "source"),
		ActiveFilters:   buildActiveFilters(opts, sourcePath(name), "source"),
		SummaryStats:    buildSummaryStats(allPosts),
		TotalPostCount:  len(fullSet),
		Pagination:      pagination,
		ExternalURL:     sourceURLFromPosts(fullSet),
		NodeStatus:      a.nodeStatus(index),
	}
	if err := a.templates.ExecuteTemplate(w, "collection.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (a *App) handleTopics(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/topics" {
		http.NotFound(w, r)
		return
	}
	index, err := a.index()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data := DirectoryPageData{
		Project:      displayProjectName(a.project),
		Version:      a.version,
		Kind:         "Topics",
		Path:         "/topics",
		APIPath:      "/api/topics",
		Now:          time.Now(),
		PageNav:      a.pageNav("/topics"),
		Items:        buildTopicDirectory(index),
		SummaryStats: buildDirectorySummaryStats(index.TopicStats, index.Posts),
		NodeStatus:   a.nodeStatus(index),
	}
	if err := a.templates.ExecuteTemplate(w, "directory.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (a *App) handleTopic(w http.ResponseWriter, r *http.Request) {
	name := pathValue("/topics/", r.URL.Path)
	if name == "" {
		http.NotFound(w, r)
		return
	}
	index, err := a.index()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	opts := readFeedOptions(r)
	opts.Topic = name
	allPosts := index.FilterPosts(opts)
	posts, pagination := paginatePosts(allPosts, opts, topicPath(name))
	if !hasTopic(index, name) {
		http.NotFound(w, r)
		return
	}
	fullSet := index.FilterPosts(FeedOptions{Topic: name, Now: opts.Now})
	data := CollectionPageData{
		Project:         displayProjectName(a.project),
		Version:         a.version,
		Kind:            "Topic",
		Name:            name,
		Path:            topicPath(name),
		DirectoryURL:    "/topics",
		APIPath:         "/api" + topicPath(name),
		Now:             time.Now(),
		Posts:           posts,
		Options:         opts,
		PageNav:         a.pageNav("/topics"),
		SortOptions:     buildSortOptions(opts, topicPath(name), "topic"),
		WindowOptions:   buildWindowOptions(opts, topicPath(name), "topic"),
		PageSizeOptions: buildPageSizeOptions(opts, topicPath(name), "topic"),
		SideLabel:       "Sources covering this topic",
		SideFacets:      buildFacetLinks(sourceStatsForPosts(fullSet), opts, topicPath(name), "source", "topic"),
		ActiveFilters:   buildActiveFilters(opts, topicPath(name), "topic"),
		SummaryStats:    buildSummaryStats(allPosts),
		TotalPostCount:  len(fullSet),
		Pagination:      pagination,
		NodeStatus:      a.nodeStatus(index),
	}
	if err := a.templates.ExecuteTemplate(w, "collection.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
