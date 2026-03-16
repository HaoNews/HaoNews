package newsplugin

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"aip2p.org/internal/apphost"
)

//go:embed web/templates/*.html web/static/*
var webFS embed.FS

type App struct {
	storeRoot  string
	project    string
	version    string
	archive    string
	rulesPath  string
	writerPath string
	netPath    string
	listenAddr string
	templates  *template.Template
	staticFS   fs.FS
	loadIndex  func(storeRoot, project string) (Index, error)
	syncIndex  func(index *Index, archiveRoot string) error
	loadRules  func(path string) (SubscriptionRules, error)
	loadWriter func(path string) (WriterPolicy, error)
	loadNet    func(path string) (NetworkBootstrapConfig, error)
	loadSync   func(storeRoot string) (SyncRuntimeStatus, error)
	loadSuper  func(path string) (SyncSupervisorState, error)
	fetchLANBT func(ctx context.Context, value, expectedNetworkID string) (NetworkBootstrapResponse, error)
	options    AppOptions
}

type AppOptions struct {
	ContentRoutes      bool
	ContentAPIRoutes   bool
	ArchiveRoutes      bool
	HistoryAPIRoutes   bool
	NetworkRoutes      bool
	NetworkAPIRoutes   bool
	WriterPolicyRoutes bool
}

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

func displayProjectName(project string) string {
	if strings.EqualFold(strings.TrimSpace(project), "aip2p.news") {
		return "AiP2P News Public"
	}
	return strings.TrimSpace(project)
}

type NavItem struct {
	Name   string
	URL    string
	Active bool
}

type DirectoryItem struct {
	Name          string
	URL           string
	ExternalURL   string
	StoryCount    int
	ReplyCount    int
	ReactionCount int
	AvgTruth      string
}

type NodeStatusEntry struct {
	Label  string
	Value  string
	Detail string
	Tone   string
}

type NodeStatusCard struct {
	Label  string
	Value  string
	Detail string
	Tone   string
}

type NodeStatus struct {
	Summary       string
	SummaryTone   string
	SummaryDetail string
	Entries       []NodeStatusEntry
	Dashboard     []NodeStatusCard
}

type HomePageData struct {
	Project         string
	Version         string
	Posts           []Post
	Now             time.Time
	ListenAddr      string
	AgentView       bool
	ShowNetworkWarn bool
	Options         FeedOptions
	PageNav         []NavItem
	TopicFacets     []FeedFacet
	SourceFacets    []FeedFacet
	SortOptions     []SortOption
	WindowOptions   []TimeWindowOption
	PageSizeOptions []PageSizeOption
	ActiveFilters   []ActiveFilter
	SummaryStats    []SummaryStat
	TotalPostCount  int
	Pagination      PaginationState
	Subscriptions   SubscriptionRules
	NodeStatus      NodeStatus
}

type CollectionPageData struct {
	Project         string
	Version         string
	Kind            string
	Name            string
	Path            string
	DirectoryURL    string
	APIPath         string
	Now             time.Time
	Posts           []Post
	Options         FeedOptions
	PageNav         []NavItem
	SortOptions     []SortOption
	WindowOptions   []TimeWindowOption
	PageSizeOptions []PageSizeOption
	SideLabel       string
	SideFacets      []FeedFacet
	ActiveFilters   []ActiveFilter
	SummaryStats    []SummaryStat
	TotalPostCount  int
	Pagination      PaginationState
	ExternalURL     string
	NodeStatus      NodeStatus
}

type DirectoryPageData struct {
	Project      string
	Version      string
	Kind         string
	Path         string
	APIPath      string
	Now          time.Time
	PageNav      []NavItem
	Items        []DirectoryItem
	SummaryStats []SummaryStat
	NodeStatus   NodeStatus
}

type PostPageData struct {
	Project    string
	Version    string
	PageNav    []NavItem
	Post       Post
	Replies    []Reply
	Reactions  []Reaction
	Related    []Post
	NodeStatus NodeStatus
}

type ArchiveIndexPageData struct {
	Project       string
	Version       string
	PageNav       []NavItem
	Now           time.Time
	Days          []ArchiveDay
	SummaryStats  []SummaryStat
	Subscriptions SubscriptionRules
	NodeStatus    NodeStatus
}

type ArchiveDayPageData struct {
	Project       string
	Version       string
	PageNav       []NavItem
	Now           time.Time
	Day           string
	Days          []ArchiveDay
	Entries       []ArchiveEntry
	SummaryStats  []SummaryStat
	Subscriptions SubscriptionRules
	NodeStatus    NodeStatus
}

type ArchiveMessagePageData struct {
	Project    string
	Version    string
	PageNav    []NavItem
	Now        time.Time
	Entry      ArchiveEntry
	Content    string
	Thread     string
	RawURL     string
	DayURL     string
	Archive    string
	NodeStatus NodeStatus
}

type NetworkPageData struct {
	Project       string
	Version       string
	ListenAddr    string
	PageNav       []NavItem
	Now           time.Time
	NodeStatus    NodeStatus
	SyncStatus    SyncRuntimeStatus
	Supervisor    SyncSupervisorState
	LANPeers      []string
	LANBTAnchors  []LANBTAnchorStatus
	LANBTOverall  string
	LANBTHasMatch bool
}

type LANBTAnchorStatus struct {
	Peer        string
	Nodes       []string
	MatchedNode string
	Error       string
}

type NetworkBootstrapResponse struct {
	Project         string   `json:"project"`
	Version         string   `json:"version"`
	NetworkID       string   `json:"network_id"`
	PeerID          string   `json:"peer_id"`
	ListenAddrs     []string `json:"listen_addrs"`
	DialAddrs       []string `json:"dial_addrs"`
	BitTorrentNodes []string `json:"bittorrent_nodes,omitempty"`
}

func New(storeRoot, project, version, archiveRoot, rulesPath, writerPath, netPath string) (*App, error) {
	return newApp(storeRoot, project, version, archiveRoot, rulesPath, writerPath, netPath, nil, FullAppOptions())
}

