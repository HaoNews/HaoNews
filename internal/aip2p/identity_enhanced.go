package aip2p

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/crypto/argon2"
)

// Encryption constants
const (
	EncryptionMethodAES256GCM = "aes-256-gcm+argon2id"
	argon2Time                = 3
	argon2Memory              = 64 * 1024 // 64 MB
	argon2Threads             = 4
	argon2KeyLen              = 32
	saltLen                   = 32
)

// EncryptMnemonic encrypts a mnemonic with a password using AES-256-GCM + Argon2id
func EncryptMnemonic(mnemonic, password string) (encrypted, salt []byte, err error) {
	if mnemonic == "" {
		return nil, nil, errors.New("mnemonic is empty")
	}
	if password == "" {
		return nil, nil, errors.New("password is empty")
	}

	salt = make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return nil, nil, err
	}

	key := argon2.IDKey([]byte(password), salt, argon2Time, argon2Memory, argon2Threads, argon2KeyLen)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(mnemonic), nil)
	return ciphertext, salt, nil
}

// DecryptMnemonic decrypts a mnemonic with a password
func DecryptMnemonic(encrypted, salt []byte, password string) (string, error) {
	if len(encrypted) == 0 {
		return "", errors.New("encrypted data is empty")
	}
	if len(salt) == 0 {
		return "", errors.New("salt is empty")
	}
	if password == "" {
		return "", errors.New("password is empty")
	}

	key := argon2.IDKey([]byte(password), salt, argon2Time, argon2Memory, argon2Threads, argon2KeyLen)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(encrypted) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertext := encrypted[:nonceSize], encrypted[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", errors.New("decryption failed: wrong password")
	}

	return string(plaintext), nil
}

// SaveEncryptedIdentity saves an identity with encrypted mnemonic
func SaveEncryptedIdentity(path string, identity AgentIdentity, password string) error {
	if identity.HDEnabled && identity.Mnemonic != "" && password != "" {
		encrypted, salt, err := EncryptMnemonic(identity.Mnemonic, password)
		if err != nil {
			return err
		}
		identity.MnemonicEncrypted = base64.StdEncoding.EncodeToString(encrypted)
		identity.EncryptionSalt = base64.StdEncoding.EncodeToString(salt)
		identity.EncryptionMethod = EncryptionMethodAES256GCM
		identity.Mnemonic = "" // clear plaintext
	}
	return SaveAgentIdentity(path, identity)
}

// LoadEncryptedIdentity loads an identity and decrypts mnemonic if needed
func LoadEncryptedIdentity(path, password string) (AgentIdentity, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return AgentIdentity{}, err
	}
	var identity AgentIdentity
	if err := json.Unmarshal(data, &identity); err != nil {
		return AgentIdentity{}, err
	}

	// If mnemonic is encrypted, decrypt it
	if identity.MnemonicEncrypted != "" {
		if password == "" {
			return AgentIdentity{}, errors.New("password required to unlock encrypted identity")
		}
		encrypted, err := base64.StdEncoding.DecodeString(identity.MnemonicEncrypted)
		if err != nil {
			return AgentIdentity{}, fmt.Errorf("invalid encrypted mnemonic: %w", err)
		}
		salt, err := base64.StdEncoding.DecodeString(identity.EncryptionSalt)
		if err != nil {
			return AgentIdentity{}, fmt.Errorf("invalid encryption salt: %w", err)
		}
		mnemonic, err := DecryptMnemonic(encrypted, salt, password)
		if err != nil {
			return AgentIdentity{}, err
		}
		identity.Mnemonic = mnemonic
	}

	if err := identity.ValidatePrivate(); err != nil {
		return AgentIdentity{}, err
	}
	return identity, nil
}

// IdentityRegistry manages known identities
type IdentityRegistry struct {
	Entries map[string]IdentityRegistryEntry `json:"entries"`
}

// IdentityRegistryEntry represents a known identity
type IdentityRegistryEntry struct {
	MasterPubKey string `json:"master_pubkey"`
	TrustLevel   string `json:"trust_level"` // "trusted", "known", "unknown"
	AddedAt      string `json:"added_at"`
	Notes        string `json:"notes,omitempty"`
}

// NewIdentityRegistry creates an empty registry
func NewIdentityRegistry() *IdentityRegistry {
	return &IdentityRegistry{
		Entries: make(map[string]IdentityRegistryEntry),
	}
}

