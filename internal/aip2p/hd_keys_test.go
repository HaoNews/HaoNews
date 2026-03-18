package aip2p

import (
	"encoding/hex"
	"testing"
)

// SLIP-0010 test vectors for Ed25519
// Reference: https://github.com/satoshilabs/slips/blob/master/slip-0010.md

func TestGenerateMnemonic(t *testing.T) {
	mnemonic, err := GenerateMnemonic()
	if err != nil {
		t.Fatalf("GenerateMnemonic failed: %v", err)
	}
	if mnemonic == "" {
		t.Fatal("Generated mnemonic is empty")
	}
	// 24 words = 256 bit entropy
	words := len(splitMnemonic(mnemonic))
	if words != 24 {
		t.Errorf("Expected 24 words, got %d", words)
	}
}

func TestMnemonicToSeed(t *testing.T) {
	// SLIP-0010 test vector 1
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	seed, err := MnemonicToSeed(mnemonic, "")
	if err != nil {
		t.Fatalf("MnemonicToSeed failed: %v", err)
	}
	if len(seed) != 64 {
		t.Errorf("Expected 64 bytes seed, got %d", len(seed))
	}
}

func TestMnemonicToSeedInvalid(t *testing.T) {
	_, err := MnemonicToSeed("invalid mnemonic words", "")
	if err == nil {
		t.Fatal("Expected error for invalid mnemonic")
	}
}

func TestNewMasterKey(t *testing.T) {
	// SLIP-0010 test vector 1
	seedHex := "000102030405060708090a0b0c0d0e0f"
	seed, _ := hex.DecodeString(seedHex)

	masterKey, err := NewMasterKey(seed)
	if err != nil {
		t.Fatalf("NewMasterKey failed: %v", err)
	}
	if !masterKey.IsPrivate {
		t.Error("Master key should be private")
	}
	if len(masterKey.Key) != 32 {
		t.Errorf("Expected 32 bytes key, got %d", len(masterKey.Key))
	}
	if len(masterKey.ChainCode) != 32 {
		t.Errorf("Expected 32 bytes chain code, got %d", len(masterKey.ChainCode))
	}

	// SLIP-0010 Ed25519 test vector 1: master key
	expectedKey := "2b4be7f19ee27bbf30c667b642d5f4aa69fd169872f8fc3059c08ebae2eb19e7"
	expectedChainCode := "90046a93de5380a72b5e45010748567d5ea02bbf6522f979e05c0d8d8ca9fffb"

	if hex.EncodeToString(masterKey.Key) != expectedKey {
		t.Errorf("Master key mismatch\nExpected: %s\nGot:      %s", expectedKey, hex.EncodeToString(masterKey.Key))
	}
	if hex.EncodeToString(masterKey.ChainCode) != expectedChainCode {
		t.Errorf("Chain code mismatch\nExpected: %s\nGot:      %s", expectedChainCode, hex.EncodeToString(masterKey.ChainCode))
	}
}

func TestDeriveChild(t *testing.T) {
	// SLIP-0010 test vector 1
	seedHex := "000102030405060708090a0b0c0d0e0f"
	seed, _ := hex.DecodeString(seedHex)
	masterKey, _ := NewMasterKey(seed)

	// Derive m/0'
	child0, err := masterKey.DeriveChild(0)
	if err != nil {
		t.Fatalf("DeriveChild(0) failed: %v", err)
	}

	// SLIP-0010 Ed25519 test vector 1: m/0'
	expectedKey := "68e0fe46dfb67e368c75379acec591dad19df3cde26e63b93a8e704f1dade7a3"
	expectedChainCode := "8b59aa11380b624e81507a27fedda59fea6d0b779a778918a2fd3590e16e9c69"

	if hex.EncodeToString(child0.Key) != expectedKey {
		t.Errorf("Child key mismatch\nExpected: %s\nGot:      %s", expectedKey, hex.EncodeToString(child0.Key))
	}
	if hex.EncodeToString(child0.ChainCode) != expectedChainCode {
		t.Errorf("Chain code mismatch\nExpected: %s\nGot:      %s", expectedChainCode, hex.EncodeToString(child0.ChainCode))
	}
}

