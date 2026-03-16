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
		"sourcePath":    SourcePath,
		"topicPath":     TopicPath,
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
