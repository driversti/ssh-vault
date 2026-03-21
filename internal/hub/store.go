package hub

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/driversti/ssh-vault/internal/model"
)

// FileStore provides thread-safe file-backed storage for the hub.
type FileStore struct {
	mu   sync.RWMutex
	path string
	data model.Store
}

// NewFileStore creates a FileStore that persists to the given file path.
// If the file does not exist, it starts with an empty store.
func NewFileStore(path string) (*FileStore, error) {
	fs := &FileStore{path: path}
	if err := fs.Load(); err != nil {
		return nil, err
	}
	return fs, nil
}

// Load reads the store data from disk. If the file does not exist,
// it initializes an empty store.
func (fs *FileStore) Load() error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	raw, err := os.ReadFile(fs.path)
	if os.IsNotExist(err) {
		fs.data = model.Store{}
		return nil
	}
	if err != nil {
		return fmt.Errorf("reading store file: %w", err)
	}
	if err := json.Unmarshal(raw, &fs.data); err != nil {
		return fmt.Errorf("parsing store file: %w", err)
	}
	return nil
}

// Save writes the store data to disk atomically.
func (fs *FileStore) Save() error {
	raw, err := json.MarshalIndent(fs.data, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling store: %w", err)
	}

	dir := filepath.Dir(fs.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating data directory: %w", err)
	}

	tmp, err := os.CreateTemp(dir, ".tmp-store-*")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpPath := tmp.Name()

	if _, err := tmp.Write(raw); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("writing temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("closing temp file: %w", err)
	}
	if err := os.Chmod(tmpPath, 0600); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("setting permissions: %w", err)
	}
	if err := os.Rename(tmpPath, fs.path); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("renaming temp file: %w", err)
	}
	return nil
}

// AddDevice adds a device and persists the change.
func (fs *FileStore) AddDevice(d model.Device) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	fs.data.Devices = append(fs.data.Devices, d)
	return fs.Save()
}

// GetDevice returns a device by ID.
func (fs *FileStore) GetDevice(id string) (*model.Device, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	for i := range fs.data.Devices {
		if fs.data.Devices[i].ID == id {
			d := fs.data.Devices[i]
			return &d, nil
		}
	}
	return nil, fmt.Errorf("device not found: %s", id)
}

// UpdateDevice replaces the device with the matching ID and persists.
func (fs *FileStore) UpdateDevice(d model.Device) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	for i := range fs.data.Devices {
		if fs.data.Devices[i].ID == d.ID {
			fs.data.Devices[i] = d
			return fs.Save()
		}
	}
	return fmt.Errorf("device not found: %s", d.ID)
}

// ListDevices returns all devices.
func (fs *FileStore) ListDevices() []model.Device {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	result := make([]model.Device, len(fs.data.Devices))
	copy(result, fs.data.Devices)
	return result
}

// ListDevicesByStatus returns devices matching the given status.
func (fs *FileStore) ListDevicesByStatus(status string) []model.Device {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	var result []model.Device
	for _, d := range fs.data.Devices {
		if d.Status == status {
			result = append(result, d)
		}
	}
	return result
}

// GetDeviceByAPIToken returns the device with the given API token.
func (fs *FileStore) GetDeviceByAPIToken(token string) (*model.Device, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	for i := range fs.data.Devices {
		if fs.data.Devices[i].APIToken == token {
			d := fs.data.Devices[i]
			return &d, nil
		}
	}
	return nil, fmt.Errorf("device not found for token")
}

// AddToken adds an onboarding token and persists.
func (fs *FileStore) AddToken(t model.Token) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	fs.data.Tokens = append(fs.data.Tokens, t)
	return fs.Save()
}

// GetToken returns a token by value.
func (fs *FileStore) GetToken(value string) (*model.Token, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	for i := range fs.data.Tokens {
		if fs.data.Tokens[i].Value == value {
			t := fs.data.Tokens[i]
			return &t, nil
		}
	}
	return nil, fmt.Errorf("token not found: %s", value)
}

// UseToken marks a token as used by the given device ID and persists.
func (fs *FileStore) UseToken(value, deviceID string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	for i := range fs.data.Tokens {
		if fs.data.Tokens[i].Value == value {
			fs.data.Tokens[i].Used = true
			fs.data.Tokens[i].UsedBy = deviceID
			return fs.Save()
		}
	}
	return fmt.Errorf("token not found: %s", value)
}

// ListTokens returns all tokens.
func (fs *FileStore) ListTokens() []model.Token {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	result := make([]model.Token, len(fs.data.Tokens))
	copy(result, fs.data.Tokens)
	return result
}

