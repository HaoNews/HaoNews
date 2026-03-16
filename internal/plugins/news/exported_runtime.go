package newsplugin

import (
	"context"
	"html/template"
	"net/http"
	"path/filepath"
)

func (a *App) Templates() *template.Template {
	return a.templates
}

func (a *App) PageNav(activePath string) []NavItem {
	return a.pageNav(activePath)
}

func (a *App) ProjectName() string {
	return displayProjectName(a.project)
}

func (a *App) ProjectID() string {
	return a.project
}

func (a *App) VersionString() string {
	return a.version
}

func (a *App) StoreRoot() string {
	return a.storeRoot
}

func (a *App) WriterPolicyPath() string {
	return a.writerPath
}

func (a *App) GovernanceSummary() []SummaryStat {
	return a.governanceSummary()
}

func (a *App) Index() (Index, error) {
	return a.index()
}

func (a *App) SubscriptionRules() (SubscriptionRules, error) {
	return a.subscriptionRules()
}

func (a *App) HTTPListenAddr() string {
	return a.httpListenAddr()
}

func (a *App) NodeStatus(index Index) NodeStatus {
	return a.nodeStatus(index)
}

func (a *App) SyncRuntimeStatus() (SyncRuntimeStatus, error) {
	return a.syncRuntimeStatus()
}

func (a *App) SyncSupervisorStatus() (SyncSupervisorState, error) {
	return a.syncSupervisorStatus()
}

func (a *App) NetworkBootstrap() (NetworkBootstrapConfig, error) {
	return a.networkBootstrap()
}

func (a *App) LANBTStatus(ctx context.Context, cfg NetworkBootstrapConfig) ([]LANBTAnchorStatus, bool, string) {
	return a.lanBTStatus(ctx, cfg)
}

func (a *App) LatestHistoryListPayload() (HistoryManifestAPIResponse, error) {
	return a.latestHistoryListPayload()
}

func DefaultWriterPolicy() WriterPolicy {
	return defaultWriterPolicy()
}

func WriterWhitelistPath(writerPolicyPath string) string {
	return filepath.Join(filepath.Dir(writerPolicyPath), writerWhitelistINFName)
}

func WriterBlacklistPath(writerPolicyPath string) string {
	return filepath.Join(filepath.Dir(writerPolicyPath), writerBlacklistINFName)
}

func RequestBootstrapHost(r *http.Request) string {
	return requestBootstrapHost(r)
}

func DialableLibP2PAddrs(status SyncRuntimeStatus, host string) []string {
	return dialableLibP2PAddrs(status, host)
}

func DialableBitTorrentNodes(status SyncRuntimeStatus, host string) []string {
	return dialableBitTorrentNodes(status, host)
}

func WriteJSON(w http.ResponseWriter, status int, payload any) {
	writeJSON(w, status, payload)
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

func BuildFeedFacets(stats []FacetStat, opts FeedOptions, basePath, key string, omit ...string) []FeedFacet {
	return buildFeedFacets(stats, opts, basePath, key, omit...)
}

func BuildFacetLinks(stats []FacetStat, opts FeedOptions, basePath, key string, omit ...string) []FeedFacet {
	return buildFacetLinks(stats, opts, basePath, key, omit...)
}

func BuildSortOptions(opts FeedOptions, basePath string, omit ...string) []SortOption {
	return buildSortOptions(opts, basePath, omit...)
}

func BuildWindowOptions(opts FeedOptions, basePath string, omit ...string) []TimeWindowOption {
	return buildWindowOptions(opts, basePath, omit...)
}

func BuildPageSizeOptions(opts FeedOptions, basePath string, omit ...string) []PageSizeOption {
	return buildPageSizeOptions(opts, basePath, omit...)
}

func BuildActiveFilters(opts FeedOptions, basePath string, omit ...string) []ActiveFilter {
	return buildActiveFilters(opts, basePath, omit...)
}

func BuildSummaryStats(posts []Post) []SummaryStat {
	return buildSummaryStats(posts)
}

func PaginatePosts(posts []Post, opts FeedOptions, basePath string) ([]Post, PaginationState) {
	return paginatePosts(posts, opts, basePath)
}

func BuildDirectorySummaryStats(stats []FacetStat, posts []Post) []SummaryStat {
	return buildDirectorySummaryStats(stats, posts)
}

func BuildSourceDirectory(index Index) []DirectoryItem {
	return buildSourceDirectory(index)
}

func BuildTopicDirectory(index Index) []DirectoryItem {
	return buildTopicDirectory(index)
}

func ChannelStatsForPosts(posts []Post) []FacetStat {
	return channelStatsForPosts(posts)
}

func TopicStatsForPosts(posts []Post) []FacetStat {
	return topicStatsForPosts(posts)
}

func SourceStatsForPosts(posts []Post) []FacetStat {
	return sourceStatsForPosts(posts)
}

func SourceURLFromPosts(posts []Post) string {
	return sourceURLFromPosts(posts)
}

func HasSource(index Index, name string) bool {
	return hasSource(index, name)
}

func HasTopic(index Index, name string) bool {
	return hasTopic(index, name)
}

func PathValue(prefix, path string) string {
	return pathValue(prefix, path)
}

func SourcePath(name string) string {
	return sourcePath(name)
}

func TopicPath(name string) string {
	return topicPath(name)
}

func APIOptions(opts FeedOptions) map[string]string {
	return apiOptions(opts)
}

func APIPosts(posts []Post) []map[string]any {
	return apiPosts(posts)
}

func APIPost(post Post, withBody bool) map[string]any {
	return apiPost(post, withBody)
}

func APIReplies(replies []Reply) []map[string]any {
	return apiReplies(replies)
}

func APIReactions(reactions []Reaction) []map[string]any {
	return apiReactions(reactions)
}
