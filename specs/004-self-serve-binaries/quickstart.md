# Quickstart: Self-Serve Binary Distribution

**Feature**: 004-self-serve-binaries
**Date**: 2026-03-21

## Build & Prepare

```bash
# Build for all supported platforms
GOOS=linux GOARCH=amd64 go build -o dist/ssh-vault_linux_amd64 ./cmd/ssh-vault
GOOS=linux GOARCH=arm64 go build -o dist/ssh-vault_linux_arm64 ./cmd/ssh-vault
GOOS=darwin GOARCH=amd64 go build -o dist/ssh-vault_darwin_amd64 ./cmd/ssh-vault
GOOS=darwin GOARCH=arm64 go build -o dist/ssh-vault_darwin_arm64 ./cmd/ssh-vault

# Optional: generate checksums
cd dist && sha256sum ssh-vault-* > checksums.txt && cd ..
```

## Run Hub

```bash
./dist/ssh-vault_$(uname -s | tr A-Z a-z)_$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/') hub \
  -addr :8443 \
  -password "your-admin-password" \
  -tls-cert cert.pem \
  -tls-key key.pem \
  -external-url https://ssh-vault.yurii.live \
  -dist-dir ./dist
```

## Admin Workflow

1. Log into the hub dashboard at `https://ssh-vault.yurii.live/tokens`
2. Click **"Generate Enrollment Link"**
3. Copy the displayed command: `curl -sSL https://ssh-vault.yurii.live/e/482917 | sh`
4. Share the command with the user (valid for 15 minutes)

## User Enrollment

```bash
# On the target device — single command, no arguments needed
curl -sSL https://ssh-vault.yurii.live/e/482917 | sh
```

The script will:
- Detect the device's OS and architecture
- Download the ssh-vault binary **from the hub** (not GitHub)
- Verify the binary checksum (if checksums.txt is available)
- Detect hostname and SSH key automatically
- Complete the enrollment handshake

## Docker Compose

```yaml
services:
  hub:
    image: ghcr.io/driversti/ssh-vault:latest
    volumes:
      - ./data:/data
      - ./dist:/dist:ro
    environment:
      - VAULT_PASSWORD=${VAULT_PASSWORD}
      - VAULT_EXTERNAL_URL=${VAULT_EXTERNAL_URL}
      - VAULT_DIST_DIR=/dist
```

## Test

```bash
go test ./...
go vet ./...

# Manual: verify binary download
curl -fsSL -o /tmp/ssh-vault http://localhost:8080/download/linux/amd64
file /tmp/ssh-vault  # Should show ELF binary
```

## Key Files

| File | Purpose |
|------|---------|
| `cmd/ssh-vault/main.go` | `--dist-dir` flag (replaces `--github-repo` / `--release-tag`) |
| `internal/hub/server.go` | Server struct with `distDir` field |
| `internal/hub/handlers_download.go` | `GET /download/{os}/{arch}` handler |
| `internal/hub/handlers_shortcode.go` | Updated enrollment script template |
| `docker-compose.yml` | Dist volume mount + env var |