// RemoveToken removes an unused token by value and persists.
// Returns an error if the token is not found or has already been used.
func (fs *FileStore) RemoveToken(value string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	for i := range fs.data.Tokens {
		if fs.data.Tokens[i].Value == value {
			if fs.data.Tokens[i].Used {
				return fmt.Errorf("cannot remove used token")
			}
			last := len(fs.data.Tokens) - 1
			fs.data.Tokens[i] = fs.data.Tokens[last]
			fs.data.Tokens = fs.data.Tokens[:last]
			return fs.Save()
		}
	}
	return fmt.Errorf("token not found: %s", value)
}

// PurgeExpiredTokens removes all expired tokens from storage and persists.
// Used tokens are retained as historical records. Returns the count of purged tokens.
func (fs *FileStore) PurgeExpiredTokens() (int, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	kept := fs.data.Tokens[:0]
	for _, t := range fs.data.Tokens {
		if !t.IsExpired() || t.Used {
			kept = append(kept, t)
		}
	}
	purged := len(fs.data.Tokens) - len(kept)
	if purged == 0 {
		return 0, nil
	}
	fs.data.Tokens = kept
	return purged, fs.Save()
}

// AddShortCode adds a short code and persists.
func (fs *FileStore) AddShortCode(sc model.ShortCode) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	fs.data.ShortCodes = append(fs.data.ShortCodes, sc)
	return fs.Save()
}

// GetShortCode returns a short code by its code value.
func (fs *FileStore) GetShortCode(code string) (*model.ShortCode, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	for i := range fs.data.ShortCodes {
		if fs.data.ShortCodes[i].Code == code {
			sc := fs.data.ShortCodes[i]
			return &sc, nil
		}
	}
	return nil, fmt.Errorf("short code not found: %s", code)
}

// UseShortCode marks a short code as used by the given IP and persists.
func (fs *FileStore) UseShortCode(code, ip string) (*model.ShortCode, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	for i := range fs.data.ShortCodes {
		if fs.data.ShortCodes[i].Code == code {
			sc := &fs.data.ShortCodes[i]
			if !sc.IsValid() {
				if sc.IsExpired() {
					return nil, fmt.Errorf("short code expired: %s", code)
				}
				return nil, fmt.Errorf("short code already used: %s", code)
			}
			sc.MarkUsed(ip)
			if err := fs.Save(); err != nil {
				return nil, err
			}
			result := *sc
			return &result, nil
		}
	}
	return nil, fmt.Errorf("short code not found: %s", code)
}

// ListShortCodes returns all short codes.
func (fs *FileStore) ListShortCodes() []model.ShortCode {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	result := make([]model.ShortCode, len(fs.data.ShortCodes))
	copy(result, fs.data.ShortCodes)
	return result
}

// PurgeExpiredShortCodes removes expired, unused short codes and their orphaned tokens.
// Returns the count of purged short codes.
func (fs *FileStore) PurgeExpiredShortCodes() (int, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	var kept []model.ShortCode
	var orphanedTokens []string

	for _, sc := range fs.data.ShortCodes {
		if sc.IsExpired() && !sc.Used {
			orphanedTokens = append(orphanedTokens, sc.TokenValue)
		} else {
			kept = append(kept, sc)
		}
	}

	purged := len(fs.data.ShortCodes) - len(kept)
	if purged == 0 {
		return 0, nil
	}

	fs.data.ShortCodes = kept

	// Clean up orphaned tokens (tokens linked to expired short codes that were never used)
	if len(orphanedTokens) > 0 {
		orphanSet := make(map[string]bool, len(orphanedTokens))
		for _, tv := range orphanedTokens {
			orphanSet[tv] = true
		}

		keptTokens := fs.data.Tokens[:0]
		for _, t := range fs.data.Tokens {
			if !orphanSet[t.Value] || t.Used {
				keptTokens = append(keptTokens, t)
			}
		}
		fs.data.Tokens = keptTokens
	}

	return purged, fs.Save()
}

// AddAuditEntry adds an audit log entry and persists.
func (fs *FileStore) AddAuditEntry(entry model.AuditEntry) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	fs.data.AuditLog = append(fs.data.AuditLog, entry)
	return fs.Save()
}

// ListAuditLog returns all audit entries.
func (fs *FileStore) ListAuditLog() []model.AuditEntry {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	result := make([]model.AuditEntry, len(fs.data.AuditLog))
	copy(result, fs.data.AuditLog)
	return result
}