func NewWithTheme(storeRoot, project, version, archiveRoot, rulesPath, writerPath, netPath string, theme apphost.WebTheme) (*App, error) {
	return newApp(storeRoot, project, version, archiveRoot, rulesPath, writerPath, netPath, theme, FullAppOptions())
}

func NewWithThemeAndOptions(storeRoot, project, version, archiveRoot, rulesPath, writerPath, netPath string, theme apphost.WebTheme, options AppOptions) (*App, error) {
	return newApp(storeRoot, project, version, archiveRoot, rulesPath, writerPath, netPath, theme, options)
}

func newApp(storeRoot, project, version, archiveRoot, rulesPath, writerPath, netPath string, theme apphost.WebTheme, options AppOptions) (*App, error) {
	if err := ensureRuntimeLayout(storeRoot, archiveRoot, rulesPath, writerPath, netPath); err != nil {
		return nil, err
	}
	funcs := template.FuncMap{
		"formatTime": func(t time.Time) string { return t.Format("2006-01-02 15:04 MST") },
		"formatOptionalTime": func(t *time.Time) string {
			if t == nil {
				return "none yet"
			}
			return t.Format("2006-01-02 15:04 MST")
		},
		"formatScore": func(value *float64) string {
			if value == nil {
				return "-"
			}
			return strings.TrimRight(strings.TrimRight(strconv.FormatFloat(*value, 'f', 2, 64), "0"), ".")
		},
		"compactIdentity": compactIdentity,
		"isPublicKeyish":  isPublicKeyish,
		"displayArchivePath": func(value string) string {
			value = filepath.ToSlash(strings.TrimSpace(value))
			if value == "" {
				return ""
			}
			if idx := strings.Index(value, "/archive/"); idx >= 0 {
				return strings.TrimPrefix(value[idx+1:], "/")
			}
			if strings.HasPrefix(value, "archive/") {
				return value
			}
			return filepath.Base(value)
		},
		"join":          strings.Join,
		"lower":         strings.ToLower,
		"reactionLabel": reactionLabel,
		"sourcePath":    sourcePath,
		"topicPath":     topicPath,
	}
	tmpl, staticFS, err := loadThemeAssets(theme, funcs)
	if err != nil {
		return nil, err
	}
	return &App{
		storeRoot:  storeRoot,
		project:    project,
		version:    strings.TrimSpace(version),
		archive:    archiveRoot,
		rulesPath:  rulesPath,
		writerPath: writerPath,
		netPath:    netPath,
		templates:  tmpl,
		staticFS:   staticFS,
		loadIndex:  LoadIndex,
		syncIndex:  SyncMarkdownArchive,
		loadRules:  LoadSubscriptionRules,
		loadWriter: LoadWriterPolicy,
		loadNet:    LoadNetworkBootstrapConfig,
		loadSync:   loadSyncRuntimeStatus,
		loadSuper:  loadSyncSupervisorState,
		fetchLANBT: fetchNetworkBootstrapResponse,
		options:    options,
	}, nil
}

func loadThemeAssets(theme apphost.WebTheme, funcs template.FuncMap) (*template.Template, fs.FS, error) {
	if theme != nil {
		tmpl, err := theme.ParseTemplates(funcs)
		if err != nil {
			return nil, nil, err
		}
		staticFS, err := theme.StaticFS()
		if err != nil {
			return nil, nil, err
		}
		return tmpl, staticFS, nil
	}
	tmpl, err := template.New("").Funcs(funcs).ParseFS(webFS, "web/templates/*.html")
	if err != nil {
		return nil, nil, err
	}
	staticFS, err := fs.Sub(webFS, "web/static")
	if err != nil {
		return nil, nil, err
	}
	return tmpl, staticFS, nil
}

func ensureRuntimeLayout(storeRoot, archiveRoot, rulesPath, writerPath, netPath string) error {
	storeRoot = strings.TrimSpace(storeRoot)
	if storeRoot != "" {
		for _, dir := range []string{
			filepath.Join(storeRoot, "data"),
			filepath.Join(storeRoot, "torrents"),
		} {
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return err
			}
		}
	}
	archiveRoot = strings.TrimSpace(archiveRoot)
	if archiveRoot != "" {
		if err := os.MkdirAll(archiveRoot, 0o755); err != nil {
			return err
		}
	}
	runtimeRoot := strings.TrimSpace(filepath.Dir(archiveRoot))
	if runtimeRoot != "" {
		for _, dir := range []string{
			filepath.Join(runtimeRoot, "bin"),
			filepath.Join(runtimeRoot, "identities"),
			filepath.Join(runtimeRoot, "delegations"),
			filepath.Join(runtimeRoot, "revocations"),
		} {
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return err
			}
		}
	}
	rulesPath = strings.TrimSpace(rulesPath)
	if rulesPath != "" {
		if err := os.MkdirAll(filepath.Dir(rulesPath), 0o755); err != nil {
			return err
		}
		if err := ensureFileIfMissing(rulesPath, []byte(defaultSubscriptionsJSON)); err != nil {
			return err
		}
	}
	writerPath = strings.TrimSpace(writerPath)
	if writerPath != "" {
		if err := os.MkdirAll(filepath.Dir(writerPath), 0o755); err != nil {
			return err
		}
		if err := ensureWriterPolicyFile(writerPath); err != nil {
			return err
		}
		if err := ensureFileIfMissing(filepath.Join(filepath.Dir(writerPath), writerWhitelistINFName), []byte(defaultWriterWhitelistINF)); err != nil {
			return err
		}
		if err := ensureFileIfMissing(filepath.Join(filepath.Dir(writerPath), writerBlacklistINFName), []byte(defaultWriterBlacklistINF)); err != nil {
			return err
		}
	}
	netPath = strings.TrimSpace(netPath)
	if netPath != "" {
		if err := os.MkdirAll(filepath.Dir(netPath), 0o755); err != nil {
			return err
		}
		if _, err := os.Stat(netPath); errors.Is(err, os.ErrNotExist) {
			content, err := buildDefaultLatestNetINF()
			if err != nil {
				return err
			}
			if err := ensureFileIfMissing(netPath, []byte(content)); err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
		if err := ensureFileIfMissing(filepath.Join(filepath.Dir(netPath), "Trackerlist.inf"), []byte(defaultTrackerListINF)); err != nil {
			return err
		}
		if err := appendNetworkIDIfMissing(netPath, latestOrgNetworkID); err != nil {
			return err
		}
		if err := appendLANPeerIfMissing(netPath, defaultLANPeer); err != nil {
			return err
		}
		if err := appendLANTorrentPeerIfMissing(netPath, defaultLANPeer); err != nil {
			return err
		}
	}
	return nil
}

