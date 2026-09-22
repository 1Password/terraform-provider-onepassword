package sdk

import (
	"errors"
	"strings"
	"testing"
)

func TestWithServiceAccountVaultHint(t *testing.T) {
	baseErr := errors.New("permission denied")
	limitation := "service accounts can only manage vaults they created"

	serviceAccountClient := &Client{config: SDKConfig{ServiceAccountToken: "token"}}
	err := serviceAccountClient.withServiceAccountVaultHint(baseErr, limitation)
	if !errors.Is(err, baseErr) {
		t.Fatal("Expected the original error to remain wrapped")
	}
	if !strings.Contains(err.Error(), "Service account limitation: "+limitation) {
		t.Fatalf("Expected service account limitation in error, got %q", err)
	}

	desktopClient := &Client{config: SDKConfig{Account: "account"}}
	if err := desktopClient.withServiceAccountVaultHint(baseErr, limitation); err != baseErr {
		t.Fatalf("Desktop authentication should return the original error, got %q", err)
	}
}