// LoadIdentityRegistry loads registry from file
func LoadIdentityRegistry(path string) (*IdentityRegistry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return NewIdentityRegistry(), nil
		}
		return nil, err
	}
	var registry IdentityRegistry
	if err := json.Unmarshal(data, &registry); err != nil {
		return nil, err
	}
	if registry.Entries == nil {
		registry.Entries = make(map[string]IdentityRegistryEntry)
	}
	return &registry, nil
}

// Save writes registry to file
func (r *IdentityRegistry) Save(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o600)
}

// Add adds or updates an identity in the registry
func (r *IdentityRegistry) Add(author, masterPubKey, trustLevel, notes string) {
	author = strings.TrimSpace(author)
	if trustLevel == "" {
		trustLevel = "known"
	}
	r.Entries[author] = IdentityRegistryEntry{
		MasterPubKey: strings.TrimSpace(masterPubKey),
		TrustLevel:   trustLevel,
		AddedAt:      time.Now().UTC().Format(time.RFC3339),
		Notes:        notes,
	}
}

// Get retrieves an identity from the registry
func (r *IdentityRegistry) Get(author string) (IdentityRegistryEntry, bool) {
	entry, ok := r.Entries[strings.TrimSpace(author)]
	return entry, ok
}

// Remove removes an identity from the registry
func (r *IdentityRegistry) Remove(author string) bool {
	author = strings.TrimSpace(author)
	if _, ok := r.Entries[author]; ok {
		delete(r.Entries, author)
		return true
	}
	return false
}

// List returns all entries
func (r *IdentityRegistry) List() map[string]IdentityRegistryEntry {
	return r.Entries
}

// Count returns the number of entries
func (r *IdentityRegistry) Count() int {
	return len(r.Entries)
}

// ChildIdentityPermissions defines permissions for a child identity
type ChildIdentityPermissions struct {
	Path            string   `json:"path"`
	Permissions     []string `json:"permissions"`                // ["publish", "subscribe"]
	AllowedChannels []string `json:"allowed_channels,omitempty"` // empty = all channels
	AllowedTags     []string `json:"allowed_tags,omitempty"`     // empty = all tags
	MaxPostsPerDay  int      `json:"max_posts_per_day,omitempty"`
	ExpiresAt       string   `json:"expires_at,omitempty"`
	Revoked         bool     `json:"revoked,omitempty"`
	Notes           string   `json:"notes,omitempty"`
}

// CheckChildPermission checks if a child identity has permission to publish a message
func CheckChildPermission(identity AgentIdentity, author, channel string, tags []string) error {
	if !identity.HDEnabled {
		return nil
	}
	if identity.Children == nil {
		return nil
	}

	childName := extractChildName(author, identity.Author)
	if childName == "" {
		return nil // main identity, no restrictions
	}

	perms, ok := identity.Children[childName]
	if !ok {
		return nil // no permissions configured, allow by default
	}

	if perms.Revoked {
		return fmt.Errorf("child identity %q has been revoked", childName)
	}

	if perms.ExpiresAt != "" {
		expiresAt, err := time.Parse(time.RFC3339, perms.ExpiresAt)
		if err == nil && time.Now().After(expiresAt) {
			return fmt.Errorf("child identity %q has expired", childName)
		}
	}

	if !containsStr(perms.Permissions, "publish") {
		return fmt.Errorf("child identity %q does not have publish permission", childName)
	}

	if len(perms.AllowedChannels) > 0 && channel != "" {
		if !containsStr(perms.AllowedChannels, channel) {
			return fmt.Errorf("channel %q not allowed for child identity %q", channel, childName)
		}
	}

	if len(perms.AllowedTags) > 0 && len(tags) > 0 {
		for _, tag := range tags {
			if !containsStr(perms.AllowedTags, tag) {
				return fmt.Errorf("tag %q not allowed for child identity %q", tag, childName)
			}
		}
	}

	return nil
}

// extractChildName extracts child name from author URI
// agent://alice/work, agent://alice -> "work"
// agent://alice, agent://alice -> ""
func extractChildName(childAuthor, parentAuthor string) string {
	childAuthor = strings.TrimSpace(childAuthor)
	parentAuthor = strings.TrimSpace(parentAuthor)
	prefix := parentAuthor + "/"
	if !strings.HasPrefix(childAuthor, prefix) {
		return ""
	}
	rest := strings.TrimPrefix(childAuthor, prefix)
	parts := strings.SplitN(rest, "/", 2)
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

func containsStr(list []string, value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	for _, item := range list {
		if strings.ToLower(strings.TrimSpace(item)) == value {
			return true
		}
	}
	return false
}
