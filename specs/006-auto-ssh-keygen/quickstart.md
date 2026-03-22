# Quickstart: Auto SSH Key Generation During Enrollment

## What This Feature Does

Automatically generates an SSH key pair during enrollment when no keys exist, eliminating the manual `ssh-keygen` step that previously broke the enrollment flow.

## How to Verify

### Enrollment Script (US1 — primary path)

1. Build and run the hub:
   ```bash
   go build -o ssh-vault ./cmd/ssh-vault
   ./ssh-vault hub --admin-password=test
   ```

2. Generate an enrollment token and short URL via the hub dashboard

3. On a test machine (or Docker container) **with no SSH keys**:
   ```bash
   # Verify no keys exist
   ls ~/.ssh/id_*.pub   # should show "No such file or directory"

   # Run enrollment — should auto-generate keys and succeed
   curl -sSL http://hub-address/e/CODE | sh
   ```

4. Check:
   - Output includes a message about generating a new SSH key pair
   - `~/.ssh/id_ed25519` and `~/.ssh/id_ed25519.pub` exist with correct permissions
   - Enrollment completes successfully

### CLI Enroll Command (US2 — secondary path)

1. On a machine with no SSH keys:
   ```bash
   ./ssh-vault enroll --hub-url http://hub-address --token TOKEN
   ```

2. Check:
   - Command prompts "No SSH key found. Generate one? (y/n)"
   - Answering "y" generates the key and proceeds with enrollment
   - Answering "n" exits with a clear message

### Existing Keys (regression check)

1. On a machine **with** existing SSH keys, run either enrollment path
2. Verify: Existing keys are used unchanged, no overwrite prompt or generation

## Files Changed

- `internal/hub/handlers_shortcode.go` — Shell script: auto-generate keys instead of error
- `internal/agent/keygen.go` — New: `EnsureSSHKey()` function
- `internal/agent/keygen_test.go` — New: Tests for key generation logic
- `cmd/ssh-vault/main.go` — CLI: call `EnsureSSHKey()` before enrollment
