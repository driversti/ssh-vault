package agent

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Config holds the agent's runtime configuration.
type Config struct {
	HubURL       string        `json:"hub_url"`
	Interval     time.Duration `json:"-"`
	IntervalStr  string        `json:"interval"`
	KeyPath      string        `json:"key_path"`
	AuthKeysPath string        `json:"auth_keys_path"`
	APIToken     string        `json:"api_token"`
	DeviceID     string        `json:"device_id"`
}

// DefaultConfigPath returns the default path for the agent config file.
func DefaultConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("determining home directory: %w", err)
	}
	return filepath.Join(home, ".ssh-vault", "agent.json"), nil
}

// LoadConfig reads the agent config from the given path.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	if cfg.IntervalStr != "" {
		cfg.Interval, err = time.ParseDuration(cfg.IntervalStr)
		if err != nil {
			return nil, fmt.Errorf("parsing interval: %w", err)
		}
	}
	return &cfg, nil
}

// SaveConfig writes the agent config to the given path.
func SaveConfig(path string, cfg *Config) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	cfg.IntervalStr = cfg.Interval.String()
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	return nil
}
