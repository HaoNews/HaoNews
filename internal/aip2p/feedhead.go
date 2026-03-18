package aip2p

import (
	"encoding/json"
	"strings"
	"sync"
	"time"
)

// v0.3: mutable feed-head mechanism
// Each agent maintains a "latest state pointer" - the infohash + timestamp of their most recent bundle.
// Feed-head updates are published to a dedicated topic so subscribers can quickly discover new content.

const (
	FeedHeadKind       = "feed-head-update"
	FeedHeadTopicPrefix = "aip2p/feedhead"
	FeedHeadGlobalTopic = "aip2p/feedheads"
)

type FeedHeadUpdate struct {
	Protocol       string `json:"protocol"`
	Kind           string `json:"kind"`
	Author         string `json:"author"`
	LatestInfoHash string `json:"latest_infohash"`
	LatestCreatedAt string `json:"latest_created_at"`
	Channel        string `json:"channel,omitempty"`
	Sequence       int64  `json:"sequence"`
	NetworkID      string `json:"network_id,omitempty"`
	UpdatedAt      string `json:"updated_at"`
}

type FeedHeadState struct {
	mu    sync.RWMutex
	heads map[string]*FeedHeadUpdate // keyed by author URI
}

func NewFeedHeadState() *FeedHeadState {
	return &FeedHeadState{
		heads: make(map[string]*FeedHeadUpdate),
	}
}

// Update stores a feed-head update, keeping only the latest per author
func (s *FeedHeadState) Update(update FeedHeadUpdate) {
	author := strings.ToLower(strings.TrimSpace(update.Author))
	if author == "" {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	existing, ok := s.heads[author]
	if ok && existing.Sequence >= update.Sequence {
		return // ignore older updates
	}
	s.heads[author] = &update
}

// Get returns the latest feed-head for an author
func (s *FeedHeadState) Get(author string) (*FeedHeadUpdate, bool) {
	author = strings.ToLower(strings.TrimSpace(author))
	s.mu.RLock()
	defer s.mu.RUnlock()
	head, ok := s.heads[author]
	return head, ok
}

// All returns all current feed-heads
func (s *FeedHeadState) All() []FeedHeadUpdate {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]FeedHeadUpdate, 0, len(s.heads))
	for _, head := range s.heads {
		result = append(result, *head)
	}
	return result
}

// BuildFeedHeadUpdate creates a feed-head update from a published message
func BuildFeedHeadUpdate(author, infoHash, createdAt, channel, networkID string, sequence int64) FeedHeadUpdate {
	return FeedHeadUpdate{
		Protocol:        ProtocolVersion,
		Kind:            FeedHeadKind,
		Author:          strings.TrimSpace(author),
		LatestInfoHash:  strings.ToLower(strings.TrimSpace(infoHash)),
		LatestCreatedAt: strings.TrimSpace(createdAt),
		Channel:         strings.TrimSpace(channel),
		Sequence:        sequence,
		NetworkID:       strings.TrimSpace(networkID),
		UpdatedAt:       time.Now().UTC().Format(time.RFC3339),
	}
}

// MarshalFeedHeadUpdate serializes a feed-head update to JSON
func MarshalFeedHeadUpdate(update FeedHeadUpdate) ([]byte, error) {
	return json.Marshal(update)
}

// UnmarshalFeedHeadUpdate deserializes a feed-head update from JSON
func UnmarshalFeedHeadUpdate(data []byte) (FeedHeadUpdate, error) {
	var update FeedHeadUpdate
	err := json.Unmarshal(data, &update)
	return update, err
}

// FeedHeadTopic returns the dedicated topic name for feed-head updates
func FeedHeadTopic(networkID string) string {
	networkID = strings.TrimSpace(networkID)
	if networkID == "" {
		return FeedHeadGlobalTopic
	}
	return FeedHeadTopicPrefix + "/" + networkID[:8]
}
