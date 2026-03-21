# Implementation Plan: Self-Serve Binary Distribution

**Branch**: `004-self-serve-binaries` | **Date**: 2026-03-21 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/004-self-serve-binaries/spec.md`

## Summary

Replace the hub's dependency on GitHub Releases for agent binary downloads with a local `--dist-dir` directory and a new `GET /download/{os}/{arch}` endpoint. The enrollment script is updated to download from the hub itself. All `--github-repo` and `--release-tag` configuration is removed.

## Technical Context

**Language/Version**: Go 1.26.1
**Primary Dependencies**: Standard library only (existing `golang.org/x/crypto/ssh` unchanged)
**Storage**: Filesystem (dist directory for binaries; existing FileStore for data)
**Testing**: Standard `testing` package, table-driven tests, `go test ./...`
**Target Platform**: Linux/macOS server (hub), Linux/macOS/ARM agents
**Project Type**: CLI tool (single binary: `hub`, `agent`, `enroll` subcommands)
**Performance Goals**: N/A — binary serving is straightforward file I/O
**Constraints**: Standard library only; no third-party dependencies
**Scale/Scope**: Single hub operator, handful of devices

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Idiomatic Go | ✅ Pass | Uses `http.ServeFile` or `io.Copy` from standard library; explicit error handling throughout |
| II. Testing | ✅ Pass | New handler gets table-driven tests; existing shortcode tests updated to remove GitHub assertions |
| III. Simplicity | ✅ Pass | Replaces two config fields with one; no new abstractions; allowlist is a simple map |
| Standard library preference | ✅ Pass | No new dependencies |
| One binary, one main | ✅ Pass | No change to binary structure |

No violations. Complexity Tracking table not needed.

## Project Structure

### Documentation (this feature)

```text
specs/004-self-serve-binaries/
├── plan.md              # This file
├── spec.md              # Feature specification
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/
│   └── api.md           # Phase 1 output
└── tasks.md             # Phase 2 output (created by /speckit.tasks)
```

### Source Code (repository root)

```text
cmd/ssh-vault/main.go         # Modify: replace --github-repo/--release-tag with --dist-dir
internal/hub/
├── server.go                  # Modify: replace githubRepo/releaseTag with distDir in Server struct + config
├── handlers_shortcode.go      # Modify: update enrollment script template; add download handler
├── handlers_shortcode_test.go # Modify: update tests for new download URL pattern
├── handlers_download.go       # NEW: GET /download/{os}/{arch} handler
└── handlers_download_test.go  # NEW: tests for download handler
docker-compose.yml             # Modify: replace VAULT_GITHUB_REPO/VAULT_RELEASE_TAG with VAULT_DIST_DIR + volume
Dockerfile                     # Modify: add /dist directory
.env.example                   # Modify: replace GitHub env vars with VAULT_DIST_DIR
README.md                      # Modify: update CLI reference and quick start examples
```

**Structure Decision**: All changes fit within the existing flat `internal/hub/` package. The download handler gets its own file (`handlers_download.go`) following the existing convention where `handlers_shortcode.go` is separated from `handlers.go`. No new packages needed.
