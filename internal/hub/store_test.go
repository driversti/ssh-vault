package hub

import (
	"fmt"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/driversti/ssh-vault/internal/model"
)

func tempStorePath(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	return filepath.Join(dir, "data.json")
}

func TestFileStore_NewEmpty(t *testing.T) {
	fs, err := NewFileStore(tempStorePath(t))
	if err != nil {
		t.Fatalf("NewFileStore: %v", err)
	}
	if len(fs.ListDevices()) != 0 {
		t.Error("expected empty device list")
	}
	if len(fs.ListTokens()) != 0 {
		t.Error("expected empty token list")
	}
}

func TestFileStore_DeviceCRUD(t *testing.T) {
	path := tempStorePath(t)
	fs, err := NewFileStore(path)
	if err != nil {
		t.Fatalf("NewFileStore: %v", err)
	}

	d := model.Device{
		ID:        "dev-1",
		Name:      "test-device",
		PublicKey: "ssh-ed25519 AAAA...",
		Status:    model.StatusPending,
	}

	// Add
	if err := fs.AddDevice(d); err != nil {
		t.Fatalf("AddDevice: %v", err)
	}

	// Get
	got, err := fs.GetDevice("dev-1")
	if err != nil {
		t.Fatalf("GetDevice: %v", err)
	}
	if got.Name != "test-device" {
		t.Errorf("Name = %q, want %q", got.Name, "test-device")
	}

	// Update
	d.Status = model.StatusApproved
	if err := fs.UpdateDevice(d); err != nil {
		t.Fatalf("UpdateDevice: %v", err)
	}
	got, _ = fs.GetDevice("dev-1")
	if got.Status != model.StatusApproved {
		t.Errorf("Status = %q, want %q", got.Status, model.StatusApproved)
	}

	// ListDevicesByStatus
	approved := fs.ListDevicesByStatus(model.StatusApproved)
	if len(approved) != 1 {
		t.Errorf("ListDevicesByStatus(approved) = %d, want 1", len(approved))
	}
	pending := fs.ListDevicesByStatus(model.StatusPending)
	if len(pending) != 0 {
		t.Errorf("ListDevicesByStatus(pending) = %d, want 0", len(pending))
	}
}

func TestFileStore_TokenCRUD(t *testing.T) {
	fs, err := NewFileStore(tempStorePath(t))
	if err != nil {
		t.Fatalf("NewFileStore: %v", err)
	}

	tok := model.Token{
		Value:     "test-token-value",
		CreatedAt: time.Now().UTC(),
		ExpiresAt: time.Now().UTC().Add(24 * time.Hour),
	}

	if err := fs.AddToken(tok); err != nil {
		t.Fatalf("AddToken: %v", err)
	}

	got, err := fs.GetToken("test-token-value")
	if err != nil {
		t.Fatalf("GetToken: %v", err)
	}
	if got.Used {
		t.Error("token should not be used")
	}

	if err := fs.UseToken("test-token-value", "dev-1"); err != nil {
		t.Fatalf("UseToken: %v", err)
	}

	got, _ = fs.GetToken("test-token-value")
	if !got.Used {
		t.Error("token should be used")
	}
	if got.UsedBy != "dev-1" {
		t.Errorf("UsedBy = %q, want %q", got.UsedBy, "dev-1")
	}
}

func TestFileStore_LoadSaveRoundTrip(t *testing.T) {
	path := tempStorePath(t)
	fs1, err := NewFileStore(path)
	if err != nil {
		t.Fatalf("NewFileStore: %v", err)
	}

	if err := fs1.AddDevice(model.Device{ID: "d1", Name: "dev1", Status: model.StatusPending}); err != nil {
		t.Fatalf("AddDevice: %v", err)
	}
	if err := fs1.AddAuditEntry(model.NewAuditEntry(model.EventEnrolled, "d1", "test")); err != nil {
		t.Fatalf("AddAuditEntry: %v", err)
	}

	// Load into a new FileStore
	fs2, err := NewFileStore(path)
	if err != nil {
		t.Fatalf("NewFileStore reload: %v", err)
	}
	devices := fs2.ListDevices()
	if len(devices) != 1 || devices[0].ID != "d1" {
		t.Errorf("reloaded devices = %+v, want 1 device with ID d1", devices)
	}
	audit := fs2.ListAuditLog()
	if len(audit) != 1 {
		t.Errorf("reloaded audit log length = %d, want 1", len(audit))
	}
}

func TestFileStore_ConcurrentAccess(t *testing.T) {
	fs, err := NewFileStore(tempStorePath(t))
	if err != nil {
		t.Fatalf("NewFileStore: %v", err)
	}

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			d := model.Device{
				ID:     fmt.Sprintf("dev-%d", n),
				Name:   fmt.Sprintf("device-%d", n),
				Status: model.StatusPending,
			}
			if err := fs.AddDevice(d); err != nil {
				t.Errorf("AddDevice(%d): %v", n, err)
			}
		}(i)
	}
	wg.Wait()

	devices := fs.ListDevices()
	if len(devices) != 10 {
		t.Errorf("device count = %d, want 10", len(devices))
	}
}

func TestFileStore_GetDevice_NotFound(t *testing.T) {
	fs, err := NewFileStore(tempStorePath(t))
	if err != nil {
		t.Fatalf("NewFileStore: %v", err)
	}
	_, err = fs.GetDevice("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent device")
	}
}