func (a *App) ListenAndServe(addr string) error {
	a.listenAddr = addr
	return http.ListenAndServe(addr, a.handler())
}

func (a *App) Handler() http.Handler {
	return a.handler()
}

func (a *App) handler() http.Handler {
	mux := http.NewServeMux()
	staticFS := a.staticFS
	if staticFS == nil {
		var err error
		staticFS, err = fs.Sub(webFS, "web/static")
		if err != nil {
			panic(err)
		}
	}
	a.registerContentRoutes(mux)
	a.registerArchiveRoutes(mux)
	a.registerGovernanceRoutes(mux)
	a.registerOpsRoutes(mux)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))
	return mux
}

func (a *App) registerContentRoutes(mux *http.ServeMux) {
	if !a.options.ContentRoutes {
		return
	}
	mux.HandleFunc("/", a.handleHome)
	mux.HandleFunc("/posts/", a.handlePost)
	mux.HandleFunc("/sources", a.handleSources)
	mux.HandleFunc("/sources/", a.handleSource)
	mux.HandleFunc("/topics", a.handleTopics)
	mux.HandleFunc("/topics/", a.handleTopic)
	if !a.options.ContentAPIRoutes {
		return
	}
	mux.HandleFunc("/api/feed", a.handleAPIFeed)
	mux.HandleFunc("/api/posts/", a.handleAPIPost)
	mux.HandleFunc("/api/torrents/", a.handleAPITorrent)
	mux.HandleFunc("/api/sources", a.handleAPISources)
	mux.HandleFunc("/api/sources/", a.handleAPISource)
	mux.HandleFunc("/api/topics", a.handleAPITopics)
	mux.HandleFunc("/api/topics/", a.handleAPITopic)
}

func (a *App) registerArchiveRoutes(mux *http.ServeMux) {
	if a.options.ArchiveRoutes {
		mux.HandleFunc("/archive", a.handleArchiveIndex)
		mux.HandleFunc("/archive/", a.handleArchiveSubtree)
	}
	if !a.options.HistoryAPIRoutes {
		return
	}
	mux.HandleFunc("/api/history/list", a.handleAPIHistoryList)
	mux.HandleFunc("/api/history/manifest", a.handleAPIHistoryList)
}

func (a *App) registerGovernanceRoutes(mux *http.ServeMux) {
	if !a.options.WriterPolicyRoutes {
		return
	}
	mux.HandleFunc("/writer-policy", a.handleWriterPolicy)
}

func (a *App) registerOpsRoutes(mux *http.ServeMux) {
	if a.options.NetworkRoutes {
		mux.HandleFunc("/network", a.handleNetwork)
	}
	if !a.options.NetworkAPIRoutes {
		return
	}
	mux.HandleFunc("/api/network/bootstrap", a.handleAPINetworkBootstrap)
}

func (a *App) index() (Index, error) {
	index, err := a.loadIndex(a.storeRoot, a.project)
	if err != nil {
		return Index{}, err
	}
	if a.loadRules != nil {
		rules, err := a.loadRules(a.rulesPath)
		if err != nil {
			return Index{}, err
		}
		index = ApplySubscriptionRules(index, a.project, rules)
	}
	index, err = a.governanceIndex(index)
	if err != nil {
		return Index{}, err
	}
	if a.syncIndex != nil {
		if err := a.syncIndex(&index, a.archive); err != nil {
			return Index{}, err
		}
	}
	return index, nil
}

func (a *App) subscriptionRules() (SubscriptionRules, error) {
	if a.loadRules == nil {
		return SubscriptionRules{}, nil
	}
	return a.loadRules(a.rulesPath)
}

func (a *App) httpListenAddr() string {
	if strings.TrimSpace(a.listenAddr) == "" {
		return "0.0.0.0:51818"
	}
	return a.listenAddr
}

func (a *App) pageNav(activePath string) []NavItem {
	items := make([]NavItem, 0, 7)
	if a.options.ContentRoutes {
		items = append(items,
			NavItem{Name: "Feed", URL: "/", Active: activePath == "/"},
			NavItem{Name: "Sources", URL: "/sources", Active: strings.HasPrefix(activePath, "/sources")},
			NavItem{Name: "Topics", URL: "/topics", Active: strings.HasPrefix(activePath, "/topics")},
		)
	}
	if a.options.NetworkRoutes {
		items = append(items, NavItem{Name: "Network", URL: "/network", Active: strings.HasPrefix(activePath, "/network")})
	}
	if a.options.WriterPolicyRoutes {
		items = append(items, NavItem{Name: "Policy", URL: "/writer-policy", Active: strings.HasPrefix(activePath, "/writer-policy")})
	}
	if a.options.ArchiveRoutes {
		items = append(items, NavItem{Name: "Archive", URL: "/archive", Active: strings.HasPrefix(activePath, "/archive")})
	}
	if apiURL := a.primaryAPIURL(); apiURL != "" {
		items = append(items, NavItem{Name: "API", URL: apiURL, Active: strings.HasPrefix(activePath, "/api")})
	}
	return items
}

func (a *App) primaryAPIURL() string {
	switch {
	case a.options.ContentAPIRoutes:
		return "/api/feed"
	case a.options.HistoryAPIRoutes:
		return "/api/history/list"
	case a.options.NetworkAPIRoutes:
		return "/api/network/bootstrap"
	default:
		return ""
	}
}

