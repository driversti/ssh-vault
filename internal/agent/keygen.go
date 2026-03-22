package agent

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// EnsureSSHKey checks whether the SSH private key at keyPath exists. If it does,
// it returns keyPath unchanged. If it does not, it calls promptFn to ask the caller
// whether to generate a new ed25519 key pair. If promptFn returns true (or is nil),
// the key is generated using ssh-keygen. Returns the key path or an error.
func EnsureSSHKey(keyPath string, promptFn func(string) bool) (string, error) {
	if _, err := os.Stat(keyPath); err == nil {
		return keyPath, nil
	}

	if promptFn != nil && !promptFn(fmt.Sprintf("SSH key not found at %s. Generate a new ed25519 key pair?", keyPath)) {
		return "", fmt.Errorf("SSH key not found at %s; generate one with: ssh-keygen -t ed25519", keyPath)
	}

	sshKeygenPath, err := exec.LookPath("ssh-keygen")
	if err != nil {
		return "", fmt.Errorf("ssh-keygen not found: %w", err)
	}

	dir := filepath.Dir(keyPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", fmt.Errorf("creating directory %s: %w", dir, err)
	}

	cmd := exec.Command(sshKeygenPath, "-t", "ed25519", "-f", keyPath, "-N", "", "-q")
	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("generating SSH key: %w\n%s", err, output)
	}

	fmt.Printf("Generated new SSH key pair: %s\n", keyPath)
	return keyPath, nil
}
