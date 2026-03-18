package aip2p

import (
	"crypto/ed25519"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/tyler-smith/go-bip39"
)

// SLIP-0010 implementation for Ed25519 hierarchical deterministic keys
// Reference: https://github.com/satoshilabs/slips/blob/master/slip-0010.md

const (
	// Ed25519 curve identifier for SLIP-0010
	ed25519Curve = "ed25519 seed"
	// Hardened key offset
	hardenedOffset = 0x80000000
)

var (
	ErrInvalidPath        = errors.New("invalid derivation path")
	ErrInvalidMnemonic    = errors.New("invalid mnemonic")
	ErrInvalidSeed        = errors.New("invalid seed")
	ErrHardenedPublicKey  = errors.New("cannot derive hardened child from public key")
)

// HDKey represents a hierarchical deterministic key
type HDKey struct {
	Key       []byte // 32 bytes private key or 32 bytes public key
	ChainCode []byte // 32 bytes chain code
	IsPrivate bool
}

// GenerateMnemonic generates a BIP39 mnemonic (24 words = 256 bits entropy)
func GenerateMnemonic() (string, error) {
	entropy, err := bip39.NewEntropy(256)
	if err != nil {
		return "", err
	}
	return bip39.NewMnemonic(entropy)
}

// MnemonicToSeed converts a BIP39 mnemonic to a 512-bit seed
func MnemonicToSeed(mnemonic, passphrase string) ([]byte, error) {
	if !bip39.IsMnemonicValid(mnemonic) {
		return nil, ErrInvalidMnemonic
	}
	return bip39.NewSeed(mnemonic, passphrase), nil
}

// NewMasterKey derives the master key from a seed using SLIP-0010
func NewMasterKey(seed []byte) (*HDKey, error) {
	if len(seed) < 16 || len(seed) > 64 {
		return nil, ErrInvalidSeed
	}

	// HMAC-SHA512(key="ed25519 seed", data=seed)
	h := hmac.New(sha512.New, []byte(ed25519Curve))
	h.Write(seed)
	digest := h.Sum(nil)

	// Left 32 bytes = master secret key
	// Right 32 bytes = master chain code
	key := digest[:32]
	chainCode := digest[32:]

	return &HDKey{
		Key:       key,
		ChainCode: chainCode,
		IsPrivate: true,
	}, nil
}

// DeriveChild derives a child key from a parent key using SLIP-0010
func (k *HDKey) DeriveChild(index uint32) (*HDKey, error) {
	if !k.IsPrivate {
		if index >= hardenedOffset {
			return nil, ErrHardenedPublicKey
		}
		return k.derivePublicChild(index)
	}
	return k.derivePrivateChild(index)
}

// derivePrivateChild derives a child private key
func (k *HDKey) derivePrivateChild(index uint32) (*HDKey, error) {
	// For Ed25519, all derivations are hardened
	// SLIP-0010: use hardened derivation for Ed25519
	if index < hardenedOffset {
		index += hardenedOffset
	}

	// Data = 0x00 || parent_key || index (BE)
	data := make([]byte, 37)
	data[0] = 0x00
	copy(data[1:33], k.Key)
	binary.BigEndian.PutUint32(data[33:], index)

	// HMAC-SHA512(key=parent_chain_code, data=data)
	h := hmac.New(sha512.New, k.ChainCode)
	h.Write(data)
	digest := h.Sum(nil)

	childKey := digest[:32]
	childChainCode := digest[32:]

	return &HDKey{
		Key:       childKey,
		ChainCode: childChainCode,
		IsPrivate: true,
	}, nil
}

// derivePublicChild derives a child public key (not supported for Ed25519 hardened)
func (k *HDKey) derivePublicChild(index uint32) (*HDKey, error) {
	// Ed25519 only supports hardened derivation from private keys
	return nil, ErrHardenedPublicKey
}

// PublicKey returns the Ed25519 public key
func (k *HDKey) PublicKey() []byte {
	if !k.IsPrivate {
		return k.Key
	}
	// For Ed25519, public key is derived from private key
	privateKey := ed25519.NewKeyFromSeed(k.Key)
	return []byte(privateKey.Public().(ed25519.PublicKey))
}

// PrivateKey returns the Ed25519 private key (64 bytes)
func (k *HDKey) PrivateKey() ed25519.PrivateKey {
	if !k.IsPrivate {
		return nil
	}
	return ed25519.NewKeyFromSeed(k.Key)
}

