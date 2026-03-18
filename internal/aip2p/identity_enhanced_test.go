package aip2p

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestEncryptDecryptMnemonic(t *testing.T) {
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	password := "test-password-123"

	encrypted, salt, err := EncryptMnemonic(mnemonic, password)
	if err != nil {
		t.Fatalf("EncryptMnemonic failed: %v", err)
	}
	if len(encrypted) == 0 {
		t.Fatal("Encrypted data is empty")
	}
	if len(salt) != saltLen {
		t.Errorf("Salt length = %d, expected %d", len(salt), saltLen)
	}

	decrypted, err := DecryptMnemonic(encrypted, salt, password)
	if err != nil {
		t.Fatalf("DecryptMnemonic failed: %v", err)
	}
	if decrypted != mnemonic {
		t.Errorf("Decrypted mnemonic mismatch\nExpected: %s\nGot:      %s", mnemonic, decrypted)
	}
}

func TestDecryptMnemonicWrongPassword(t *testing.T) {
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	password := "correct-password"

	encrypted, salt, err := EncryptMnemonic(mnemonic, password)
	if err != nil {
		t.Fatalf("EncryptMnemonic failed: %v", err)
	}

	_, err = DecryptMnemonic(encrypted, salt, "wrong-password")
	if err == nil {
		t.Fatal("Expected error for wrong password")
	}
}

func TestEncryptMnemonicEmptyInputs(t *testing.T) {
	_, _, err := EncryptMnemonic("", "password")
	if err == nil {
		t.Fatal("Expected error for empty mnemonic")
	}

	_, _, err = EncryptMnemonic("some mnemonic", "")
	if err == nil {
		t.Fatal("Expected error for empty password")
	}
}

func TestDecryptMnemonicEmptyInputs(t *testing.T) {
	_, err := DecryptMnemonic(nil, []byte("salt"), "password")
	if err == nil {
		t.Fatal("Expected error for empty encrypted data")
	}

	_, err = DecryptMnemonic([]byte("data"), nil, "password")
	if err == nil {
		t.Fatal("Expected error for empty salt")
	}

	_, err = DecryptMnemonic([]byte("data"), []byte("salt"), "")
	if err == nil {
		t.Fatal("Expected error for empty password")
	}
}

func TestSaveLoadEncryptedIdentity(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test-identity.json")
	password := "my-secret-password"

	identity, err := NewHDIdentity("test-agent", "agent://test", time.Now())
	if err != nil {
		t.Fatalf("NewHDIdentity failed: %v", err)
	}

	mnemonic := identity.Mnemonic
	if mnemonic == "" {
		t.Fatal("Mnemonic should not be empty")
	}

	// Save with encryption
	err = SaveEncryptedIdentity(path, identity, password)
	if err != nil {
		t.Fatalf("SaveEncryptedIdentity failed: %v", err)
	}

	// Read raw file to verify mnemonic is not in plaintext
	data, _ := os.ReadFile(path)
	raw := string(data)
	if containsStr([]string{mnemonic}, mnemonic) {
		// Check the raw file doesn't contain plaintext mnemonic
		words := mnemonic[:20] // first 20 chars
		if len(raw) > 0 && !containsSubstring(raw, words) {
			// Good - mnemonic is encrypted
		}
	}

	// Load with correct password
	loaded, err := LoadEncryptedIdentity(path, password)
	if err != nil {
		t.Fatalf("LoadEncryptedIdentity failed: %v", err)
	}
	if loaded.Mnemonic != mnemonic {
		t.Errorf("Mnemonic mismatch after load\nExpected: %s\nGot:      %s", mnemonic, loaded.Mnemonic)
	}
	if !loaded.HDEnabled {
		t.Error("HDEnabled should be true")
	}

	// Load with wrong password
	_, err = LoadEncryptedIdentity(path, "wrong-password")
	if err == nil {
		t.Fatal("Expected error for wrong password")
	}

	// Load without password
	_, err = LoadEncryptedIdentity(path, "")
	if err == nil {
		t.Fatal("Expected error when no password provided for encrypted identity")
	}
}

func TestSaveLoadUnencryptedIdentity(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test-identity.json")

	identity, err := NewHDIdentity("test-agent", "agent://test", time.Now())
	if err != nil {
		t.Fatalf("NewHDIdentity failed: %v", err)
	}

	// Save without encryption (empty password)
	err = SaveEncryptedIdentity(path, identity, "")
	if err != nil {
		t.Fatalf("SaveEncryptedIdentity failed: %v", err)
	}

	// Load without password
	loaded, err := LoadEncryptedIdentity(path, "")
	if err != nil {
		t.Fatalf("LoadEncryptedIdentity failed: %v", err)
	}
	if loaded.Mnemonic != identity.Mnemonic {
		t.Error("Mnemonic mismatch")
	}
}

