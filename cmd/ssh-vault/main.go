package main

import (
	"bufio"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/driversti/ssh-vault/internal/agent"
	"github.com/driversti/ssh-vault/internal/hub"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "hub":
		if err := runHub(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "hub: %v\n", err)
			os.Exit(1)
		}
	case "agent":
		if err := runAgent(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "agent: %v\n", err)
			os.Exit(1)
		}
	case "enroll":
		if err := runEnroll(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "enroll: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage: ssh-vault <command> [flags]\n\n")
	fmt.Fprintf(os.Stderr, "Commands:\n")
	fmt.Fprintf(os.Stderr, "  hub      Start the hub server\n")
	fmt.Fprintf(os.Stderr, "  agent    Start the sync agent\n")
	fmt.Fprintf(os.Stderr, "  enroll   Enroll this device with a hub\n")
}

func runHub(args []string) error {
	fs := flag.NewFlagSet("hub", flag.ExitOnError)
	addr := fs.String("addr", ":8080", "Listen address")
	dataDir := fs.String("data", "./data", "Data directory")
	password := fs.String("password", "", "Dashboard password (or env VAULT_PASSWORD)")
	tlsCert := fs.String("tls-cert", "", "Path to TLS certificate file")
	tlsKey := fs.String("tls-key", "", "Path to TLS private key file")
	externalURL := fs.String("external-url", "", "Public URL for enrollment links (or env VAULT_EXTERNAL_URL)")
	distDir := fs.String("dist-dir", "", "Directory of pre-built binaries (or env VAULT_DIST_DIR)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *password == "" {
		*password = os.Getenv("VAULT_PASSWORD")
	}
	if *password == "" {
		return fmt.Errorf("password required: use --password or VAULT_PASSWORD env var")
	}
	if *externalURL == "" {
		*externalURL = os.Getenv("VAULT_EXTERNAL_URL")
	}
	if *distDir == "" {
		*distDir = os.Getenv("VAULT_DIST_DIR")
	}

	storePath := filepath.Join(*dataDir, "data.json")
	store, err := hub.NewFileStore(storePath)
	if err != nil {
		return fmt.Errorf("initializing store: %w", err)
	}

	srv := hub.NewServer(hub.ServerConfig{
		Store:       store,
		Password:    *password,
		Addr:        *addr,
		TLSCert:     *tlsCert,
		TLSKey:      *tlsKey,
		ExternalURL: *externalURL,
		DistDir:     *distDir,
	})

	return srv.ListenAndServe()
}

func runAgent(args []string) error {
	fs := flag.NewFlagSet("agent", flag.ExitOnError)
	hubURL := fs.String("hub-url", "", "Hub base URL")
	interval := fs.String("interval", "5m", "Sync interval")
	keyPath := fs.String("key", "~/.ssh/id_ed25519", "Path to SSH private key")
	authKeysPath := fs.String("auth-keys", "~/.ssh/authorized_keys", "Path to authorized_keys file")
	if err := fs.Parse(args); err != nil {
		return err
	}

	dur, err := time.ParseDuration(*interval)
	if err != nil {
		return fmt.Errorf("invalid interval: %w", err)
	}

	// Try to load saved config
	configPath, err := agent.DefaultConfigPath()
	if err != nil {
		return err
	}

	cfg, err := agent.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("loading agent config: %w (run 'ssh-vault enroll' first)", err)
	}

	// Track which flags were explicitly set by the user.
	explicitFlags := make(map[string]bool)
	fs.Visit(func(f *flag.Flag) { explicitFlags[f.Name] = true })

	// Override with CLI flags if provided; prefer saved config values
	// when flags are at their defaults.
	if *hubURL != "" {
		cfg.HubURL = *hubURL
	}
	if cfg.HubURL == "" {
		return fmt.Errorf("--hub-url is required")
	}
	if explicitFlags["interval"] || cfg.Interval == 0 {
		cfg.Interval = dur
	}
	if explicitFlags["key"] || cfg.KeyPath == "" {
		cfg.KeyPath = expandHome(*keyPath)
	}
	if explicitFlags["auth-keys"] || cfg.AuthKeysPath == "" {
		cfg.AuthKeysPath = expandHome(*authKeysPath)
	}

	return agent.Run(cfg)
}

func runEnroll(args []string) error {
	fs := flag.NewFlagSet("enroll", flag.ExitOnError)
	hubURL := fs.String("hub-url", "", "Hub base URL")
	token := fs.String("token", "", "Onboarding token")
	keyPath := fs.String("key", "~/.ssh/id_ed25519", "Path to SSH private key")
	name := fs.String("name", "", "Device display name (default: hostname)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *hubURL == "" {
		return fmt.Errorf("--hub-url is required")
	}
	if *token == "" {
		return fmt.Errorf("--token is required")
	}
	if *name == "" {
		h, err := os.Hostname()
		if err != nil {
			return fmt.Errorf("could not determine hostname: %w", err)
		}
		*name = h
	}

	resolvedKeyPath := expandHome(*keyPath)

	// Ensure SSH key exists; prompt user if missing
	resolvedKeyPath, err := agent.EnsureSSHKey(resolvedKeyPath, func(msg string) bool {
		fmt.Printf("%s (y/n): ", msg)
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			return strings.TrimSpace(strings.ToLower(scanner.Text())) == "y"
		}
		return false
	})
	if err != nil {
		return err
	}

	cfg, err := agent.Enroll(agent.EnrollConfig{
		HubURL:  *hubURL,
		Token:   *token,
		KeyPath: resolvedKeyPath,
		Name:    *name,
	})
	if err != nil {
		return err
	}

	// Save the enrollment result
	configPath, err := agent.DefaultConfigPath()
	if err != nil {
		return err
	}
	if err := agent.SaveConfig(configPath, cfg); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	// Verify the saved config is complete (read-back check).
	saved, err := agent.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("verifying saved config: %w", err)
	}
	if saved.APIToken == "" {
		return fmt.Errorf("config verification failed: API token was not persisted to %s", configPath)
	}

	fmt.Printf("Config saved to %s\n", configPath)
	return nil
}

// expandHome replaces a leading ~ with the user's home directory.
func expandHome(path string) string {
	if path == "~" || len(path) > 1 && path[:2] == "~/" {
		home, err := os.UserHomeDir()
		if err != nil {
			slog.Warn("could not determine home directory, using path as-is", "path", path, "error", err)
			return path
		}
		return filepath.Join(home, path[1:])
	}
	return path
}
