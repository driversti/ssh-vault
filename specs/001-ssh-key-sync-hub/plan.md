# Implementation Plan: SSH Key Sync Hub

**Branch**: `001-ssh-key-sync-hub` | **Date**: 2026-03-21 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/001-ssh-key-sync-hub/spec.md`

## Summary

Build a centralized SSH key distribution system consisting of a hub (web
dashboard + API) and an agent (background sync daemon), packaged as a
single Go binary with subcommands. The hub maintains a device registry and
serves the list of approved public keys. Agents poll the hub on a schedule,
download the approved key set, and inject them into a managed block within
each device's `~/.ssh/authorized_keys` file. Enrollment uses single-use
onboarding tokens; revocation propagates within one sync interval.

## Technical Context

**Language/Version**: Go (latest stable, currently 1.24)
**Primary Dependencies**: Standard library (`net/http`, `html/template`,
`encoding/json`, `crypto/rand`, `os`, `log/slog`) + `golang.org/x/crypto/ssh`
for SSH key parsing
**Storage**: File-based JSON (single `data.json` file for device registry
and tokens — sufficient for <50 personal devices)
**Testing**: `go test` with table-driven tests (per constitution)
**Target Platform**: Linux and macOS (hub: always-on server; agent: any
SSH-capable device)
**Project Type**: CLI tool with embedded web server (single binary,
subcommands: `hub`, `agent`, `enroll`)
**Performance Goals**: Sync cycle completes in <1s for <50 devices; dashboard
loads in <500ms
**Constraints**: Single-user system; <50 devices; hub must be reachable by
all agents over network
**Scale/Scope**: Personal use, 1 owner, <50 devices

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### I. Idiomatic Go — ✅ PASS

- `gofmt`/`goimports` enforced via CI and editor
- All errors handled explicitly — no `_` discards
- Standard library used for HTTP, JSON, templates, crypto
- Only external dependency: `golang.org/x/crypto/ssh` (quasi-stdlib)
- Minimal exported API surface — internal packages where appropriate

### II. Testing — ✅ PASS

- Standard `testing` package throughout
- Table-driven tests for key parsing, block manipulation, token validation
- `go test ./...` runnable from project root
- Test files alongside source (`*_test.go`)

### III. Simplicity — ✅ PASS

- Single binary with subcommands (`hub`, `agent`, `enroll`)
- One `main` package, one `cmd/` entry point
- Flat package structure: `internal/hub`, `internal/agent`,
  `internal/keyblock`, `internal/model`
- File-based JSON storage (no database)
- Server-rendered HTML dashboard (no frontend build step)
- No abstractions until duplication demands them

## Project Structure

### Documentation (this feature)

```text
specs/001-ssh-key-sync-hub/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
└── tasks.md             # Phase 2 output (/speckit.tasks)
```

### Source Code (repository root)

```text
cmd/
└── ssh-vault/
    └── main.go              # Entry point, subcommand routing

internal/
├── hub/
│   ├── server.go            # HTTP server, routing, middleware
│   ├── handlers.go          # Dashboard + API request handlers
│   ├── auth.go              # Password session auth for dashboard
│   ├── store.go             # File-based JSON storage (devices, tokens)
│   └── templates/           # Embedded HTML templates
│       ├── layout.html
│       ├── login.html
│       ├── devices.html
│       └── tokens.html
├── agent/
│   ├── agent.go             # Sync loop, hub client, enrollment
│   ├── config.go            # Agent configuration (hub URL, interval, key path)
│   └── enroll.go            # Enrollment flow (token exchange)
├── keyblock/
│   ├── keyblock.go          # Managed block read/write/replace in authorized_keys
│   └── atomic.go            # Atomic file write helper
└── model/
    ├── device.go            # Device struct, status enum
    └── token.go             # Onboarding token struct, validation

go.mod
go.sum
```

**Structure Decision**: Single-project layout with `cmd/` + `internal/`
following standard Go project conventions. Four internal packages reflect
clear domain boundaries: `hub` (server-side), `agent` (client-side),
`keyblock` (file manipulation), `model` (shared types). The `hub/templates/`
directory is embedded into the binary via `//go:embed`.

## Complexity Tracking

> No constitution violations to justify.