func TestDeriveChildHardened(t *testing.T) {
	seedHex := "000102030405060708090a0b0c0d0e0f"
	seed, _ := hex.DecodeString(seedHex)
	masterKey, _ := NewMasterKey(seed)

	// Non-hardened index should be converted to hardened for Ed25519
	child1, err := masterKey.DeriveChild(0)
	if err != nil {
		t.Fatalf("DeriveChild(0) failed: %v", err)
	}

	// Explicitly hardened should produce same result
	child2, err := masterKey.DeriveChild(0x80000000)
	if err != nil {
		t.Fatalf("DeriveChild(0x80000000) failed: %v", err)
	}

	if hex.EncodeToString(child1.Key) != hex.EncodeToString(child2.Key) {
		t.Error("Non-hardened and hardened derivation should produce same result for Ed25519")
	}
}

func TestParseDerivationPath(t *testing.T) {
	tests := []struct {
		path     string
		expected []uint32
		wantErr  bool
	}{
		{"m", []uint32{}, false},
		{"m/0", []uint32{0}, false},
		{"m/0/1", []uint32{0, 1}, false},
		{"m/0'/1'", []uint32{0x80000000, 0x80000001}, false},
		{"m/44'/0'/0'", []uint32{0x8000002c, 0x80000000, 0x80000000}, false},
		{"", []uint32{}, false},
		{"invalid", nil, true},
		{"m/", nil, true},
		{"m/abc", nil, true},
	}

	for _, tt := range tests {
		indices, err := ParseDerivationPath(tt.path)
		if tt.wantErr {
			if err == nil {
				t.Errorf("ParseDerivationPath(%q) expected error, got nil", tt.path)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseDerivationPath(%q) unexpected error: %v", tt.path, err)
			continue
		}
		if len(indices) != len(tt.expected) {
			t.Errorf("ParseDerivationPath(%q) length mismatch: expected %d, got %d", tt.path, len(tt.expected), len(indices))
			continue
		}
		for i := range indices {
			if indices[i] != tt.expected[i] {
				t.Errorf("ParseDerivationPath(%q)[%d] = %d, expected %d", tt.path, i, indices[i], tt.expected[i])
			}
		}
	}
}

func TestDerivePath(t *testing.T) {
	seedHex := "000102030405060708090a0b0c0d0e0f"
	seed, _ := hex.DecodeString(seedHex)
	masterKey, _ := NewMasterKey(seed)

	// Derive m/0/1
	key, err := DerivePath(masterKey, "m/0/1")
	if err != nil {
		t.Fatalf("DerivePath failed: %v", err)
	}
	if !key.IsPrivate {
		t.Error("Derived key should be private")
	}

	// SLIP-0010 Ed25519 test vector 1: m/0'/1'
	expectedKey := "b1d0bad404bf35da785a64ca1ac54b2617211d2777696fbffaf208f746ae84f2"
	expectedChainCode := "a320425f77d1b5c2505a6b1b27382b37368ee640e3557c315416801243552f14"

	if hex.EncodeToString(key.Key) != expectedKey {
		t.Errorf("Derived key mismatch\nExpected: %s\nGot:      %s", expectedKey, hex.EncodeToString(key.Key))
	}
	if hex.EncodeToString(key.ChainCode) != expectedChainCode {
		t.Errorf("Chain code mismatch\nExpected: %s\nGot:      %s", expectedChainCode, hex.EncodeToString(key.ChainCode))
	}
}

func TestDerivePathFromSeed(t *testing.T) {
	seedHex := "000102030405060708090a0b0c0d0e0f"
	seed, _ := hex.DecodeString(seedHex)

	key, err := DerivePathFromSeed(seed, "m/0/1/2")
	if err != nil {
		t.Fatalf("DerivePathFromSeed failed: %v", err)
	}
	if !key.IsPrivate {
		t.Error("Derived key should be private")
	}

	// SLIP-0010 Ed25519 test vector 1: m/0'/1'/2'
	expectedKey := "92a5b23c0b8a99e37d07df3fb9966917f5d06e02ddbd909c7e184371463e9fc9"
	expectedChainCode := "2e69929e00b5ab250f49c3fb1c12f252de4fed2c1db88387094a0f8c4c9ccd6c"

	if hex.EncodeToString(key.Key) != expectedKey {
		t.Errorf("Derived key mismatch\nExpected: %s\nGot:      %s", expectedKey, hex.EncodeToString(key.Key))
	}
	if hex.EncodeToString(key.ChainCode) != expectedChainCode {
		t.Errorf("Chain code mismatch\nExpected: %s\nGot:      %s", expectedChainCode, hex.EncodeToString(key.ChainCode))
	}
}

func TestPublicKey(t *testing.T) {
	seedHex := "000102030405060708090a0b0c0d0e0f"
	seed, _ := hex.DecodeString(seedHex)
	masterKey, _ := NewMasterKey(seed)

	pubKey := masterKey.PublicKey()
	if len(pubKey) != 32 {
		t.Errorf("Expected 32 bytes public key, got %d", len(pubKey))
	}

	// SLIP-0010 Ed25519 test vector 1: master public key
	expectedPubKey := "a4b2856bfec510abab89753fac1ac0e1112364e7d250545963f135f2a33188ed"
	if hex.EncodeToString(pubKey) != expectedPubKey {
		t.Errorf("Public key mismatch\nExpected: %s\nGot:      %s", expectedPubKey, hex.EncodeToString(pubKey))
	}
}

func TestPrivateKey(t *testing.T) {
	seedHex := "000102030405060708090a0b0c0d0e0f"
	seed, _ := hex.DecodeString(seedHex)
	masterKey, _ := NewMasterKey(seed)

	privKey := masterKey.PrivateKey()
	if len(privKey) != 64 {
		t.Errorf("Expected 64 bytes private key, got %d", len(privKey))
	}
}

func TestPathFromURI(t *testing.T) {
	tests := []struct {
		uri      string
		expected string
	}{
		{"agent://alice", "m/0"},
		{"agent://alice/work", "m/0/0"},
		{"agent://alice/personal", "m/0/0"}, // Same as work - sequential mapping
		{"agent://alice/bots/bot-1", "m/0/0/1"},
		{"invalid-uri", "m/0"},
		{"", "m/0"},
	}

	for _, tt := range tests {
		path := PathFromURI(tt.uri)
		if path != tt.expected {
			t.Errorf("PathFromURI(%q) = %q, expected %q", tt.uri, path, tt.expected)
		}
	}
}

func TestFormatPublicKey(t *testing.T) {
	pubKey := []byte{0x01, 0x02, 0x03, 0x04}
	formatted := FormatPublicKey(pubKey)
	expected := "ed25519:01020304"
	if formatted != expected {
		t.Errorf("FormatPublicKey = %q, expected %q", formatted, expected)
	}
}

func TestParsePublicKey(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		wantErr  bool
	}{
		{"ed25519:01020304", "01020304", false},
		{"01020304", "01020304", false},
		{"  ed25519:01020304  ", "01020304", false},
		{"invalid", "", true},
	}

	for _, tt := range tests {
		pubKey, err := ParsePublicKey(tt.input)
		if tt.wantErr {
			if err == nil {
				t.Errorf("ParsePublicKey(%q) expected error, got nil", tt.input)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParsePublicKey(%q) unexpected error: %v", tt.input, err)
			continue
		}
		if hex.EncodeToString(pubKey) != tt.expected {
			t.Errorf("ParsePublicKey(%q) = %s, expected %s", tt.input, hex.EncodeToString(pubKey), tt.expected)
		}
	}
}

func splitMnemonic(mnemonic string) []string {
	words := []string{}
	for _, word := range []rune(mnemonic) {
		if word == ' ' {
			continue
		}
		if len(words) == 0 || words[len(words)-1] != "" {
			words = append(words, "")
		}
		words[len(words)-1] += string(word)
	}
	// Proper word splitting
	result := []string{}
	current := ""
	for _, r := range mnemonic {
		if r == ' ' {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(r)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}
