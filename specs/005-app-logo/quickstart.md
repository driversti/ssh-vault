# Quickstart: App Logo Integration

## What This Feature Does

Adds the SSH Vault logo as both a browser tab favicon and a visual element in the header navigation bar.

## How to Verify

1. Build and run the hub server:
   ```bash
   go build -o ssh-vault ./cmd/ssh-vault
   ./ssh-vault hub --admin-password=test
   ```

2. Open `http://localhost:8080` in a browser

3. Check:
   - Browser tab shows the lock logo icon (not the default blank page icon)
   - Header shows the logo image to the left of "SSH Vault" text
   - Navigate to different pages (Devices, Tokens, Audit) — logo persists everywhere

## Files Changed

- `internal/hub/templates/logo.svg` — Logo asset (copied from repo root)
- `internal/hub/server.go` — New handler + route for `/static/logo.svg`
- `internal/hub/templates/layout.html` — Favicon link + header image tag
