package ssh

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"strings"
	"testing"
)

func TestPrivateKeyToOpenSSH_PKCS8_Ed25519(t *testing.T) {
	// Generate ed25519 key
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate ed25519 key: %v", err)
	}

	// Marshal to PKCS#8 format
	pkcs8Bytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		t.Fatalf("Failed to marshal PKCS#8 key: %v", err)
	}

	pemBlock := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: pkcs8Bytes,
	}
	pemBytes := pem.EncodeToMemory(pemBlock)

	uuid := "test-uuid-123"
	result, err := PrivateKeyToOpenSSH(pemBytes, uuid)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify result is OpenSSH format
	if !strings.Contains(result, "-----BEGIN OPENSSH PRIVATE KEY-----") {
		t.Error("Result should contain OpenSSH header")
	}
	if !strings.Contains(result, "-----END OPENSSH PRIVATE KEY-----") {
		t.Error("Result should contain OpenSSH footer")
	}
}

func TestPrivateKeyToOpenSSH_PKCS8_RSA(t *testing.T) {
	// Generate RSA key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate RSA key: %v", err)
	}

	// Marshal to PKCS#8 format
	pkcs8Bytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		t.Fatalf("Failed to marshal PKCS#8 key: %v", err)
	}

	pemBlock := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: pkcs8Bytes,
	}
	pemBytes := pem.EncodeToMemory(pemBlock)

	uuid := "test-uuid-456"
	result, err := PrivateKeyToOpenSSH(pemBytes, uuid)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify result is OpenSSH format
	if !strings.Contains(result, "-----BEGIN OPENSSH PRIVATE KEY-----") {
		t.Error("Result should contain OpenSSH header")
	}
	if !strings.Contains(result, "-----END OPENSSH PRIVATE KEY-----") {
		t.Error("Result should contain OpenSSH footer")
	}
}

func TestPrivateKeyToOpenSSH_PKCS1_RSA(t *testing.T) {
	// Generate RSA key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate RSA key: %v", err)
	}

	// Marshal to PKCS#1 format
	pkcs1Bytes := x509.MarshalPKCS1PrivateKey(privateKey)

	pemBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: pkcs1Bytes,
	}
	pemBytes := pem.EncodeToMemory(pemBlock)

	uuid := "test-uuid-789"
	result, err := PrivateKeyToOpenSSH(pemBytes, uuid)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify result is OpenSSH format
	if !strings.Contains(result, "-----BEGIN OPENSSH PRIVATE KEY-----") {
		t.Error("Result should contain OpenSSH header")
	}
	if !strings.Contains(result, "-----END OPENSSH PRIVATE KEY-----") {
		t.Error("Result should contain OpenSSH footer")
	}
}

func TestPrivateKeyToOpenSSH_OpenSSH_Ed25519(t *testing.T) {
	// Generate ed25519 key
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate ed25519 key: %v", err)
	}

	// First convert to OpenSSH format using our function
	pkcs8Bytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		t.Fatalf("Failed to marshal PKCS#8 key: %v", err)
	}

	pemBlock := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: pkcs8Bytes,
	}
	pemBytes := pem.EncodeToMemory(pemBlock)

	uuid := "test-uuid-openssh-ed25519"
	openSSHKey, err := PrivateKeyToOpenSSH(pemBytes, uuid)
	if err != nil {
		t.Fatalf("Failed to convert to OpenSSH format: %v", err)
	}

	// Now test that passing the OpenSSH format back works (the fix)
	openSSHBytes := []byte(openSSHKey)
	result, err := PrivateKeyToOpenSSH(openSSHBytes, uuid)
	if err != nil {
		t.Fatalf("Expected no error when passing OpenSSH format, got: %v", err)
	}

	// Verify result is still OpenSSH format
	if !strings.Contains(result, "-----BEGIN OPENSSH PRIVATE KEY-----") {
		t.Error("Result should contain OpenSSH header")
	}
	if !strings.Contains(result, "-----END OPENSSH PRIVATE KEY-----") {
		t.Error("Result should contain OpenSSH footer")
	}
}