func buildFeedFacets(stats []FacetStat, opts FeedOptions, basePath, key string, omit ...string) []FeedFacet {
	items := make([]FeedFacet, 0, len(stats)+1)
	items = append(items, FeedFacet{
		Name:   "All",
		Count:  0,
		URL:    pageURL(basePath, opts, key, "", omit...),
		Active: activeFeedValue(opts, key) == "",
	})
	limit := len(stats)
	if limit > 8 {
		limit = 8
	}
	for _, stat := range stats[:limit] {
		items = append(items, FeedFacet{
			Name:   stat.Name,
			Count:  stat.Count,
			URL:    pageURL(basePath, opts, key, stat.Name, omit...),
			Active: strings.EqualFold(activeFeedValue(opts, key), stat.Name),
		})
	}
	return items
}

func buildFacetLinks(stats []FacetStat, opts FeedOptions, basePath, key string, omit ...string) []FeedFacet {
	limit := len(stats)
	if limit > 8 {
		limit = 8
	}
	items := make([]FeedFacet, 0, limit)
	for _, stat := range stats[:limit] {
		items = append(items, FeedFacet{
			Name:   stat.Name,
			Count:  stat.Count,
			URL:    pageURL(basePath, opts, key, stat.Name, omit...),
			Active: strings.EqualFold(activeFeedValue(opts, key), stat.Name),
		})
	}
	return items
}

func buildSortOptions(opts FeedOptions, basePath string, omit ...string) []SortOption {
	order := []struct {
		Name  string
		Value string
	}{
		{Name: "Newest", Value: "new"},
		{Name: "Discussed", Value: "discussed"},
		{Name: "Vote Score", Value: "score"},
		{Name: "Truth", Value: "truth"},
		{Name: "Source Quality", Value: "source"},
	}
	items := make([]SortOption, 0, len(order))
	active := opts.Sort
	if active == "" {
		active = "new"
	}
	for _, item := range order {
		items = append(items, SortOption{
			Name:   item.Name,
			Value:  item.Value,
			URL:    pageURL(basePath, opts, "sort", item.Value, omit...),
			Active: item.Value == active,
		})
	}
	return items
}

func buildWindowOptions(opts FeedOptions, basePath string, omit ...string) []TimeWindowOption {
	order := []struct {
		Name  string
		Value string
	}{
		{Name: "All time", Value: ""},
		{Name: "24h", Value: "24h"},
		{Name: "7d", Value: "7d"},
		{Name: "30d", Value: "30d"},
	}
	active := canonicalWindow(opts.Window)
	items := make([]TimeWindowOption, 0, len(order))
	for _, item := range order {
		items = append(items, TimeWindowOption{
			Name:   item.Name,
			Value:  item.Value,
			URL:    pageURL(basePath, opts, "window", item.Value, omit...),
			Active: canonicalWindow(item.Value) == active,
		})
		if item.Value == "" && active == "" {
			items[len(items)-1].Active = true
		}
	}
	return items
}

func buildPageSizeOptions(opts FeedOptions, basePath string, omit ...string) []PageSizeOption {
	sizes := []int{20, 50, 100}
	items := make([]PageSizeOption, 0, len(sizes))
	active := opts.PageSize
	if active == 0 {
		active = 20
	}
	for _, size := range sizes {
		items = append(items, PageSizeOption{
			Name:   strconv.Itoa(size),
			Value:  size,
			URL:    pageURL(basePath, opts, "page_size", strconv.Itoa(size), omit...),
			Active: size == active,
		})
	}
	return items
}

func pageURL(basePath string, opts FeedOptions, key, value string, omit ...string) string {
	next := withOption(opts, key, value)
	encoded := encodeOptions(next, omit...)
	if encoded == "" {
		return basePath
	}
	return basePath + "?" + encoded
}

func withOption(opts FeedOptions, key, value string) FeedOptions {
	next := opts
	switch key {
	case "channel":
		next.Channel = value
	case "topic":
		next.Topic = value
	case "source":
		next.Source = value
	case "sort":
		next.Sort = value
	case "q":
		next.Query = value
	case "window":
		next.Window = canonicalWindow(value)
	case "page":
		next.Page = parsePositiveInt(value, 1)
	case "page_size":
		next.PageSize = parseFeedPageSize(value)
	}
	if key != "page" {
		next.Page = 1
	}
	return next
}

func encodeOptions(opts FeedOptions, omit ...string) string {
	query := url.Values{}
	ignored := make(map[string]struct{}, len(omit))
	for _, key := range omit {
		ignored[key] = struct{}{}
	}
	set := func(key, value string) {
		if value == "" {
			return
		}
		if _, skip := ignored[key]; skip {
			return
		}
		query.Set(key, value)
	}
	set("channel", opts.Channel)
	set("topic", opts.Topic)
	set("source", opts.Source)
	if opts.Sort != "" && opts.Sort != "new" {
		set("sort", opts.Sort)
	}
	set("q", opts.Query)
	if window := canonicalWindow(opts.Window); window != "" {
		set("window", window)
	}
	if opts.Page > 1 {
		query.Set("page", strconv.Itoa(opts.Page))
	}
	if opts.PageSize > 0 && opts.PageSize != 20 {
		query.Set("page_size", strconv.Itoa(opts.PageSize))
	}
	return query.Encode()
}

func activeFeedValue(opts FeedOptions, key string) string {
	switch key {
	case "channel":
		return opts.Channel
	case "topic":
		return opts.Topic
	case "source":
		return opts.Source
	case "window":
		return canonicalWindow(opts.Window)
	default:
		return ""
	}
}

func buildActiveFilters(opts FeedOptions, basePath string, omit ...string) []ActiveFilter {
	filters := make([]ActiveFilter, 0, 5)
	if opts.Query != "" {
		filters = append(filters, ActiveFilter{
			Label: "Search: " + opts.Query,
			URL:   pageURL(basePath, opts, "q", "", omit...),
		})
	}
	if opts.Window != "" {
		filters = append(filters, ActiveFilter{
			Label: "Window: " + strings.ToUpper(opts.Window),
			URL:   pageURL(basePath, opts, "window", "", omit...),
		})
	}
	if opts.Channel != "" {
		filters = append(filters, ActiveFilter{
			Label: "Channel: " + opts.Channel,
			URL:   pageURL(basePath, opts, "channel", "", omit...),
		})
	}
	if opts.Topic != "" && !contains(omit, "topic") {
		filters = append(filters, ActiveFilter{
			Label: "Topic: " + opts.Topic,
			URL:   pageURL(basePath, opts, "topic", "", omit...),
		})
	}
	if opts.Source != "" && !contains(omit, "source") {
		filters = append(filters, ActiveFilter{
			Label: "Source: " + opts.Source,
			URL:   pageURL(basePath, opts, "source", "", omit...),
		})
	}
	if opts.PageSize > 0 && opts.PageSize != 20 {
		filters = append(filters, ActiveFilter{
			Label: "Per page: " + strconv.Itoa(opts.PageSize),
			URL:   pageURL(basePath, opts, "page_size", "20", omit...),
		})
	}
	return filters
}

