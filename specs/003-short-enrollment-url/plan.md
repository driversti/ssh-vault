# Implementation Plan: Short Enrollment URL

**Branch**: `003-short-enrollment-url` | **Date**: 2026-03-21 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/003-short-enrollment-url/spec.md`

## Summary

Add a simplified enrollment flow where admins generate a 6-digit short code on the hub dashboard, which produces a `curl -sSL https://hub/e/CODE | sh` command. When executed on a target device, the hub serves a POSIX shell script that auto-detects OS/arch, downloads the ssh-vault binary from GitHub Releases, discovers the local SSH key and hostname, and completes the existing 3-step enrollment handshake automatically. Short codes are single-use with a 15-minute default TTL and per-IP rate limiting on the `/e/{code}` endpoint.

## Technical Context

**Language/Version**: Go 1.26.1
**Primary Dependencies**: `golang.org/x/crypto/ssh` (existing), standard library only
**Storage**: JSON file via `FileStore` (atomic write with temp file + rename, `sync.RWMutex` for concurrency)
**Testing**: Standard `testing` package, table-driven tests, `go test ./...`
**Target Platform**: Linux (amd64, arm64), macOS (arm64) — hub runs on Linux server
**Project Type**: CLI tool + embedded web server (single binary)
**Performance Goals**: Enrollment script served in <1s; rate limiter at 10 req/min/IP on `/e/{code}`
**Constraints**: Standard library only (per constitution); single JSON file storage; no external dependencies
**Scale/Scope**: Small-scale personal infrastructure; ~10 devices; ~1-5 concurrent enrollment attempts max

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Idiomatic Go | PASS | Standard library HTTP handlers, explicit error handling, `gofmt`/`goimports` |
| II. Testing | PASS | Table-driven tests for short code model, handler tests for `/e/{code}` endpoint |
| III. Simplicity | PASS | Extends existing patterns (model + store + handler); no new abstractions; single binary maintained |
| Prefer standard library | PASS | Rate limiter uses `sync.Mutex` + `time.Now()` — no external packages needed |
| One binary, one `main` | PASS | No new binaries; new handler registered on existing `ServeMux` |

No violations. All gates pass.

## Project Structure

### Documentation (this feature)

```text
specs/003-short-enrollment-url/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
└── tasks.md             # Phase 2 output (/speckit.tasks)
```

### Source Code (repository root)

```text
internal/
├── model/
│   ├── shortcode.go         # NEW: ShortCode struct, NewShortCode, IsValid, IsExpired
│   ├── shortcode_test.go    # NEW: Table-driven tests for ShortCode
│   ├── audit.go             # MODIFIED: New event constants (shortcode_created, shortcode_used)
│   └── store.go             # MODIFIED: Add ShortCodes []ShortCode to Store struct
├── hub/
│   ├── handlers.go          # MODIFIED: Add handleShortCodeEnroll (GET /e/{code})
│   ├── handlers_shortcode.go # NEW: Short code generation handler, enrollment script template
│   ├── handlers_shortcode_test.go # NEW: Tests for short code endpoints
│   ├── ratelimit.go         # NEW: Per-IP rate limiter (in-memory, token bucket or sliding window)
│   ├── ratelimit_test.go    # NEW: Rate limiter tests
│   ├── server.go            # MODIFIED: Register new routes
│   ├── store.go             # MODIFIED: AddShortCode, GetShortCode, UseShortCode, ListShortCodes, PurgeExpiredShortCodes
│   └── templates/
│       └── tokens.html      # MODIFIED: Add "Generate Enrollment Link" section
cmd/ssh-vault/
└── main.go                  # MODIFIED: Add -external-url, -github-repo, -release-tag flags
```

**Structure Decision**: Follows the existing single-project Go structure. New files are placed in existing packages following established patterns: model types in `internal/model/`, handlers in `internal/hub/`, tests alongside source files.

## Complexity Tracking

No constitution violations — this section is intentionally empty.