func TestPrivateKeyToOpenSSH_OpenSSH_RSA(t *testing.T) {
	// Generate RSA key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate RSA key: %v", err)
	}

	// First convert to OpenSSH format using our function
	pkcs8Bytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		t.Fatalf("Failed to marshal PKCS#8 key: %v", err)
	}

	pemBlock := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: pkcs8Bytes,
	}
	pemBytes := pem.EncodeToMemory(pemBlock)

	uuid := "test-uuid-openssh-rsa"
	openSSHKey, err := PrivateKeyToOpenSSH(pemBytes, uuid)
	if err != nil {
		t.Fatalf("Failed to convert to OpenSSH format: %v", err)
	}

	// Now test that passing the OpenSSH format back works (the fix)
	openSSHBytes := []byte(openSSHKey)
	result, err := PrivateKeyToOpenSSH(openSSHBytes, uuid)
	if err != nil {
		t.Fatalf("Expected no error when passing OpenSSH format, got: %v", err)
	}

	// Verify result is still OpenSSH format
	if !strings.Contains(result, "-----BEGIN OPENSSH PRIVATE KEY-----") {
		t.Error("Result should contain OpenSSH header")
	}
	if !strings.Contains(result, "-----END OPENSSH PRIVATE KEY-----") {
		t.Error("Result should contain OpenSSH footer")
	}
}

func TestPrivateKeyToOpenSSH_InvalidPEM(t *testing.T) {
	invalidPEM := []byte("not a valid PEM block")
	_, err := PrivateKeyToOpenSSH(invalidPEM, "test-uuid")
	if err == nil {
		t.Error("Expected error for invalid PEM, got nil")
	}
	if !strings.Contains(err.Error(), "invalid PEM private key") {
		t.Errorf("Expected error about invalid PEM, got: %v", err)
	}
}

func TestPrivateKeyToOpenSSH_UnsupportedKeyType(t *testing.T) {
	// Create a PEM block with unsupported type
	pemBlock := &pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: []byte("fake key data"),
	}
	pemBytes := pem.EncodeToMemory(pemBlock)

	_, err := PrivateKeyToOpenSSH(pemBytes, "test-uuid")
	if err == nil {
		t.Error("Expected error for unsupported key type, got nil")
	}
	if !strings.Contains(err.Error(), "unsupported key type") {
		t.Errorf("Expected error about unsupported key type, got: %v", err)
	}
}

func TestPrivateKeyToOpenSSH_ExtraDataAfterPEM(t *testing.T) {
	// Generate ed25519 key
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate ed25519 key: %v", err)
	}

	// Marshal to PKCS#8 format
	pkcs8Bytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		t.Fatalf("Failed to marshal PKCS#8 key: %v", err)
	}

	pemBlock := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: pkcs8Bytes,
	}
	pemBytes := pem.EncodeToMemory(pemBlock)
	// Add extra data after PEM block
	pemBytesWithExtra := append(pemBytes, []byte("extra data")...)

	_, err = PrivateKeyToOpenSSH(pemBytesWithExtra, "test-uuid")
	if err == nil {
		t.Error("Expected error for extra data after PEM, got nil")
	}
	if !strings.Contains(err.Error(), "more than just private key") {
		t.Errorf("Expected error about extra data, got: %v", err)
	}
}

func TestPrivateKeyToOpenSSH_UUIDConsistency(t *testing.T) {
	// Generate ed25519 key
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate ed25519 key: %v", err)
	}

	// Marshal to PKCS#8 format
	pkcs8Bytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		t.Fatalf("Failed to marshal PKCS#8 key: %v", err)
	}

	pemBlock := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: pkcs8Bytes,
	}
	pemBytes := pem.EncodeToMemory(pemBlock)

	uuid1 := "test-uuid-consistent-1"
	uuid2 := "test-uuid-consistent-2"

	result1, err := PrivateKeyToOpenSSH(pemBytes, uuid1)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	result2, err := PrivateKeyToOpenSSH(pemBytes, uuid2)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Results should be different due to UUID-based check bytes
	if result1 == result2 {
		t.Error("Results should differ for different UUIDs")
	}
}
