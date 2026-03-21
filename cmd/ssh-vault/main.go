package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
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

	// Override with CLI flags if provided
	if *hubURL != "" {
		cfg.HubURL = *hubURL
	}
	if cfg.HubURL == "" {
		return fmt.Errorf("--hub-url is required")
	}
	cfg.Interval = dur
	cfg.KeyPath = *keyPath
	cfg.AuthKeysPath = *authKeysPath

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

	cfg, err := agent.Enroll(agent.EnrollConfig{
		HubURL:  *hubURL,
		Token:   *token,
		KeyPath: *keyPath,
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

	fmt.Printf("Config saved to %s\n", configPath)
	return nil
}
