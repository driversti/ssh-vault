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
