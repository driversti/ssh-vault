package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/driversti/ssh-vault/internal/agent"
)

func writeTestConfig(t *testing.T, dir string, cfg *agent.Config) string {
	t.Helper()
	path := filepath.Join(dir, "agent.json")
	if err := agent.SaveConfig(path, cfg); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}
	return path
}

func TestRunSync_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"keys":       []string{"ssh-ed25519 AAAA... test"},
			"updated_at": "2026-03-25T00:00:00Z",
		})
	}))
	defer srv.Close()

	dir := t.TempDir()
	cfgPath := writeTestConfig(t, dir, &agent.Config{
		HubURL:       srv.URL,
		APIToken:     "valid-token",
		AuthKeysPath: filepath.Join(dir, "authorized_keys"),
	})

	code := runSync([]string{"--config", cfgPath})
	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
}

func TestRunSync_MissingConfig(t *testing.T) {
	code := runSync([]string{"--config", "/nonexistent/path/agent.json"})
	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
}

func TestRunSync_EmptyAPIToken(t *testing.T) {
	dir := t.TempDir()
	cfgPath := writeTestConfig(t, dir, &agent.Config{
		HubURL:   "http://localhost",
		APIToken: "",
	})

	code := runSync([]string{"--config", cfgPath})
	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
}

func TestRunSync_HubUnreachable(t *testing.T) {
	dir := t.TempDir()
	cfgPath := writeTestConfig(t, dir, &agent.Config{
		HubURL:       "http://127.0.0.1:1", // unreachable
		APIToken:     "valid-token",
		AuthKeysPath: filepath.Join(dir, "authorized_keys"),
	})

	code := runSync([]string{"--config", cfgPath})
	if code != 2 {
		t.Errorf("exit code = %d, want 2", code)
	}
}

func TestRunSync_DeviceRevoked(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "device revoked"})
	}))
	defer srv.Close()

	dir := t.TempDir()
	cfgPath := writeTestConfig(t, dir, &agent.Config{
		HubURL:       srv.URL,
		APIToken:     "revoked-token",
		AuthKeysPath: filepath.Join(dir, "authorized_keys"),
	})

	code := runSync([]string{"--config", cfgPath})
	if code != 3 {
		t.Errorf("exit code = %d, want 3", code)
	}
}

func TestRunSync_CustomConfigFlag(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"keys":       []string{},
			"updated_at": "2026-03-25T00:00:00Z",
		})
	}))
	defer srv.Close()

	// Put config in a non-default location
	dir := t.TempDir()
	customDir := filepath.Join(dir, "custom", "location")
	os.MkdirAll(customDir, 0700)
	cfgPath := writeTestConfig(t, customDir, &agent.Config{
		HubURL:       srv.URL,
		APIToken:     "valid-token",
		AuthKeysPath: filepath.Join(dir, "authorized_keys"),
	})

	code := runSync([]string{"--config", cfgPath})
	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
}