func buildSummaryStats(posts []Post) []SummaryStat {
	return []SummaryStat{
		{Label: "Visible stories", Value: strconv.Itoa(len(posts))},
		{Label: "Replies", Value: strconv.Itoa(CountReplies(posts))},
		{Label: "Reactions", Value: strconv.Itoa(CountReactions(posts))},
		{Label: "Avg truth", Value: formatAverageTruth(posts)},
	}
}

func paginatePosts(posts []Post, opts FeedOptions, basePath string) ([]Post, PaginationState) {
	pageSize := opts.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	totalItems := len(posts)
	totalPages := 1
	if totalItems > 0 {
		totalPages = (totalItems + pageSize - 1) / pageSize
	}
	page := opts.Page
	if page < 1 {
		page = 1
	}
	if page > totalPages {
		page = totalPages
	}
	start := 0
	end := totalItems
	fromItem := 0
	toItem := 0
	if totalItems > 0 {
		start = (page - 1) * pageSize
		if start > totalItems {
			start = totalItems
		}
		end = start + pageSize
		if end > totalItems {
			end = totalItems
		}
		fromItem = start + 1
		toItem = end
	}
	currentOpts := opts
	currentOpts.Page = page
	currentOpts.PageSize = pageSize
	state := PaginationState{
		Page:       page,
		PageSize:   pageSize,
		TotalItems: totalItems,
		TotalPages: totalPages,
		FromItem:   fromItem,
		ToItem:     toItem,
	}
	if page > 1 {
		state.PrevURL = pageURL(basePath, currentOpts, "page", strconv.Itoa(page-1))
	}
	if page < totalPages {
		state.NextURL = pageURL(basePath, currentOpts, "page", strconv.Itoa(page+1))
	}
	startPage := page - 2
	if startPage < 1 {
		startPage = 1
	}
	endPage := startPage + 4
	if endPage > totalPages {
		endPage = totalPages
	}
	if endPage-startPage < 4 {
		startPage = endPage - 4
		if startPage < 1 {
			startPage = 1
		}
	}
	for p := startPage; p <= endPage; p++ {
		state.Links = append(state.Links, PaginationLink{
			Label:  strconv.Itoa(p),
			URL:    pageURL(basePath, currentOpts, "page", strconv.Itoa(p)),
			Active: p == page,
		})
	}
	return posts[start:end], state
}

func buildDirectorySummaryStats(stats []FacetStat, posts []Post) []SummaryStat {
	return []SummaryStat{
		{Label: "Tracked groups", Value: strconv.Itoa(len(stats))},
		{Label: "Stories", Value: strconv.Itoa(len(posts))},
		{Label: "Replies", Value: strconv.Itoa(CountReplies(posts))},
		{Label: "Avg truth", Value: formatAverageTruth(posts)},
	}
}

func formatAverageTruth(posts []Post) string {
	var sum float64
	var count int
	for _, post := range posts {
		if post.TruthScoreAverage == nil {
			continue
		}
		sum += *post.TruthScoreAverage
		count++
	}
	if count == 0 {
		return "-"
	}
	return strings.TrimRight(strings.TrimRight(strconv.FormatFloat(sum/float64(count), 'f', 2, 64), "0"), ".")
}

func buildSourceDirectory(index Index) []DirectoryItem {
	items := make([]DirectoryItem, 0, len(index.SourceStats))
	for _, stat := range index.SourceStats {
		posts := index.FilterPosts(FeedOptions{Source: stat.Name, Now: time.Now()})
		items = append(items, DirectoryItem{
			Name:          stat.Name,
			URL:           sourcePath(stat.Name),
			ExternalURL:   sourceURLFromPosts(posts),
			StoryCount:    len(posts),
			ReplyCount:    CountReplies(posts),
			ReactionCount: CountReactions(posts),
			AvgTruth:      formatAverageTruth(posts),
		})
	}
	return items
}

func buildTopicDirectory(index Index) []DirectoryItem {
	items := make([]DirectoryItem, 0, len(index.TopicStats))
	for _, stat := range index.TopicStats {
		posts := index.FilterPosts(FeedOptions{Topic: stat.Name, Now: time.Now()})
		items = append(items, DirectoryItem{
			Name:          stat.Name,
			URL:           topicPath(stat.Name),
			StoryCount:    len(posts),
			ReplyCount:    CountReplies(posts),
			ReactionCount: CountReactions(posts),
			AvgTruth:      formatAverageTruth(posts),
		})
	}
	return items
}

func channelStatsForPosts(posts []Post) []FacetStat {
	counts := make(map[string]int)
	for _, post := range posts {
		if post.ChannelGroup == "" {
			continue
		}
		counts[post.ChannelGroup]++
	}
	return facetStats(counts)
}

func topicStatsForPosts(posts []Post) []FacetStat {
	counts := make(map[string]int)
	for _, post := range posts {
		for _, topic := range post.Topics {
			counts[topic]++
		}
	}
	return facetStats(counts)
}

func sourceStatsForPosts(posts []Post) []FacetStat {
	counts := make(map[string]int)
	for _, post := range posts {
		if !post.HasSourcePage || post.SourceName == "" {
			continue
		}
		counts[post.SourceName]++
	}
	return facetStats(counts)
}

func sourceURLFromPosts(posts []Post) string {
	for _, post := range posts {
		if post.SourceURL != "" {
			return post.SourceURL
		}
	}
	return ""
}

func hasSource(index Index, name string) bool {
	for _, stat := range index.SourceStats {
		if strings.EqualFold(stat.Name, name) {
			return true
		}
	}
	return false
}

