package agent

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/driversti/ssh-vault/internal/keyblock"
)

// keysResponse matches the hub's response to GET /api/keys.
type keysResponse struct {
	Keys      []string  `json:"keys"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Run starts the agent sync loop. It blocks until interrupted or an
// unrecoverable error occurs (e.g., device revoked).
func Run(cfg *Config) error {
	if cfg.APIToken == "" {
		return fmt.Errorf("no API token configured — run 'ssh-vault enroll' first and get approved")
	}

	slog.Info("starting sync agent",
		"hub", cfg.HubURL,
		"interval", cfg.Interval,
		"auth_keys", cfg.AuthKeysPath,
	)

	hasSynced := false
	hubDown := false

	// Run initial sync immediately
	if err := SyncOnce(cfg); err != nil {
		if isHubUnreachable(err) {
			slog.Info("hub not reachable — first sync pending, will retry", "error", err)
		} else {
			slog.Error("initial sync failed", "error", err)
		}
	} else {
		hasSynced = true
	}

	ticker := time.NewTicker(cfg.Interval)
	defer ticker.Stop()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-ticker.C:
			if err := SyncOnce(cfg); err != nil {
				// Revocation — unrecoverable
				if strings.Contains(err.Error(), "device revoked") {
					slog.Error("device has been revoked — stopping agent", "error", err)
					return fmt.Errorf("device revoked")
				}

				// Hub unreachable — log appropriately and do NOT modify authorized_keys
				if isHubUnreachable(err) {
					if !hubDown {
						hubDown = true
						if !hasSynced {
							slog.Info("hub not reachable — first sync pending, will retry")
						} else {
							slog.Warn("hub unreachable — retaining current authorized_keys", "error", err)
						}
					}
					continue
				}

				slog.Error("sync failed", "error", err)
			} else {
				if hubDown {
					slog.Info("hub connection restored — sync resumed")
					hubDown = false
				}
				hasSynced = true
			}
		case sig := <-sigCh:
			slog.Info("received signal, shutting down", "signal", sig)
			return nil
		}
	}
}

// SyncOnce performs a single sync: fetches keys from the hub and writes
// them to the authorized_keys file.
func SyncOnce(cfg *Config) error {
	keys, err := fetchKeys(cfg.HubURL, cfg.APIToken)
	if err != nil {
		return err
	}

	if err := keyblock.WriteBlock(cfg.AuthKeysPath, keys); err != nil {
		return fmt.Errorf("writing authorized_keys: %w", err)
	}

	slog.Info("sync complete", "keys", len(keys))
	return nil
}

// fetchKeys calls GET /api/keys on the hub and returns the key list.
func fetchKeys(hubURL, apiToken string) ([]string, error) {
	req, err := http.NewRequest("GET", strings.TrimRight(hubURL, "/")+"/api/keys", nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("hub unreachable: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusUnauthorized {
		var errResp map[string]string
		json.Unmarshal(body, &errResp)
		if errResp["error"] == "device revoked" {
			return nil, fmt.Errorf("device revoked")
		}
		return nil, fmt.Errorf("unauthorized (HTTP 401): %s", body)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected response (HTTP %d): %s", resp.StatusCode, body)
	}

	var keysResp keysResponse
	if err := json.Unmarshal(body, &keysResp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	return keysResp.Keys, nil
}

// isHubUnreachable checks if the error indicates the hub is not reachable.
func isHubUnreachable(err error) bool {
	return err != nil && strings.Contains(err.Error(), "hub unreachable")
}
