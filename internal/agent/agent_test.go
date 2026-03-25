package agent

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/driversti/ssh-vault/internal/keyblock"
)

func TestSyncOnce_HubUp(t *testing.T) {
	// Mock hub server
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		json.NewEncoder(w).Encode(keysResponse{
			Keys: []string{"ssh-ed25519 AAAA... device-a", "ssh-ed25519 BBBB... device-b"},
		})
	}))
	defer srv.Close()

	authKeysPath := filepath.Join(t.TempDir(), "authorized_keys")
	cfg := &Config{
		HubURL:       srv.URL,
		APIToken:     "test-token",
		AuthKeysPath: authKeysPath,
	}

	if err := SyncOnce(cfg); err != nil {
		t.Fatalf("syncOnce: %v", err)
	}

	// Verify keys were written
	keys, err := keyblock.ReadBlock(authKeysPath)
	if err != nil {
		t.Fatalf("ReadBlock: %v", err)
	}
	if len(keys) != 2 {
		t.Errorf("expected 2 keys, got %d", len(keys))
	}
}

func TestSyncOnce_HubUnreachable_DoesNotModifyFile(t *testing.T) {
	authKeysPath := filepath.Join(t.TempDir(), "authorized_keys")

	// Write initial content
	existing := "ssh-rsa AAAA... existing-key\n" +
		keyblock.BlockBegin + "\n" +
		"ssh-ed25519 CCCC... managed-key\n" +
		keyblock.BlockEnd + "\n"
	os.WriteFile(authKeysPath, []byte(existing), 0600)

	cfg := &Config{
		HubURL:       "http://127.0.0.1:1", // unreachable
		APIToken:     "test-token",
		AuthKeysPath: authKeysPath,
	}

	err := SyncOnce(cfg)
	if err == nil {
		t.Fatal("expected error for unreachable hub")
	}
	if !isHubUnreachable(err) {
		t.Logf("error is not hub-unreachable type: %v", err)
	}

	// File should be unchanged
	data, _ := os.ReadFile(authKeysPath)
	if string(data) != existing {
		t.Error("authorized_keys should not be modified when hub is unreachable")
	}
}

func TestSyncOnce_DeviceRevoked(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "device revoked"})
	}))
	defer srv.Close()

	cfg := &Config{
		HubURL:       srv.URL,
		APIToken:     "revoked-token",
		AuthKeysPath: filepath.Join(t.TempDir(), "authorized_keys"),
	}

	err := SyncOnce(cfg)
	if err == nil {
		t.Fatal("expected error for revoked device")
	}
	if err.Error() != "device revoked" {
		t.Errorf("error = %q, want 'device revoked'", err.Error())
	}
}

func TestSyncOnce_HubRecovers(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		json.NewEncoder(w).Encode(keysResponse{
			Keys: []string{"ssh-ed25519 AAAA... device-a"},
		})
	}))
	defer srv.Close()

	authKeysPath := filepath.Join(t.TempDir(), "authorized_keys")
	cfg := &Config{
		HubURL:       srv.URL,
		APIToken:     "test-token",
		AuthKeysPath: authKeysPath,
	}

	// First sync works
	if err := SyncOnce(cfg); err != nil {
		t.Fatalf("first sync: %v", err)
	}

	// Second sync also works (simulating recovery)
	if err := SyncOnce(cfg); err != nil {
		t.Fatalf("second sync: %v", err)
	}

	if callCount != 2 {
		t.Errorf("expected 2 calls, got %d", callCount)
	}

	keys, _ := keyblock.ReadBlock(authKeysPath)
	if len(keys) != 1 {
		t.Errorf("expected 1 key, got %d", len(keys))
	}
}

func TestFetchKeys_InvalidToken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid token"})
	}))
	defer srv.Close()

	_, err := fetchKeys(srv.URL, "bad-token")
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
}
