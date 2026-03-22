package agent

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureSSHKey(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(t *testing.T, dir string) string // returns keyPath
		promptFn  func(string) bool
		wantErr   bool
		wantFile  bool // expect key file to exist after call
	}{
		{
			name: "key already exists",
			setup: func(t *testing.T, dir string) string {
				t.Helper()
				keyPath := filepath.Join(dir, "id_ed25519")
				if err := os.WriteFile(keyPath, []byte("existing-key"), 0600); err != nil {
					t.Fatal(err)
				}
				return keyPath
			},
			promptFn: nil,
			wantErr:  false,
			wantFile: true,
		},
		{
			name: "key missing and prompt accepts",
			setup: func(t *testing.T, dir string) string {
				t.Helper()
				return filepath.Join(dir, "id_ed25519")
			},
			promptFn: func(string) bool { return true },
			wantErr:  false,
			wantFile: true,
		},
		{
			name: "key missing and nil prompt auto-accepts",
			setup: func(t *testing.T, dir string) string {
				t.Helper()
				return filepath.Join(dir, "id_ed25519")
			},
			promptFn: nil,
			wantErr:  false,
			wantFile: true,
		},
		{
			name: "key missing and prompt declines",
			setup: func(t *testing.T, dir string) string {
				t.Helper()
				return filepath.Join(dir, "id_ed25519")
			},
			promptFn: func(string) bool { return false },
			wantErr:  true,
			wantFile: false,
		},
		{
			name: "creates parent directory",
			setup: func(t *testing.T, dir string) string {
				t.Helper()
				return filepath.Join(dir, "subdir", "id_ed25519")
			},
			promptFn: nil,
			wantErr:  false,
			wantFile: true,
		},
		{
			name: "never overwrites existing key",
			setup: func(t *testing.T, dir string) string {
				t.Helper()
				keyPath := filepath.Join(dir, "id_ed25519")
				if err := os.WriteFile(keyPath, []byte("original-content"), 0600); err != nil {
					t.Fatal(err)
				}
				return keyPath
			},
			promptFn: nil,
			wantErr:  false,
			wantFile: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			keyPath := tt.setup(t, dir)

			// Save original content if file exists (for overwrite check)
			var originalContent []byte
			if data, err := os.ReadFile(keyPath); err == nil {
				originalContent = data
			}

			result, err := EnsureSSHKey(keyPath, tt.promptFn)
			if (err != nil) != tt.wantErr {
				t.Fatalf("EnsureSSHKey() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantFile {
				if _, err := os.Stat(keyPath); os.IsNotExist(err) {
					t.Fatalf("expected key file to exist at %s", keyPath)
				}
				if result != keyPath {
					t.Fatalf("expected result %q, got %q", keyPath, result)
				}
			}

			// Verify existing keys are not overwritten
			if originalContent != nil {
				data, err := os.ReadFile(keyPath)
				if err != nil {
					t.Fatalf("reading key file: %v", err)
				}
				if string(data) != string(originalContent) {
					t.Fatal("existing key was overwritten")
				}
			}
		})
	}
}
