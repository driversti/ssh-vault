# ssh-vault Development Guidelines

Auto-generated from all feature plans. Last updated: 2026-03-22

## Active Technologies
- Go 1.26.1 + `golang.org/x/crypto/ssh` (SSH key parsing), standard library only (002-token-management)
- JSON file via `FileStore` (atomic write with temp file + rename, `sync.RWMutex` for concurrency) (002-token-management)
- Go 1.26.1 + `golang.org/x/crypto/ssh` (existing), standard library only (003-short-enrollment-url)
- Go 1.26.1 + Standard library only (existing `golang.org/x/crypto/ssh` unchanged) (004-self-serve-binaries)
- Filesystem (dist directory for binaries; existing FileStore for data) (004-self-serve-binaries)
- Go 1.26.1 + Standard library only (`embed`, `net/http`, `html/template`) (005-app-logo)
- N/A (embedded static asset) (005-app-logo)
- Go 1.26.1 (CLI binary) + POSIX shell (enrollment script generated server-side) + Standard library, `os/exec` for `ssh-keygen` invocation in Go, `ssh-keygen` CLI tool in shell scrip (006-auto-ssh-keygen)
- Filesystem (`~/.ssh/id_ed25519`, `~/.ssh/id_ed25519.pub`) (006-auto-ssh-keygen)
- Go 1.26.1 + Standard library (`net/http`, `html/template`, `strconv`, `math`), existing `golang.org/x/crypto/ssh` (008-paginated-audit-log)
- JSON file via `FileStore` (no changes to persistence layer) (008-paginated-audit-log)
- Go 1.26.1 + Standard library (`html/template`, `strings`), existing `golang.org/x/crypto/ssh` (009-device-public-keys)

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
- 009-device-public-keys: Added Go 1.26.1 + Standard library (`html/template`, `strings`), existing `golang.org/x/crypto/ssh`
- 008-paginated-audit-log: Added Go 1.26.1 + Standard library (`net/http`, `html/template`, `strconv`, `math`), existing `golang.org/x/crypto/ssh`
- 007-device-rename: Added Go 1.26.1 + `golang.org/x/crypto/ssh` (existing), standard library only


<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->
