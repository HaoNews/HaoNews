package aip2p

import (
	"math"
	"strings"
	"sync"
	"time"
)

// v0.3: peer scoring and exponential backoff reconnection

type PeerScoreSource int

const (
	PeerSourceMDNS      PeerScoreSource = iota // LAN discovery
	PeerSourceDHT                              // DHT routing
	PeerSourceBootstrap                        // bootstrap node
	PeerSourceRelay                            // relay connection
)

const (
	scoreMDNSBonus      = 50
	scoreDHTBonus       = 20
	scoreBootstrapBonus = 30
	scoreRelayPenalty   = -10
	scoreActiveBonus    = 20
	scoreFailPenalty    = -15
	maxBackoffDuration  = 5 * time.Minute
	baseBackoffDuration = 30 * time.Second
)

type PeerScore struct {
	PeerID       string
	Address      string
	Score        int
	Source       PeerScoreSource
	LastSeen     time.Time
	LastFailed   time.Time
	FailCount    int
	Connected    bool
	LatencyMs    int64
}

type PeerScorer struct {
	mu    sync.RWMutex
	peers map[string]*PeerScore
}

func NewPeerScorer() *PeerScorer {
	return &PeerScorer{
		peers: make(map[string]*PeerScore),
	}
}

// RecordSeen updates a peer's score when it's discovered or active
func (s *PeerScorer) RecordSeen(peerID, address string, source PeerScoreSource) {
	peerID = strings.TrimSpace(peerID)
	if peerID == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	p, ok := s.peers[peerID]
	if !ok {
		p = &PeerScore{PeerID: peerID, Address: address}
		s.peers[peerID] = p
	}
	p.LastSeen = time.Now().UTC()
	p.Source = source
	p.Address = address

	switch source {
	case PeerSourceMDNS:
		p.Score += scoreMDNSBonus
	case PeerSourceDHT:
		p.Score += scoreDHTBonus
	case PeerSourceBootstrap:
		p.Score += scoreBootstrapBonus
	case PeerSourceRelay:
		p.Score += scoreRelayPenalty
	}
}

// RecordConnected marks a peer as connected and boosts score
func (s *PeerScorer) RecordConnected(peerID string, latencyMs int64) {
	peerID = strings.TrimSpace(peerID)
	s.mu.Lock()
	defer s.mu.Unlock()

	p, ok := s.peers[peerID]
	if !ok {
		p = &PeerScore{PeerID: peerID}
		s.peers[peerID] = p
	}
	p.Connected = true
	p.LatencyMs = latencyMs
	p.Score += scoreActiveBonus
	p.FailCount = 0
}

// RecordFailed marks a connection failure and applies penalty
func (s *PeerScorer) RecordFailed(peerID string) {
	peerID = strings.TrimSpace(peerID)
	s.mu.Lock()
	defer s.mu.Unlock()

	p, ok := s.peers[peerID]
	if !ok {
		p = &PeerScore{PeerID: peerID}
		s.peers[peerID] = p
	}
	p.Connected = false
	p.LastFailed = time.Now().UTC()
	p.FailCount++
	p.Score += scoreFailPenalty
}

// BackoffDuration returns the exponential backoff duration for a peer
func (s *PeerScorer) BackoffDuration(peerID string) time.Duration {
	peerID = strings.TrimSpace(peerID)
	s.mu.RLock()
	defer s.mu.RUnlock()

	p, ok := s.peers[peerID]
	if !ok || p.FailCount == 0 {
		return 0
	}
	backoff := time.Duration(float64(baseBackoffDuration) * math.Pow(2, float64(p.FailCount-1)))
	if backoff > maxBackoffDuration {
		backoff = maxBackoffDuration
	}
	return backoff
}

// ShouldReconnect checks if enough time has passed since last failure
func (s *PeerScorer) ShouldReconnect(peerID string) bool {
	peerID = strings.TrimSpace(peerID)
	s.mu.RLock()
	defer s.mu.RUnlock()

	p, ok := s.peers[peerID]
	if !ok || p.FailCount == 0 {
		return true
	}
	backoff := time.Duration(float64(baseBackoffDuration) * math.Pow(2, float64(p.FailCount-1)))
	if backoff > maxBackoffDuration {
		backoff = maxBackoffDuration
	}
	return time.Since(p.LastFailed) >= backoff
}

// TopPeers returns peers sorted by score (highest first), up to limit
func (s *PeerScorer) TopPeers(limit int) []PeerScore {
	s.mu.RLock()
	defer s.mu.RUnlock()

	all := make([]PeerScore, 0, len(s.peers))
	for _, p := range s.peers {
		all = append(all, *p)
	}
	// Simple insertion sort (small N)
	for i := 1; i < len(all); i++ {
		for j := i; j > 0 && all[j].Score > all[j-1].Score; j-- {
			all[j], all[j-1] = all[j-1], all[j]
		}
	}
	if limit > 0 && len(all) > limit {
		all = all[:limit]
	}
	return all
}

// Prune removes peers not seen in the given duration
func (s *PeerScorer) Prune(maxAge time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cutoff := time.Now().UTC().Add(-maxAge)
	for id, p := range s.peers {
		if p.LastSeen.Before(cutoff) && !p.Connected {
			delete(s.peers, id)
		}
	}
}
