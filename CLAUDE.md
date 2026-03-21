# ssh-vault Development Guidelines

Auto-generated from all feature plans. Last updated: 2026-03-21

## Active Technologies
- Go 1.26.1 + `golang.org/x/crypto/ssh` (SSH key parsing), standard library only (002-token-management)
- JSON file via `FileStore` (atomic write with temp file + rename, `sync.RWMutex` for concurrency) (002-token-management)

- Go (latest stable) + standard library (001-ssh-key-sync-hub)
- `golang.org/x/crypto/ssh` for SSH key parsing

## Project Structure

```text
cmd/ssh-vault/         # Single binary entry point
internal/hub/          # Hub server (dashboard + API)
internal/agent/        # Sync agent + enrollment
internal/keyblock/     # authorized_keys file manipulation
internal/model/        # Shared types (Device, Token)
```

## Commands

```bash
go build -o ssh-vault ./cmd/ssh-vault   # Build
go test ./...                            # Run all tests
go vet ./...                             # Static analysis
golangci-lint run                        # Linting
```

## Code Style

Go: Follow Effective Go and Go Code Review Comments guidelines.
Use `gofmt`/`goimports`. Explicit error handling. Table-driven tests.

## Recent Changes
- 002-token-management: Added Go 1.26.1 + `golang.org/x/crypto/ssh` (SSH key parsing), standard library only

- 001-ssh-key-sync-hub: SSH key distribution hub with device agents

<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->