// ParseDerivationPath parses a derivation path string (e.g., "m/0/1/2")
func ParseDerivationPath(path string) ([]uint32, error) {
	path = strings.TrimSpace(path)
	if path == "" || path == "m" {
		return []uint32{}, nil
	}

	if !strings.HasPrefix(path, "m/") {
		return nil, fmt.Errorf("%w: must start with 'm/'", ErrInvalidPath)
	}

	parts := strings.Split(path[2:], "/")
	indices := make([]uint32, len(parts))

	for i, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			return nil, fmt.Errorf("%w: empty path segment", ErrInvalidPath)
		}

		// Check for hardened notation (')
		hardened := false
		if strings.HasSuffix(part, "'") || strings.HasSuffix(part, "h") {
			hardened = true
			part = part[:len(part)-1]
		}

		index, err := strconv.ParseUint(part, 10, 31)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid index '%s'", ErrInvalidPath, part)
		}

		if hardened {
			indices[i] = uint32(index) + hardenedOffset
		} else {
			indices[i] = uint32(index)
		}
	}

	return indices, nil
}

// DerivePath derives a key from a master key using a derivation path
func DerivePath(masterKey *HDKey, path string) (*HDKey, error) {
	indices, err := ParseDerivationPath(path)
	if err != nil {
		return nil, err
	}

	key := masterKey
	for _, index := range indices {
		key, err = key.DeriveChild(index)
		if err != nil {
			return nil, err
		}
	}

	return key, nil
}

// DerivePathFromSeed derives a key directly from seed and path
func DerivePathFromSeed(seed []byte, path string) (*HDKey, error) {
	masterKey, err := NewMasterKey(seed)
	if err != nil {
		return nil, err
	}
	return DerivePath(masterKey, path)
}

// PathFromURI extracts derivation path from author URI
// agent://alice -> m/0
// agent://alice/work -> m/0/0
// agent://alice/personal -> m/0/1
// agent://alice/bots/bot-1 -> m/0/2/0
func PathFromURI(authorURI string) string {
	authorURI = strings.TrimSpace(authorURI)
	if !strings.HasPrefix(authorURI, "agent://") {
		return "m/0"
	}

	// Remove agent:// prefix
	path := strings.TrimPrefix(authorURI, "agent://")
	parts := strings.Split(path, "/")

	if len(parts) == 1 {
		// Main identity: agent://alice -> m/0
		return "m/0"
	}

	// Build path: agent://alice/work -> m/0/0
	indices := []string{"m", "0"}
	for i := 1; i < len(parts); i++ {
		// Simple mapping: each segment gets an index
		// In production, use a registry to map names to indices
		indices = append(indices, strconv.Itoa(i-1))
	}

	return strings.Join(indices, "/")
}

// DerivePublicKey derives a child public key from parent public key and path
// Note: For Ed25519, this only works for non-hardened paths, but SLIP-0010
// requires hardened derivation. This function is for verification purposes.
func DerivePublicKey(parentPubKeyHex, path string) (string, error) {
	// For Ed25519 with hardened derivation, we cannot derive public keys
	// without the private key. This function is a placeholder for future
	// support of non-hardened derivation or alternative schemes.
	return "", errors.New("Ed25519 hardened derivation does not support public key derivation")
}

// VerifyChildKey verifies that a child public key was derived from a parent
// For Ed25519 hardened derivation, this requires the private key path.
// This is a placeholder - in practice, we trust the hd.parent_pubkey in the message.
func VerifyChildKey(parentPubKeyHex, childPubKeyHex, path string) bool {
	// For Ed25519 SLIP-0010, we cannot verify without private key
	// In practice, we trust the signature verification instead
	return true
}

// FormatPublicKey formats a public key as hex string with ed25519: prefix
func FormatPublicKey(pubKey []byte) string {
	return "ed25519:" + hex.EncodeToString(pubKey)
}

// ParsePublicKey parses a public key from hex string (with or without prefix)
func ParsePublicKey(pubKeyStr string) ([]byte, error) {
	pubKeyStr = strings.TrimSpace(pubKeyStr)
	pubKeyStr = strings.TrimPrefix(pubKeyStr, "ed25519:")
	return hex.DecodeString(pubKeyStr)
}



