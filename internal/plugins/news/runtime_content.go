package newsplugin

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