func hasTopic(index Index, name string) bool {
	for _, stat := range index.TopicStats {
		if strings.EqualFold(stat.Name, name) {
			return true
		}
	}
	return false
}

func pathValue(prefix, path string) string {
	if !strings.HasPrefix(path, prefix) {
		return ""
	}
	value := strings.TrimPrefix(path, prefix)
	if value == "" || strings.Contains(value, "/") {
		return ""
	}
	decoded, err := url.PathUnescape(value)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(decoded)
}

func sourcePath(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}
	return "/sources/" + url.PathEscape(name)
}

func topicPath(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}
	return "/topics/" + url.PathEscape(name)
}

func contains(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

func buildArchiveDays(index Index) []ArchiveDay {
	dayMap := make(map[string]*ArchiveDay)
	for _, bundle := range index.Bundles {
		day := bundle.CreatedAt.UTC().Format("2006-01-02")
		item := dayMap[day]
		if item == nil {
			item = &ArchiveDay{
				Date: day,
				URL:  "/archive/" + day,
			}
			dayMap[day] = item
		}
		switch bundle.Message.Kind {
		case "post":
			item.StoryCount++
		case "reply":
			item.ReplyCount++
		case "reaction":
			item.ReactionCount++
		}
	}
	out := make([]ArchiveDay, 0, len(dayMap))
	for _, item := range dayMap {
		out = append(out, *item)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Date > out[j].Date
	})
	return out
}

func markArchiveDayActive(days []ArchiveDay, active string) []ArchiveDay {
	out := make([]ArchiveDay, 0, len(days))
	for _, day := range days {
		day.Active = day.Date == active
		out = append(out, day)
	}
	return out
}

func hasArchiveDay(days []ArchiveDay, target string) bool {
	for _, day := range days {
		if day.Date == target {
			return true
		}
	}
	return false
}

func buildArchiveSummaryStats(days []ArchiveDay, bundles int) []SummaryStat {
	return []SummaryStat{
		{Label: "Archive days", Value: strconv.Itoa(len(days))},
		{Label: "Mirrored bundles", Value: strconv.Itoa(bundles)},
	}
}

func buildArchiveDayStats(entries []ArchiveEntry) []SummaryStat {
	stories := 0
	replies := 0
	reactions := 0
	for _, entry := range entries {
		switch entry.Kind {
		case "post":
			stories++
		case "reply":
			replies++
		case "reaction":
			reactions++
		}
	}
	return []SummaryStat{
		{Label: "Entries", Value: strconv.Itoa(len(entries))},
		{Label: "Stories", Value: strconv.Itoa(stories)},
		{Label: "Replies", Value: strconv.Itoa(replies)},
		{Label: "Reactions", Value: strconv.Itoa(reactions)},
	}
}

func buildArchiveEntries(index Index, day string) []ArchiveEntry {
	entries := make([]ArchiveEntry, 0)
	for _, bundle := range index.Bundles {
		if bundle.ArchiveMD == "" || bundle.CreatedAt.UTC().Format("2006-01-02") != day {
			continue
		}
		entries = append(entries, archiveEntry(bundle))
	}
	sort.Slice(entries, func(i, j int) bool {
		if !entries[i].CreatedAt.Equal(entries[j].CreatedAt) {
			return entries[i].CreatedAt.After(entries[j].CreatedAt)
		}
		return entries[i].InfoHash < entries[j].InfoHash
	})
	return entries
}

func findArchiveEntry(index Index, infoHash string) (ArchiveEntry, bool) {
	infoHash = strings.ToLower(infoHash)
	for _, bundle := range index.Bundles {
		if strings.ToLower(bundle.InfoHash) == infoHash && bundle.ArchiveMD != "" {
			return archiveEntry(bundle), true
		}
	}
	return ArchiveEntry{}, false
}

func archiveEntry(bundle Bundle) ArchiveEntry {
	title := strings.TrimSpace(bundle.Message.Title)
	if title == "" {
		title = strings.ToUpper(bundle.Message.Kind) + " " + bundle.InfoHash
	}
	day := bundle.CreatedAt.UTC().Format("2006-01-02")
	return ArchiveEntry{
		InfoHash:   bundle.InfoHash,
		Kind:       bundle.Message.Kind,
		Title:      title,
		Author:     bundle.Message.Author,
		CreatedAt:  bundle.CreatedAt,
		ArchiveMD:  bundle.ArchiveMD,
		Day:        day,
		ThreadURL:  bundleThreadURL(bundle),
		ViewerURL:  "/archive/messages/" + bundle.InfoHash,
		RawURL:     "/archive/raw/" + bundle.InfoHash,
		Channel:    bundle.Message.Channel,
		SourceName: nestedString(bundle.Message.Extensions, "source", "name"),
	}
}

func bundleThreadURL(bundle Bundle) string {
	switch bundle.Message.Kind {
	case "post":
		return "/posts/" + bundle.InfoHash
	case "reply":
		if bundle.Message.ReplyTo != nil && bundle.Message.ReplyTo.InfoHash != "" {
			return "/posts/" + bundle.Message.ReplyTo.InfoHash
		}
	case "reaction":
		if infoHash := nestedString(bundle.Message.Extensions, "subject", "infohash"); infoHash != "" {
			return "/posts/" + infoHash
		}
	}
	return ""
}

func apiOptions(opts FeedOptions) map[string]string {
	result := map[string]string{
		"channel": opts.Channel,
		"topic":   opts.Topic,
		"source":  opts.Source,
		"sort":    opts.Sort,
		"q":       opts.Query,
		"window":  canonicalWindow(opts.Window),
	}
	if opts.Page > 1 {
		result["page"] = strconv.Itoa(opts.Page)
	}
	if opts.PageSize > 0 {
		result["page_size"] = strconv.Itoa(opts.PageSize)
	}
	return result
}

func apiPosts(posts []Post) []map[string]any {
	out := make([]map[string]any, 0, len(posts))
	for _, post := range posts {
		out = append(out, apiPost(post, false))
	}
	return out
}

