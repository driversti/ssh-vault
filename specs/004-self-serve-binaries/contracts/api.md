# API Contracts: Self-Serve Binary Distribution

**Feature**: 004-self-serve-binaries
**Date**: 2026-03-21

## New Endpoints

### GET /download/{os}/{arch} — Serve Pre-Built Binary

**Auth**: None (public)
**Rate limit**: None (binaries are open-source; the endpoint is just file serving)

**Path Parameters**:

| Parameter | Allowed Values | Example |
|-----------|---------------|---------|
| `{os}`    | `linux`, `darwin` | `linux` |
| `{arch}`  | `amd64`, `arm64` | `arm64` |

**Success Response** (200 OK):
- Content-Type: `application/octet-stream`
- Content-Disposition: `attachment; filename="ssh-vault_{os}_{arch}"`
- Content-Length: file size in bytes
- Body: raw binary file

**Error Responses**:

| Status | Condition | Body |
|--------|-----------|------|
| 400 | Invalid OS or architecture value | `unsupported platform: {os}/{arch}. Supported: linux/amd64, linux/arm64, darwin/amd64, darwin/arm64` |
| 404 | Binary file not found in dist directory | `binary not available for {os}/{arch}` |
| 501 | `--dist-dir` not configured | `binary distribution not configured` |

**Example**:
```bash
curl -fsSL -o ssh-vault https://hub.example.com/download/linux/amd64
chmod +x ssh-vault
```

---

### GET /download/checksums.txt — Serve Checksums File

**Auth**: None (public)

**Success Response** (200 OK):
- Content-Type: `text/plain; charset=utf-8`
- Body: checksums file content

**Error Responses**:

| Status | Condition | Body |
|--------|-----------|------|
| 404 | `checksums.txt` not found in dist directory | `checksums not available` |
| 501 | `--dist-dir` not configured | `binary distribution not configured` |

---

## Modified Endpoints

### GET /e/{code} — Serve Enrollment Script

**Changes**:
- The `downloadBaseURL` variable in the generated script now points to `{external-url}/download` instead of `https://github.com/{repo}/releases/download/{tag}`.
- The binary download URL in the script changes from `${VAULT_DOWNLOAD_BASE}/ssh-vault_{os}_{arch}` to `${VAULT_DOWNLOAD_BASE}/${OS}/${ARCH}`.
- The checksum download URL changes from `${VAULT_DOWNLOAD_BASE}/checksums.txt` to `${VAULT_DOWNLOAD_BASE}/checksums.txt` (same relative path, different base).

**Before** (current):
```
VAULT_DOWNLOAD_BASE="https://github.com/driversti/ssh-vault/releases/download/v1.0.0"
BINARY_URL="${VAULT_DOWNLOAD_BASE}/ssh-vault_${OS}_${ARCH}"
CHECKSUMS_URL="${VAULT_DOWNLOAD_BASE}/checksums.txt"
```

**After** (new):
```
VAULT_DOWNLOAD_BASE="https://hub.example.com/download"
BINARY_URL="${VAULT_DOWNLOAD_BASE}/${OS}/${ARCH}"
CHECKSUMS_URL="${VAULT_DOWNLOAD_BASE}/checksums.txt"
```

---

### POST /tokens/generate-link — Generate Short Enrollment Code

**Changes**:
- The configuration guard now checks for `externalURL` only (instead of `externalURL && githubRepo`). The `--dist-dir` absence is handled gracefully by the download endpoint, not at link generation time.

---

## Removed Configuration

| Old Flag | Old Env Var | Replacement |
|----------|-------------|-------------|
| `--github-repo` | `VAULT_GITHUB_REPO` | Removed (no replacement needed) |
| `--release-tag` | `VAULT_RELEASE_TAG` | Removed (no replacement needed) |

## New Configuration

| Flag | Env Var | Default | Description |
|------|---------|---------|-------------|
| `--dist-dir` | `VAULT_DIST_DIR` | (empty — not configured) | Path to directory containing pre-built `ssh-vault_{os}_{arch}` binaries |

---

## Existing Endpoints (unchanged)

- `POST /api/enroll` — No changes
- `POST /api/enroll/verify` — No changes
- `GET /api/keys` — No changes
- `GET /healthz` — No changes
