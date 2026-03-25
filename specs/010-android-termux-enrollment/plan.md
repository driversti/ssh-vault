# Implementation Plan: Android Termux Enrollment

**Branch**: `010-android-termux-enrollment` | **Date**: 2026-03-25 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/010-android-termux-enrollment/spec.md`

## Summary

Enable Android device enrollment via Termux by: (1) adding `linux/arm64` to the Docker build pipeline, (2) adapting the enrollment script to detect Termux and use appropriate paths (`$PREFIX/bin`, no sudo), (3) adding a `ssh-vault sync` one-shot subcommand for cron-based key syncing, and (4) setting up a 15-minute cron job via `crond` on Termux after enrollment.

## Technical Context

**Language/Version**: Go 1.26.1
**Primary Dependencies**: `golang.org/x/crypto/ssh` (existing), standard library only
**Storage**: JSON file via `FileStore` (unchanged)
**Testing**: `go test ./...` with table-driven tests
**Target Platform**: Linux ARM64 (Termux on Android), plus existing linux/amd64, darwin/arm64
**Project Type**: CLI (single binary: hub server + agent + enrollment)
**Performance Goals**: One-shot sync completes in <5s; cron job adds negligible battery impact at 15-minute intervals
**Constraints**: No CGO (static binary), no sudo on Termux, no new dependencies
**Scale/Scope**: Single device enrollment; ~4 files modified, ~1 new file

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Idiomatic Go | PASS | All changes follow standard Go conventions. New `sync` subcommand uses same `flag.NewFlagSet` pattern as existing commands. |
| II. Testing | PASS | New `SyncOnce()` export tested via existing `syncOnce()` test coverage. New `sync` CLI command tested with table-driven tests. Termux detection logic tested with mock environment variables. |
| III. Simplicity | PASS | No new abstractions. Reuses existing `syncOnce()` function. Enrollment script changes are conditional branches in existing code. One binary, one `main` package maintained. |
| Technology Stack | PASS | Standard library only. No new dependencies. |
| Development Workflow | PASS | Feature branch + PR. |

**Post-Phase 1 re-check**: PASS — No new abstractions, patterns, or dependencies introduced. All changes are minimal additions to existing code.

## Project Structure

### Documentation (this feature)

```text
specs/010-android-termux-enrollment/
├── plan.md              # This file
├── spec.md              # Feature specification
├── research.md          # Phase 0 research findings
├── data-model.md        # Data model (no changes needed)
├── quickstart.md        # Termux enrollment guide
├── contracts/           # CLI contract for sync command
│   └── cli-sync-command.md
├── checklists/
│   └── requirements.md  # Spec quality checklist
└── tasks.md             # Phase 2 output (created by /speckit.tasks)
```

### Source Code (repository root)

```text
cmd/ssh-vault/
└── main.go                          # MODIFY: Add "sync" subcommand routing + runSync()

internal/agent/
├── agent.go                         # MODIFY: Export syncOnce() → SyncOnce()
└── agent_test.go                    # MODIFY: Add SyncOnce() tests

internal/hub/
└── handlers_shortcode.go            # MODIFY: Termux detection + path adaptation + cron setup in buildEnrollmentScript()

Dockerfile                           # MODIFY: Add build-linux-arm64 stage + copy to dist
```

**Structure Decision**: Existing single-project structure maintained. All changes are modifications to existing files. No new packages or directories in source code.

## Implementation Steps

### Step 1: Add linux/arm64 Build to Dockerfile

**Files**: `Dockerfile`

Add a `build-linux-arm64` stage and copy its output to the dist assembly stage:
- New stage: `FROM base AS build-linux-arm64` with `CGO_ENABLED=0 GOOS=linux GOARCH=arm64`
- Update dist stage: `COPY --from=build-linux-arm64 /out/ /dist/`
- Checksum generation already covers all `ssh-vault_*` files via wildcard

**Verification**: Build Docker image, verify `ssh-vault_linux_arm64` exists in dist/ with valid checksum.

### Step 2: Export SyncOnce and Add Sync Subcommand

**Files**: `internal/agent/agent.go`, `cmd/ssh-vault/main.go`

a) In `agent.go`: Rename `syncOnce()` to `SyncOnce()` (export). Update all internal callers in `Run()`.

b) In `main.go`: Add `sync` case to the command switch:
- New `runSync()` function with `--config` flag (default: `~/.ssh-vault/agent.json`)
- Loads config via `agent.LoadConfig()`
- Calls `agent.SyncOnce(cfg)`
- Returns appropriate exit codes (0 success, 1 config error, 2 hub unreachable, 3 revoked)

c) Update `printUsage()` to list the `sync` command.

**Verification**: `go build && ./ssh-vault sync` performs one sync and exits. `go test ./...` passes.

### Step 3: Adapt Enrollment Script for Termux

**Files**: `internal/hub/handlers_shortcode.go`

In `buildEnrollmentScript()`, add Termux detection and conditional logic:

a) **Detection** (after OS/arch detection):
```
IS_TERMUX=false
if [ -n "${TERMUX_VERSION:-}" ]; then
    IS_TERMUX=true
fi
```

b) **Binary installation** (replace existing install section):
- If `IS_TERMUX=true`: install to `$PREFIX/bin/ssh-vault` (no sudo)
- Else: keep existing `/usr/local/bin` → `~/.local/bin` fallback

c) **Tool checks** (add Termux-specific messages):
- If Termux and missing tools: suggest `pkg install openssh curl termux-services`

d) **Post-enrollment agent startup** (replace `nohup` section):
- If `IS_TERMUX=true`: Set up cron job instead of `nohup` daemon
  - Check `crond` availability, suggest `pkg install termux-services && sv-enable crond` if missing
  - Install crontab: `*/15 * * * * $PREFIX/bin/ssh-vault sync >> ~/.ssh-vault/sync.log 2>&1`
  - Run one immediate sync: `ssh-vault sync`
- Else: keep existing `nohup` daemon launch

**Verification**: Generate enrollment script, inspect output for Termux conditionals. Test in Termux if available.

### Step 4: Tests

**Files**: `internal/agent/agent_test.go`, `cmd/ssh-vault/main.go` (or `main_test.go`)

a) Test `SyncOnce()` export works correctly (may already be covered by existing tests on `syncOnce`).

b) Test `sync` CLI subcommand:
- Missing config file → exit code 1
- Successful sync → exit code 0
- Hub unreachable → exit code 2

c) Test enrollment script Termux detection:
- Verify Termux-detected script contains `$PREFIX/bin` install path
- Verify Termux-detected script contains crontab setup
- Verify non-Termux script is unchanged

## Complexity Tracking

No constitution violations. No complexity tracking entries needed.
