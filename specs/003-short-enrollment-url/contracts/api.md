# API Contracts: Short Enrollment URL

**Feature**: 003-short-enrollment-url
**Date**: 2026-03-21

## New Endpoints

### GET /e/{code} — Serve Enrollment Script

**Auth**: None (public, rate-limited)
**Rate limit**: 10 requests/minute per IP

**Success Response** (200 OK):
- Content-Type: `text/plain; charset=utf-8`
- Body: POSIX shell script (dynamically generated)

**Error Responses**:

| Status | Condition | Body |
|--------|-----------|------|
| 404 | Code not found or already used | `echo "Error: Invalid or expired enrollment code."; exit 1` |
| 410 | Code expired | `echo "Error: This enrollment code has expired."; exit 1` |
| 429 | Rate limit exceeded | `echo "Error: Too many requests. Please wait and try again."; exit 1` |

**Notes**:
- Error responses are also valid shell scripts (echo + exit) so that `curl | sh` displays the error gracefully instead of failing silently.
- The code is consumed (marked used) only when the enrollment script initiates the actual enrollment POST, not when the script is downloaded. This allows retries if the script fails mid-execution.

---

### POST /tokens/generate-link — Generate Short Enrollment Code

**Auth**: Session cookie (requireSession middleware)

**Request**: Form POST (no body parameters needed — TTL uses server default of 15 minutes)

**Success Response** (302 Redirect to `/tokens`):
- Sets flash message (or query param) with the generated code and curl command

**Side Effects**:
1. Creates a new Token (24h TTL)
2. Creates a new ShortCode (15min TTL) linked to the token
3. Adds `shortcode_created` audit entry

**Error Responses**:

| Status | Condition | Body |
|--------|-----------|------|
| 500 | Code generation failure (e.g., uniqueness exhaustion) | Redirect to `/tokens` with error message |

---

## Enrollment Script Contract

The dynamically generated shell script performs these steps in order:

```
1. set -euo pipefail
2. Detect OS (uname -s → linux/darwin) and arch (uname -m → amd64/arm64)
3. Map to download filename: ssh-vault_{os}_{arch}
4. Verify curl/wget is available
5. Download binary from GitHub Releases URL (provided by hub template)
6. Download checksums.txt from GitHub Release, verify binary SHA-256
7. chmod +x the binary
8. Detect hostname → device name
9. Find first SSH public key in ~/.ssh/id_*.pub
10. If no key found → echo error, exit 1
11. Run: ./ssh-vault enroll --hub-url {HUB_URL} --token {TOKEN} --key {KEY_PATH} --name {HOSTNAME}
12. Print success message with next steps (await admin approval)
13. Clean up temporary files on exit (trap)
```

**Template Variables** (injected by the hub handler):

| Variable | Example | Source |
|----------|---------|--------|
| `HubURL` | `https://ssh-vault.yurii.live` | Hub's own external URL (from config) |
| `Token` | `a1b2c3d4...` (64-hex) | The linked enrollment token value |
| `DownloadBaseURL` | `https://github.com/driversti/ssh-vault/releases/download/v1.0.0` | From hub config (repo + tag) |

---

## Modified Endpoints

### GET/POST /tokens — Token Management Page

**Changes**:
- GET: Now also lists active short codes in a "Quick Enrollment" section above existing token list
- POST with action `generate-link`: Creates a short code (delegates to the generate-link handler logic)

### Existing Enrollment Endpoints (unchanged)

- `POST /api/enroll` — No changes
- `POST /api/enroll/verify` — No changes
- `GET /api/keys` — No changes

---

## Hub Configuration (new CLI flags / env vars)

| Flag | Env Var | Default | Description |
|------|---------|---------|-------------|
| `-external-url` | `VAULT_EXTERNAL_URL` | (required for short codes) | Hub's public URL (e.g., `https://ssh-vault.yurii.live`) |
| `-github-repo` | `VAULT_GITHUB_REPO` | (required for short codes) | GitHub repo path (e.g., `driversti/ssh-vault`) |
| `-release-tag` | `VAULT_RELEASE_TAG` | `latest` | GitHub release tag for binary downloads |