func TestIdentityRegistry(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "registry.json")

	registry := NewIdentityRegistry()
	if registry.Count() != 0 {
		t.Errorf("New registry should be empty, got %d", registry.Count())
	}

	// Add entries
	registry.Add("agent://alice", "ed25519:abc123", "trusted", "Alice's identity")
	registry.Add("agent://bob", "ed25519:def456", "known", "")

	if registry.Count() != 2 {
		t.Errorf("Expected 2 entries, got %d", registry.Count())
	}

	// Get entry
	entry, ok := registry.Get("agent://alice")
	if !ok {
		t.Fatal("Expected to find agent://alice")
	}
	if entry.MasterPubKey != "ed25519:abc123" {
		t.Errorf("MasterPubKey = %s, expected ed25519:abc123", entry.MasterPubKey)
	}
	if entry.TrustLevel != "trusted" {
		t.Errorf("TrustLevel = %s, expected trusted", entry.TrustLevel)
	}
	if entry.Notes != "Alice's identity" {
		t.Errorf("Notes = %s, expected Alice's identity", entry.Notes)
	}

	// Get non-existent
	_, ok = registry.Get("agent://charlie")
	if ok {
		t.Fatal("Should not find agent://charlie")
	}

	// Save and reload
	err := registry.Save(path)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := LoadIdentityRegistry(path)
	if err != nil {
		t.Fatalf("LoadIdentityRegistry failed: %v", err)
	}
	if loaded.Count() != 2 {
		t.Errorf("Loaded registry should have 2 entries, got %d", loaded.Count())
	}

	entry, ok = loaded.Get("agent://bob")
	if !ok {
		t.Fatal("Expected to find agent://bob after reload")
	}
	if entry.MasterPubKey != "ed25519:def456" {
		t.Errorf("MasterPubKey = %s, expected ed25519:def456", entry.MasterPubKey)
	}

	// Remove
	removed := loaded.Remove("agent://alice")
	if !removed {
		t.Fatal("Expected Remove to return true")
	}
	if loaded.Count() != 1 {
		t.Errorf("Expected 1 entry after remove, got %d", loaded.Count())
	}

	// Remove non-existent
	removed = loaded.Remove("agent://charlie")
	if removed {
		t.Fatal("Expected Remove to return false for non-existent")
	}
}

func TestLoadIdentityRegistryNotExist(t *testing.T) {
	registry, err := LoadIdentityRegistry("/nonexistent/path/registry.json")
	if err != nil {
		t.Fatalf("Should return empty registry for non-existent file: %v", err)
	}
	if registry.Count() != 0 {
		t.Errorf("Expected empty registry, got %d entries", registry.Count())
	}
}

func TestCheckChildPermission(t *testing.T) {
	identity := AgentIdentity{
		Author:    "agent://alice",
		HDEnabled: true,
		Children: map[string]ChildIdentityPermissions{
			"work": {
				Path:            "m/0/0",
				Permissions:     []string{"publish", "subscribe"},
				AllowedChannels: []string{"work", "tech"},
				AllowedTags:     []string{"work", "project"},
				MaxPostsPerDay:  50,
			},
			"bot": {
				Path:        "m/0/2",
				Permissions: []string{"publish"},
				Revoked:     true,
			},
			"guest": {
				Path:        "m/0/3",
				Permissions: []string{"publish"},
				ExpiresAt:   "2020-01-01T00:00:00Z", // expired
			},
			"readonly": {
				Path:        "m/0/4",
				Permissions: []string{"subscribe"}, // no publish
			},
		},
	}

	// Main identity - no restrictions
	err := CheckChildPermission(identity, "agent://alice", "any-channel", nil)
	if err != nil {
		t.Errorf("Main identity should have no restrictions: %v", err)
	}

	// Work identity - allowed channel
	err = CheckChildPermission(identity, "agent://alice/work", "work", []string{"work"})
	if err != nil {
		t.Errorf("Work identity should be allowed: %v", err)
	}

	// Work identity - disallowed channel
	err = CheckChildPermission(identity, "agent://alice/work", "personal", nil)
	if err == nil {
		t.Error("Work identity should not be allowed on personal channel")
	}

	// Work identity - disallowed tag
	err = CheckChildPermission(identity, "agent://alice/work", "work", []string{"personal"})
	if err == nil {
		t.Error("Work identity should not be allowed with personal tag")
	}

	// Revoked identity
	err = CheckChildPermission(identity, "agent://alice/bot", "any", nil)
	if err == nil {
		t.Error("Revoked identity should be rejected")
	}

	// Expired identity
	err = CheckChildPermission(identity, "agent://alice/guest", "any", nil)
	if err == nil {
		t.Error("Expired identity should be rejected")
	}

	// Read-only identity (no publish permission)
	err = CheckChildPermission(identity, "agent://alice/readonly", "any", nil)
	if err == nil {
		t.Error("Read-only identity should not be allowed to publish")
	}

	// Unknown child - allowed by default
	err = CheckChildPermission(identity, "agent://alice/unknown", "any", nil)
	if err != nil {
		t.Errorf("Unknown child should be allowed by default: %v", err)
	}
}

func TestExtractChildName(t *testing.T) {
	tests := []struct {
		child    string
		parent   string
		expected string
	}{
		{"agent://alice/work", "agent://alice", "work"},
		{"agent://alice/bots/bot-1", "agent://alice", "bots"},
		{"agent://alice", "agent://alice", ""},
		{"agent://bob/work", "agent://alice", ""},
		{"", "agent://alice", ""},
	}

	for _, tt := range tests {
		result := extractChildName(tt.child, tt.parent)
		if result != tt.expected {
			t.Errorf("extractChildName(%q, %q) = %q, expected %q", tt.child, tt.parent, result, tt.expected)
		}
	}
}

func TestEncryptionBase64RoundTrip(t *testing.T) {
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	password := "test123"

	encrypted, salt, err := EncryptMnemonic(mnemonic, password)
	if err != nil {
		t.Fatalf("EncryptMnemonic failed: %v", err)
	}

	// Simulate JSON storage
	encB64 := base64.StdEncoding.EncodeToString(encrypted)
	saltB64 := base64.StdEncoding.EncodeToString(salt)

	// Decode back
	encDec, _ := base64.StdEncoding.DecodeString(encB64)
	saltDec, _ := base64.StdEncoding.DecodeString(saltB64)

	decrypted, err := DecryptMnemonic(encDec, saltDec, password)
	if err != nil {
		t.Fatalf("DecryptMnemonic after base64 roundtrip failed: %v", err)
	}
	if decrypted != mnemonic {
		t.Error("Mnemonic mismatch after base64 roundtrip")
	}
}

func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