func apiPost(post Post, withBody bool) map[string]any {
	origin := apiOrigin(post.Message.Origin)
	payload := map[string]any{
		"infohash":             post.InfoHash,
		"magnet":               post.Magnet,
		"archive_md":           post.ArchiveMD,
		"title":                post.Message.Title,
		"author":               post.Message.Author,
		"origin":               origin,
		"origin_signed":        origin != nil,
		"delegation":           apiDelegation(post.Delegation),
		"shared_by_local_node": post.SharedByLocalNode,
		"created_at":           post.CreatedAt.Format(time.RFC3339),
		"channel":              post.Message.Channel,
		"channel_group":        post.ChannelGroup,
		"source_name":          post.SourceName,
		"source_site_name":     post.SourceSiteName,
		"source_url":           post.SourceURL,
		"origin_public_key":    post.OriginPublicKey,
		"topics":               post.Topics,
		"post_type":            post.PostType,
		"summary":              post.Summary,
		"reply_count":          post.ReplyCount,
		"reaction_count":       post.ReactionCount,
		"vote_score":           post.VoteScore,
		"truth_score":          scoreValue(post.TruthScoreAverage),
		"source_quality":       scoreValue(post.SourceScoreAverage),
		"thread_path":          "/posts/" + post.InfoHash,
		"source_path":          sourcePathForPost(post),
		"latest_reaction":      post.LatestReactionAuthor,
		"event_time":           timeValue(post.EventTime),
		"topic_paths":          topicPaths(post.Topics),
		"message_tags":         post.Message.Tags,
		"message_protocol":     post.Message.Protocol,
	}
	if withBody {
		payload["body"] = post.Body
	}
	return payload
}

func sourcePathForPost(post Post) string {
	if !post.HasSourcePage || strings.TrimSpace(post.SourceName) == "" {
		return ""
	}
	return sourcePath(post.SourceName)
}

func apiReplies(replies []Reply) []map[string]any {
	out := make([]map[string]any, 0, len(replies))
	for _, reply := range replies {
		origin := apiOrigin(reply.Message.Origin)
		out = append(out, map[string]any{
			"infohash":             reply.InfoHash,
			"magnet":               reply.Magnet,
			"archive_md":           reply.ArchiveMD,
			"author":               reply.Message.Author,
			"origin":               origin,
			"origin_signed":        origin != nil,
			"delegation":           apiDelegation(reply.Delegation),
			"shared_by_local_node": reply.SharedByLocalNode,
			"created_at":           reply.CreatedAt.Format(time.RFC3339),
			"parent_hash":          reply.ParentInfoHash,
			"body":                 reply.Body,
		})
	}
	return out
}

func apiReactions(reactions []Reaction) []map[string]any {
	out := make([]map[string]any, 0, len(reactions))
	for _, reaction := range reactions {
		origin := apiOrigin(reaction.Message.Origin)
		out = append(out, map[string]any{
			"infohash":             reaction.InfoHash,
			"magnet":               reaction.Magnet,
			"archive_md":           reaction.ArchiveMD,
			"author":               reaction.Message.Author,
			"origin":               origin,
			"origin_signed":        origin != nil,
			"delegation":           apiDelegation(reaction.Delegation),
			"shared_by_local_node": reaction.SharedByLocalNode,
			"created_at":           reaction.CreatedAt.Format(time.RFC3339),
			"subject_hash":         reaction.SubjectInfoHash,
			"reaction_type":        reaction.ReactionType,
			"vote_value":           reaction.VoteValue,
			"score_value":          scoreValue(reaction.ScoreValue),
			"explanation":          reaction.Explanation,
		})
	}
	return out
}

func apiOrigin(origin *MessageOrigin) map[string]any {
	if origin == nil {
		return nil
	}
	return map[string]any{
		"author":     origin.Author,
		"agent_id":   origin.AgentID,
		"key_type":   origin.KeyType,
		"public_key": origin.PublicKey,
		"signature":  origin.Signature,
	}
}

func apiDelegation(info *DelegationInfo) map[string]any {
	if info == nil || !info.Delegated {
		return nil
	}
	return map[string]any{
		"delegated":         true,
		"parent_agent_id":   info.ParentAgentID,
		"parent_key_type":   info.ParentKeyType,
		"parent_public_key": info.ParentPublicKey,
		"scopes":            append([]string(nil), info.Scopes...),
		"created_at":        info.CreatedAt,
		"expires_at":        info.ExpiresAt,
	}
}

func originSummary(origin *MessageOrigin) (author, agentID, keyType, publicKey string, signed bool) {
	if origin == nil {
		return "", "", "", "", false
	}
	return strings.TrimSpace(origin.Author),
		strings.TrimSpace(origin.AgentID),
		strings.TrimSpace(origin.KeyType),
		strings.TrimSpace(origin.PublicKey),
		strings.TrimSpace(origin.Signature) != ""
}

func delegationSummary(info *DelegationInfo) (delegated bool, parentAgentID, parentKeyType, parentPublicKey string) {
	if info == nil || !info.Delegated {
		return false, "", "", ""
	}
	return true,
		strings.TrimSpace(info.ParentAgentID),
		strings.TrimSpace(info.ParentKeyType),
		strings.TrimSpace(info.ParentPublicKey)
}

func delegationDirForWriterPolicy(writerPolicyPath string) string {
	root := strings.TrimSpace(filepath.Dir(strings.TrimSpace(writerPolicyPath)))
	if root == "" || root == "." {
		return ""
	}
	return filepath.Join(root, "delegations")
}

func revocationDirForWriterPolicy(writerPolicyPath string) string {
	root := strings.TrimSpace(filepath.Dir(strings.TrimSpace(writerPolicyPath)))
	if root == "" || root == "." {
		return ""
	}
	return filepath.Join(root, "revocations")
}

func topicPaths(topics []string) map[string]string {
	out := make(map[string]string, len(topics))
	for _, topic := range topics {
		out[topic] = topicPath(topic)
	}
	return out
}

func scoreValue(value *float64) any {
	if value == nil {
		return nil
	}
	return *value
}

func timeValue(value *time.Time) any {
	if value == nil {
		return nil
	}
	return value.Format(time.RFC3339)
}

