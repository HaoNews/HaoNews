package aip2p

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"strings"
)

// v0.4: Optional Hashcash-style Proof-of-Work for anti-spam.
// Messages can include pow.nonce, pow.difficulty, pow.hash in extensions.

// ComputePoW finds a nonce such that SHA-256(challenge + nonce) has at least
// `difficulty` leading zero bits. Returns the nonce and resulting hash.
func ComputePoW(challenge string, difficulty int) (nonce uint64, hash string) {
	if difficulty <= 0 || difficulty > 64 {
		return 0, ""
	}
	for n := uint64(0); n < math.MaxUint64; n++ {
		h := powHash(challenge, n)
		if leadingZeroBits(h) >= difficulty {
			return n, hex.EncodeToString(h)
		}
	}
	return 0, ""
}

// VerifyPoW checks that SHA-256(challenge + nonce) has at least `difficulty`
// leading zero bits and matches the provided hash.
func VerifyPoW(challenge string, nonce uint64, difficulty int, hash string) bool {
	if difficulty <= 0 || difficulty > 64 {
		return false
	}
	h := powHash(challenge, nonce)
	if leadingZeroBits(h) < difficulty {
		return false
	}
	return hex.EncodeToString(h) == strings.ToLower(strings.TrimSpace(hash))
}

// powHash computes SHA-256(challenge || nonce_bytes).
func powHash(challenge string, nonce uint64) []byte {
	buf := make([]byte, len(challenge)+8)
	copy(buf, challenge)
	binary.BigEndian.PutUint64(buf[len(challenge):], nonce)
	sum := sha256.Sum256(buf)
	return sum[:]
}

// leadingZeroBits counts the number of leading zero bits in a byte slice.
func leadingZeroBits(data []byte) int {
	count := 0
	for _, b := range data {
		if b == 0 {
			count += 8
			continue
		}
		// Count leading zeros in this byte
		for i := 7; i >= 0; i-- {
			if b&(1<<uint(i)) == 0 {
				count++
			} else {
				return count
			}
		}
	}
	return count
}

// ApplyPoW computes PoW for a message and sets the pow.* extensions.
// The challenge is derived from author + title + body (first 128 bytes).
func ApplyPoW(extensions map[string]any, author, title, body string, difficulty int) error {
	challenge := powChallenge(author, title, body)
	nonce, hash := ComputePoW(challenge, difficulty)
	if hash == "" {
		return fmt.Errorf("failed to compute PoW with difficulty %d", difficulty)
	}
	extensions["pow.nonce"] = nonce
	extensions["pow.difficulty"] = difficulty
	extensions["pow.hash"] = hash
	return nil
}

// VerifyMessagePoW checks the PoW extensions on a message.
// Returns nil if no PoW is present (optional), error if PoW is invalid.
func VerifyMessagePoW(extensions map[string]any, author, title, body string) error {
	diffVal, ok := extensions["pow.difficulty"]
	if !ok {
		return nil // no PoW required
	}
	difficulty := toInt(diffVal)
	if difficulty <= 0 {
		return fmt.Errorf("invalid pow.difficulty")
	}
	nonceVal, ok := extensions["pow.nonce"]
	if !ok {
		return fmt.Errorf("pow.nonce missing")
	}
	nonce := toUint64(nonceVal)
	hashStr, _ := extensions["pow.hash"].(string)
	if hashStr == "" {
		return fmt.Errorf("pow.hash missing")
	}
	challenge := powChallenge(author, title, body)
	if !VerifyPoW(challenge, nonce, difficulty, hashStr) {
		return fmt.Errorf("invalid proof-of-work")
	}
	return nil
}

func powChallenge(author, title, body string) string {
	b := body
	if len(b) > 128 {
		b = b[:128]
	}
	return author + "|" + title + "|" + b
}

func toInt(v any) int {
	switch n := v.(type) {
	case int:
		return n
	case int64:
		return int(n)
	case float64:
		return int(n)
	case json.Number:
		i, _ := n.Int64()
		return int(i)
	}
	return 0
}

func toUint64(v any) uint64 {
	switch n := v.(type) {
	case uint64:
		return n
	case int:
		return uint64(n)
	case int64:
		return uint64(n)
	case float64:
		return uint64(n)
	case json.Number:
		i, _ := n.Int64()
		return uint64(i)
	}
	return 0
}
