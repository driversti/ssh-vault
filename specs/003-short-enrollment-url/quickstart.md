# Quickstart: Short Enrollment URL

**Feature**: 003-short-enrollment-url
**Date**: 2026-03-21

## Build & Run

```bash
# Build
go build -o ssh-vault ./cmd/ssh-vault

# Run hub with short enrollment URL support
./ssh-vault hub \
  -addr :8443 \
  -password "your-admin-password" \
  -tls-cert cert.pem \
  -tls-key key.pem \
  -external-url https://ssh-vault.yurii.live \
  -github-repo driversti/ssh-vault \
  -release-tag v1.0.0
```

## Admin Workflow

1. Log into the hub dashboard at `https://ssh-vault.yurii.live/tokens`
2. Click **"Generate Enrollment Link"** in the Quick Enrollment section
3. Copy the displayed command: `curl -sSL https://ssh-vault.yurii.live/e/482917 | sh`
4. Share the command with the user (valid for 15 minutes)

## User Enrollment

```bash
# On the target device — single command, no arguments needed
curl -sSL https://ssh-vault.yurii.live/e/482917 | sh
```

The script will:
- Detect the device's OS and architecture
- Download the ssh-vault binary from GitHub Releases
- Verify the binary checksum
- Detect hostname and SSH key automatically
- Complete the enrollment handshake

The device will appear as "pending" on the admin dashboard, awaiting approval.

## Test

```bash
go test ./...
go vet ./...
```

## Key Files

| File | Purpose |
|------|---------|
| `internal/model/shortcode.go` | ShortCode struct and validation |
| `internal/hub/handlers_shortcode.go` | `/e/{code}` handler + script template |
| `internal/hub/ratelimit.go` | Per-IP rate limiter |
| `internal/hub/store.go` | ShortCode persistence (AddShortCode, GetShortCode, etc.) |
| `internal/hub/templates/tokens.html` | Dashboard UI for generating enrollment links |