func requestBootstrapHost(r *http.Request) string {
	host := strings.TrimSpace(r.Host)
	if host == "" {
		return ""
	}
	if value, _, err := net.SplitHostPort(host); err == nil {
		return strings.Trim(value, "[]")
	}
	return strings.Trim(host, "[]")
}

func dialableLibP2PAddrs(status SyncRuntimeStatus, host string) []string {
	peerID := strings.TrimSpace(status.LibP2P.PeerID)
	if peerID == "" {
		return nil
	}
	requestIP := net.ParseIP(strings.TrimSpace(host))
	values := append([]string(nil), status.LibP2P.ListenAddrs...)
	values = append(values, status.LibP2P.ConfiguredListen...)
	out := make([]string, 0, len(values))
	seen := make(map[string]struct{})
	for _, value := range values {
		value = rewriteBootstrapAddrForHost(strings.TrimSpace(value), host)
		if value == "" {
			continue
		}
		if !bootstrapAddrMatchesRequestHost(value, requestIP) {
			continue
		}
		if !strings.Contains(value, "/p2p/") {
			value += "/p2p/" + peerID
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func dialableBitTorrentNodes(status SyncRuntimeStatus, host string) []string {
	values := make([]string, 0, 1+len(status.BitTorrentDHT.ListenAddrs))
	if value := rewriteBitTorrentListenForHost(strings.TrimSpace(status.BitTorrentDHT.ConfiguredListen), host); value != "" {
		values = append(values, value)
	}
	for _, value := range status.BitTorrentDHT.ListenAddrs {
		if value := rewriteBitTorrentListenForHost(strings.TrimSpace(value), host); value != "" {
			values = append(values, value)
		}
	}
	requestIP := net.ParseIP(strings.TrimSpace(host))
	out := make([]string, 0, len(values))
	seen := make(map[string]struct{})
	for _, value := range values {
		if !torrentNodeMatchesRequestHost(value, requestIP) {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func bootstrapAddrMatchesRequestHost(value string, requestIP net.IP) bool {
	if requestIP == nil {
		return true
	}
	addrIP := multiaddrIP(value)
	if addrIP == nil {
		return true
	}
	return addrIP.Equal(requestIP)
}

func torrentNodeMatchesRequestHost(value string, requestIP net.IP) bool {
	if requestIP == nil {
		return true
	}
	host, _, err := net.SplitHostPort(strings.TrimSpace(value))
	if err != nil {
		return false
	}
	host = strings.Trim(host, "[]")
	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}
	return ip.Equal(requestIP)
}

func multiaddrIP(value string) net.IP {
	parts := strings.Split(strings.TrimSpace(value), "/")
	for i := 0; i+1 < len(parts); i++ {
		switch parts[i] {
		case "ip4", "ip6":
			if ip := net.ParseIP(parts[i+1]); ip != nil {
				return ip
			}
		}
	}
	return nil
}

func rewriteBootstrapAddrForHost(value, host string) string {
	host = strings.TrimSpace(host)
	if value == "" || host == "" {
		return value
	}
	ip := net.ParseIP(host)
	switch {
	case ip != nil && ip.To4() != nil:
		if strings.Contains(value, "/ip4/0.0.0.0/") {
			value = strings.Replace(value, "/ip4/0.0.0.0/", "/ip4/"+host+"/", 1)
		}
		if strings.Contains(value, "/ip4/127.0.0.1/") {
			value = strings.Replace(value, "/ip4/127.0.0.1/", "/ip4/"+host+"/", 1)
		}
	case ip != nil:
		if strings.Contains(value, "/ip6/::/") {
			value = strings.Replace(value, "/ip6/::/", "/ip6/"+host+"/", 1)
		}
		if strings.Contains(value, "/ip6/::1/") {
			value = strings.Replace(value, "/ip6/::1/", "/ip6/"+host+"/", 1)
		}
	}
	return value
}

func rewriteBitTorrentListenForHost(value, host string) string {
	value = strings.TrimSpace(value)
	host = strings.TrimSpace(host)
	if value == "" {
		return ""
	}
	listenHost, port, err := net.SplitHostPort(value)
	if err != nil {
		return ""
	}
	switch strings.TrimSpace(listenHost) {
	case "", "0.0.0.0", "::", "[::]", "127.0.0.1", "::1", "[::1]":
		if host == "" {
			return ""
		}
		return net.JoinHostPort(host, port)
	default:
		return net.JoinHostPort(strings.Trim(listenHost, "[]"), port)
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	_ = encoder.Encode(payload)
}

func fetchNetworkBootstrapResponse(ctx context.Context, value, expectedNetworkID string) (NetworkBootstrapResponse, error) {
	endpoint, err := latestLANBootstrapEndpoint(value)
	if err != nil {
		return NetworkBootstrapResponse{}, err
	}
	reqCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, endpoint, nil)
	if err != nil {
		return NetworkBootstrapResponse{}, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return NetworkBootstrapResponse{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return NetworkBootstrapResponse{}, fmt.Errorf("status %d", resp.StatusCode)
	}
	var payload NetworkBootstrapResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return NetworkBootstrapResponse{}, err
	}
	if normalizeNetworkID(expectedNetworkID) != "" && strings.TrimSpace(payload.NetworkID) != "" && payload.NetworkID != expectedNetworkID {
		return NetworkBootstrapResponse{}, fmt.Errorf("network id mismatch")
	}
	return payload, nil
}

func latestLANBootstrapEndpoint(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("empty lan bt peer")
	}
	if !strings.Contains(value, "://") {
		value = "http://" + value
	}
	u, err := url.Parse(value)
	if err != nil {
		return "", err
	}
	host := strings.TrimSpace(u.Host)
	if host == "" {
		host = strings.TrimSpace(u.Path)
		u.Path = ""
	}
	if host == "" {
		return "", fmt.Errorf("missing host")
	}
	if _, _, err := net.SplitHostPort(host); err != nil {
		host = net.JoinHostPort(strings.Trim(host, "[]"), "51818")
	}
	u.Scheme = "http"
	u.Host = host
	u.Path = "/api/network/bootstrap"
	u.RawQuery = ""
	u.Fragment = ""
	return u.String(), nil
}
