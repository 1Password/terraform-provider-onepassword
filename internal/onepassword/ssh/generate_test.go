package ssh

import (
	"strings"
	"testing"
)

func TestGenerateKeyPairEd25519(t *testing.T) {
	t.Parallel()

	key, err := GenerateKeyPair("ed25519", 0)
	if err != nil {
		t.Fatalf("GenerateKeyPair() error = %v", err)
	}

	if !strings.HasPrefix(key.PublicKey, "ssh-ed25519 ") {
		t.Errorf("PublicKey: got %q, want ssh-ed25519 prefix", key.PublicKey)
	}
	if !strings.HasPrefix(key.Fingerprint, "SHA256:") {
		t.Errorf("Fingerprint: got %q, want SHA256: prefix", key.Fingerprint)
	}
	if key.KeyType != "Ed25519" {
		t.Errorf("KeyType: got %q, want Ed25519", key.KeyType)
	}

	// The stored PKCS#8 key must convert to OpenSSH like the data source does.
	openSSH, err := PrivateKeyToOpenSSH([]byte(key.PrivateKeyPKCS8), "test-uuid")
	if err != nil {
		t.Fatalf("PrivateKeyToOpenSSH() error = %v", err)
	}
	if !strings.Contains(openSSH, "OPENSSH PRIVATE KEY") {
		t.Errorf("OpenSSH conversion: got %q, want OPENSSH PRIVATE KEY block", openSSH)
	}

	// Round-trip back to PKCS#8 must reproduce the stored key exactly.
	pkcs8, err := OpenSSHToPKCS8(openSSH)
	if err != nil {
		t.Fatalf("OpenSSHToPKCS8() error = %v", err)
	}
	if pkcs8 != key.PrivateKeyPKCS8 {
		t.Errorf("OpenSSHToPKCS8() round-trip mismatch")
	}
}

func TestGenerateKeyPairRSA(t *testing.T) {
	t.Parallel()

	key, err := GenerateKeyPair("rsa", 2048)
	if err != nil {
		t.Fatalf("GenerateKeyPair() error = %v", err)
	}

	if !strings.HasPrefix(key.PublicKey, "ssh-rsa ") {
		t.Errorf("PublicKey: got %q, want ssh-rsa prefix", key.PublicKey)
	}
	if key.KeyType != "RSA, 2048-bit" {
		t.Errorf("KeyType: got %q, want \"RSA, 2048-bit\"", key.KeyType)
	}

	fingerprint, err := PublicKeyFingerprint(key.PublicKey)
	if err != nil {
		t.Fatalf("PublicKeyFingerprint() error = %v", err)
	}
	if fingerprint != key.Fingerprint {
		t.Errorf("PublicKeyFingerprint() = %q, want %q", fingerprint, key.Fingerprint)
	}

	keyType, err := KeyTypeFromPublicKey(key.PublicKey)
	if err != nil {
		t.Fatalf("KeyTypeFromPublicKey() error = %v", err)
	}
	if keyType != key.KeyType {
		t.Errorf("KeyTypeFromPublicKey() = %q, want %q", keyType, key.KeyType)
	}
}

func TestGenerateKeyPairValidation(t *testing.T) {
	t.Parallel()

	if _, err := GenerateKeyPair("dsa", 0); err == nil {
		t.Error("dsa: want error, got nil")
	}
	if _, err := GenerateKeyPair("rsa", 1024); err == nil {
		t.Error("rsa 1024 bits: want error, got nil")
	}
}

func TestPublicKeyFromPrivateKey(t *testing.T) {
	t.Parallel()

	ed25519Key, err := GenerateKeyPair("ed25519", 0)
	if err != nil {
		t.Fatalf("GenerateKeyPair() error = %v", err)
	}
	rsaKey, err := GenerateKeyPair("rsa", 2048)
	if err != nil {
		t.Fatalf("GenerateKeyPair() error = %v", err)
	}

	tests := map[string]struct {
		privateKeyPEM string
		expected      string
	}{
		"derives from PKCS#8 ed25519": {
			privateKeyPEM: ed25519Key.PrivateKeyPKCS8,
			expected:      ed25519Key.PublicKey,
		},
		"derives from PKCS#8 rsa": {
			privateKeyPEM: rsaKey.PrivateKeyPKCS8,
			expected:      rsaKey.PublicKey,
		},
	}

	for description, test := range tests {
		t.Run(description, func(t *testing.T) {
			t.Parallel()
			got, err := PublicKeyFromPrivateKey(test.privateKeyPEM)
			if err != nil {
				t.Fatalf("PublicKeyFromPrivateKey() error = %v", err)
			}
			if got != test.expected {
				t.Errorf("PublicKeyFromPrivateKey() = %q, want %q", got, test.expected)
			}
		})
	}
}