func TestFileStore_GetDeviceByAPIToken(t *testing.T) {
	fs, err := NewFileStore(tempStorePath(t))
	if err != nil {
		t.Fatalf("NewFileStore: %v", err)
	}

	d := model.Device{
		ID:       "dev-1",
		Name:     "test",
		Status:   model.StatusApproved,
		APIToken: "secret-token-123",
	}
	if err := fs.AddDevice(d); err != nil {
		t.Fatalf("AddDevice: %v", err)
	}

	got, err := fs.GetDeviceByAPIToken("secret-token-123")
	if err != nil {
		t.Fatalf("GetDeviceByAPIToken: %v", err)
	}
	if got.ID != "dev-1" {
		t.Errorf("ID = %q, want %q", got.ID, "dev-1")
	}

	_, err = fs.GetDeviceByAPIToken("wrong-token")
	if err == nil {
		t.Error("expected error for wrong token")
	}
}

func TestFileStore_RemoveToken(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(fs *FileStore)
		removeVal string
		wantErr   string
	}{
		{
			name: "remove valid unused token",
			setup: func(fs *FileStore) {
				fs.AddToken(model.Token{
					Value:     "tok-to-remove",
					ExpiresAt: time.Now().UTC().Add(1 * time.Hour),
				})
			},
			removeVal: "tok-to-remove",
		},
		{
			name: "remove used token returns error",
			setup: func(fs *FileStore) {
				fs.AddToken(model.Token{
					Value:     "tok-used",
					ExpiresAt: time.Now().UTC().Add(1 * time.Hour),
					Used:      true,
					UsedBy:    "dev-1",
				})
			},
			removeVal: "tok-used",
			wantErr:   "cannot remove used token",
		},
		{
			name:      "remove nonexistent token returns error",
			setup:     func(fs *FileStore) {},
			removeVal: "nonexistent",
			wantErr:   "token not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs, err := NewFileStore(tempStorePath(t))
			if err != nil {
				t.Fatalf("NewFileStore: %v", err)
			}
			tt.setup(fs)

			err = fs.RemoveToken(tt.removeVal)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				if !contains(err.Error(), tt.wantErr) {
					t.Fatalf("error = %q, want containing %q", err.Error(), tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("RemoveToken: %v", err)
			}

			// Verify token is gone
			_, err = fs.GetToken(tt.removeVal)
			if err == nil {
				t.Error("expected error after removing token, got nil")
			}
		})
	}
}

func TestFileStore_RemoveToken_PreservesOthers(t *testing.T) {
	fs, err := NewFileStore(tempStorePath(t))
	if err != nil {
		t.Fatalf("NewFileStore: %v", err)
	}

	for _, v := range []string{"tok-a", "tok-b", "tok-c"} {
		fs.AddToken(model.Token{
			Value:     v,
			ExpiresAt: time.Now().UTC().Add(1 * time.Hour),
		})
	}

	if err := fs.RemoveToken("tok-b"); err != nil {
		t.Fatalf("RemoveToken: %v", err)
	}

	// tok-a and tok-c should remain
	for _, v := range []string{"tok-a", "tok-c"} {
		if _, err := fs.GetToken(v); err != nil {
			t.Errorf("GetToken(%q) should exist after removing tok-b: %v", v, err)
		}
	}
	if _, err := fs.GetToken("tok-b"); err == nil {
		t.Error("tok-b should not exist after removal")
	}
}

func TestFileStore_PurgeExpiredTokens(t *testing.T) {
	tests := []struct {
		name      string
		tokens    []model.Token
		wantPurge int
		wantLeft  int
	}{
		{
			name: "purges expired, keeps valid",
			tokens: []model.Token{
				{Value: "valid", ExpiresAt: time.Now().UTC().Add(1 * time.Hour)},
				{Value: "expired", ExpiresAt: time.Now().UTC().Add(-1 * time.Second)},
			},
			wantPurge: 1,
			wantLeft:  1,
		},
		{
			name: "keeps used tokens even if expired",
			tokens: []model.Token{
				{Value: "used-expired", ExpiresAt: time.Now().UTC().Add(-1 * time.Second), Used: true, UsedBy: "dev-1"},
				{Value: "valid", ExpiresAt: time.Now().UTC().Add(1 * time.Hour)},
			},
			wantPurge: 0,
			wantLeft:  2,
		},
		{
			name: "no expired tokens",
			tokens: []model.Token{
				{Value: "a", ExpiresAt: time.Now().UTC().Add(1 * time.Hour)},
				{Value: "b", ExpiresAt: time.Now().UTC().Add(2 * time.Hour)},
			},
			wantPurge: 0,
			wantLeft:  2,
		},
		{
			name: "all expired unused",
			tokens: []model.Token{
				{Value: "x", ExpiresAt: time.Now().UTC().Add(-1 * time.Hour)},
				{Value: "y", ExpiresAt: time.Now().UTC().Add(-2 * time.Hour)},
			},
			wantPurge: 2,
			wantLeft:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs, err := NewFileStore(tempStorePath(t))
			if err != nil {
				t.Fatalf("NewFileStore: %v", err)
			}
			for _, tok := range tt.tokens {
				fs.AddToken(tok)
			}

			purged, err := fs.PurgeExpiredTokens()
			if err != nil {
				t.Fatalf("PurgeExpiredTokens: %v", err)
			}
			if purged != tt.wantPurge {
				t.Errorf("purged = %d, want %d", purged, tt.wantPurge)
			}
			left := fs.ListTokens()
			if len(left) != tt.wantLeft {
				t.Errorf("tokens remaining = %d, want %d", len(left), tt.wantLeft)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
