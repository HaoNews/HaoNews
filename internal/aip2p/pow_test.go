package aip2p

import (
	"testing"
)

func TestComputeAndVerifyPoW(t *testing.T) {
	challenge := "agent://test|Hello|body content"
	difficulty := 8 // 8 leading zero bits = first byte is 0x00

	nonce, hash := ComputePoW(challenge, difficulty)
	if hash == "" {
		t.Fatal("ComputePoW returned empty hash")
	}
	if !VerifyPoW(challenge, nonce, difficulty, hash) {
		t.Fatalf("VerifyPoW failed for nonce=%d hash=%s", nonce, hash)
	}
}

func TestVerifyPoWRejectsWrongNonce(t *testing.T) {
	challenge := "agent://test|Hello|body"
	difficulty := 8

	nonce, hash := ComputePoW(challenge, difficulty)
	if VerifyPoW(challenge, nonce+1, difficulty, hash) {
		t.Fatal("VerifyPoW should reject wrong nonce")
	}
}

func TestVerifyPoWRejectsWrongHash(t *testing.T) {
	challenge := "agent://test|Hello|body"
	difficulty := 8

	nonce, _ := ComputePoW(challenge, difficulty)
	if VerifyPoW(challenge, nonce, difficulty, "deadbeef") {
		t.Fatal("VerifyPoW should reject wrong hash")
	}
}

func TestVerifyPoWRejectsInvalidDifficulty(t *testing.T) {
	if VerifyPoW("x", 0, 0, "abc") {
		t.Fatal("should reject difficulty 0")
	}
	if VerifyPoW("x", 0, 65, "abc") {
		t.Fatal("should reject difficulty > 64")
	}
}

func TestComputePoWInvalidDifficulty(t *testing.T) {
	_, hash := ComputePoW("x", 0)
	if hash != "" {
		t.Fatal("should return empty for difficulty 0")
	}
	_, hash = ComputePoW("x", 65)
	if hash != "" {
		t.Fatal("should return empty for difficulty > 64")
	}
}

func TestApplyAndVerifyMessagePoW(t *testing.T) {
	ext := make(map[string]any)
	err := ApplyPoW(ext, "agent://a", "title", "body text", 8)
	if err != nil {
		t.Fatalf("ApplyPoW failed: %v", err)
	}
	if _, ok := ext["pow.nonce"]; !ok {
		t.Fatal("missing pow.nonce")
	}
	if _, ok := ext["pow.difficulty"]; !ok {
		t.Fatal("missing pow.difficulty")
	}
	if _, ok := ext["pow.hash"]; !ok {
		t.Fatal("missing pow.hash")
	}

	err = VerifyMessagePoW(ext, "agent://a", "title", "body text")
	if err != nil {
		t.Fatalf("VerifyMessagePoW failed: %v", err)
	}
}

func TestVerifyMessagePoWSkipsWhenNoPow(t *testing.T) {
	ext := make(map[string]any)
	err := VerifyMessagePoW(ext, "agent://a", "title", "body")
	if err != nil {
		t.Fatalf("should skip when no PoW: %v", err)
	}
}

func TestVerifyMessagePoWRejectsTampered(t *testing.T) {
	ext := make(map[string]any)
	ApplyPoW(ext, "agent://a", "title", "body", 8)

	err := VerifyMessagePoW(ext, "agent://a", "title", "TAMPERED body")
	if err == nil {
		t.Fatal("should reject tampered body")
	}
}

func TestLeadingZeroBits(t *testing.T) {
	tests := []struct {
		data []byte
		want int
	}{
		{[]byte{0x00, 0x00}, 16},
		{[]byte{0x00, 0x80}, 8},
		{[]byte{0x00, 0x01}, 15},
		{[]byte{0x80}, 0},
		{[]byte{0x40}, 1},
		{[]byte{0x01}, 7},
		{[]byte{}, 0},
	}
	for _, tt := range tests {
		got := leadingZeroBits(tt.data)
		if got != tt.want {
			t.Errorf("leadingZeroBits(%x) = %d, want %d", tt.data, got, tt.want)
		}
	}
}
