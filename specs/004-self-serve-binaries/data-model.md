# Data Model: Self-Serve Binary Distribution

**Feature**: 004-self-serve-binaries
**Date**: 2026-03-21

## Entity Changes

This feature has no new data model entities. It modifies the hub's runtime configuration only.

### Modified: Server struct (`internal/hub/server.go`)

**Fields removed**:
- `githubRepo string` — GitHub repository path (e.g., `driversti/ssh-vault`)
- `releaseTag string` — GitHub release tag (e.g., `v1.0.0`)

**Fields added**:
- `distDir string` — Filesystem path to the directory containing pre-built binaries

### Modified: ServerConfig struct (`internal/hub/server.go`)

**Fields removed**:
- `GithubRepo string`
- `ReleaseTag string`

**Fields added**:
- `DistDir string`

## Dist Directory Layout (Filesystem Convention)

The dist directory is not a data model entity but a filesystem convention:

```text
/path/to/dist/
├── ssh-vault_linux_amd64     # Binary for Linux x86_64
├── ssh-vault_linux_arm64     # Binary for Linux ARM64
├── ssh-vault_darwin_amd64    # Binary for macOS x86_64
├── ssh-vault_darwin_arm64    # Binary for macOS Apple Silicon
└── checksums.txt             # Optional: SHA256 checksums
```

**Naming convention**: `ssh-vault_{os}_{arch}` (no file extension)

**Supported values**:

| Parameter | Allowed Values |
|-----------|---------------|
| `{os}`    | `linux`, `darwin` |
| `{arch}`  | `amd64`, `arm64` |

**checksums.txt format** (matches existing GitHub Releases convention):
```text
e3b0c44298fc1c149afb... ssh-vault_linux_amd64
a591a6d40bf420404a011... ssh-vault_linux_arm64
```
