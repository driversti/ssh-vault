# Tasks: Android Termux Enrollment

**Input**: Design documents from `/specs/010-android-termux-enrollment/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, contracts/

**Tests**: Included — the constitution requires tests for non-trivial code.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: No setup needed — existing project structure is sufficient. All changes modify existing files.

*(No tasks in this phase)*

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Export the existing sync function so it can be called by the new `sync` subcommand and by cron jobs. This is cross-cutting and blocks US4.

**CRITICAL**: Must complete before User Story 4 can begin.

- [x] T001 Export `syncOnce()` as `SyncOnce()` in `internal/agent/agent.go` and update all internal callers in `Run()` to use the new name
- [x] T002 Add test for `SyncOnce()` in `internal/agent/agent_test.go` — verify it calls fetchKeys and writes authorized_keys (table-driven: success case, hub unreachable, device revoked)

**Checkpoint**: `go test ./internal/agent/...` passes with exported `SyncOnce()`

---

## Phase 3: User Story 2 - Build System Produces Android-Compatible Binary (Priority: P1) MVP

**Goal**: The Docker build pipeline produces a `linux/arm64` binary that can run in Termux.

**Independent Test**: Build the Docker image and verify `ssh-vault_linux_arm64` exists in the dist output with a valid checksum entry.

### Implementation for User Story 2

- [x] T003 [US2] Add `build-linux-arm64` stage to `Dockerfile` — `FROM base AS build-linux-arm64` with `CGO_ENABLED=0 GOOS=linux GOARCH=arm64`, output to `/out/ssh-vault_linux_arm64`
- [x] T004 [US2] Update dist assembly stage in `Dockerfile` — add `COPY --from=build-linux-arm64 /out/ /dist/` alongside existing linux-amd64 and darwin-arm64 copies

**Checkpoint**: `docker build` produces dist/ containing `ssh-vault_linux_arm64` with SHA-256 in checksums.txt. Download handler already supports `linux/arm64` validation — no code changes needed.

---

## Phase 4: User Story 3 - Enrollment Script Adapts to Termux Environment (Priority: P1)

**Goal**: The enrollment shell script detects Termux and uses correct paths (`$PREFIX/bin`, no sudo) and provides Termux-specific error messages.

**Independent Test**: Generate an enrollment script from the hub and verify the output contains Termux detection logic, `$PREFIX/bin` install path, and `pkg install` guidance for missing tools.

### Implementation for User Story 3

- [x] T005 [US3] Add Termux environment detection in `buildEnrollmentScript()` in `internal/hub/handlers_shortcode.go` — add `IS_TERMUX` variable check via `$TERMUX_VERSION` env var after the existing OS/arch detection block
- [x] T006 [US3] Add Termux-specific binary installation path in `buildEnrollmentScript()` in `internal/hub/handlers_shortcode.go` — when `IS_TERMUX=true`, install to `$PREFIX/bin/ssh-vault` without sudo; keep existing `/usr/local/bin` → `~/.local/bin` fallback for non-Termux
- [x] T007 [US3] Add Termux-specific tool-check error messages in `buildEnrollmentScript()` in `internal/hub/handlers_shortcode.go` — when Termux detected and tools missing, suggest `pkg install openssh curl` instead of generic messages
- [x] T008 [US3] Add test for Termux enrollment script generation in `internal/hub/handlers_shortcode_test.go` — verify generated script contains `TERMUX_VERSION` check, `$PREFIX/bin` install path, and `pkg install` instructions; also verify non-Termux script is unchanged

**Checkpoint**: User Story 1 (Admin Enrolls Android Device) is now functional end-to-end — the binary exists (US2) and the enrollment script adapts to Termux (US3). An admin can generate an enrollment link, run it in Termux, and the device enrolls successfully. The agent starts via `nohup` on Termux as a temporary measure until US4 adds cron.

---

## Phase 5: User Story 4 - Agent Syncs Persistently on Android via Cron (Priority: P2)

**Goal**: Add a `ssh-vault sync` one-shot subcommand (all platforms) and set up a 15-minute cron job via `crond` on Termux after enrollment.

**Independent Test**: Run `ssh-vault sync` manually and verify it performs one sync and exits. On Termux, verify `crontab -l` shows the sync job after enrollment.

### Implementation for User Story 4

- [x] T009 [US4] Add `sync` subcommand routing in `cmd/ssh-vault/main.go` — add `case "sync"` to the command switch, implement `runSync()` with `--config` flag (default `~/.ssh-vault/agent.json`), load config and call `agent.SyncOnce()`, exit with appropriate codes (0=success, 1=config error, 2=hub unreachable, 3=revoked)
- [x] T010 [US4] Update `printUsage()` in `cmd/ssh-vault/main.go` to list the `sync` command alongside existing `hub`, `agent`, and `enroll` commands
- [x] T011 [US4] Add cron job setup in `buildEnrollmentScript()` in `internal/hub/handlers_shortcode.go` — when `IS_TERMUX=true`, replace the `nohup` daemon launch section with: check `crond` availability (suggest `pkg install termux-services && sv-enable crond` if missing), install crontab entry `*/15 * * * * $PREFIX/bin/ssh-vault sync >> ~/.ssh-vault/sync.log 2>&1`, run one immediate `ssh-vault sync`
- [x] T012 [P] [US4] Add test for `sync` subcommand in `internal/agent/agent_test.go` — verify SyncOnce returns correct errors for missing config, hub unreachable, and revoked device scenarios
- [x] T013 [P] [US4] Add test for Termux cron setup in enrollment script in `internal/hub/handlers_shortcode_test.go` — verify Termux-detected script contains crontab entry with `*/15` interval and `ssh-vault sync`, verify non-Termux script still uses `nohup` daemon

**Checkpoint**: Full feature complete. Android devices can enroll via Termux, sync keys on a 15-minute cron schedule, and persist across Termux restarts.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Final validation and cleanup

- [x] T014 Run `go vet ./...`, `golangci-lint run`, and `go test ./...` to verify all changes pass static analysis, linting, and tests
- [x] T015 Run `go build -o ssh-vault ./cmd/ssh-vault` and verify the binary includes the new `sync` subcommand (`./ssh-vault sync --help` or `./ssh-vault` shows sync in usage)
- [x] T016 Validate quickstart.md instructions against actual implementation — ensure documented commands and paths match the code

---

## Dependencies & Execution Order

### Phase Dependencies

- **Foundational (Phase 2)**: No dependencies — export SyncOnce first
- **US2 Build (Phase 3)**: No dependencies on Phase 2 — can run in parallel with Foundational
- **US3 Script Adaptation (Phase 4)**: No strict code dependency on US2, but logically follows (binary must exist for enrollment to work end-to-end)
- **US4 Cron Sync (Phase 5)**: Depends on Phase 2 (SyncOnce exported) and Phase 4 (Termux detection in script)
- **Polish (Phase 6)**: Depends on all previous phases

### User Story Dependencies

- **User Story 2 (P1)**: Can start immediately — Dockerfile only, no code dependencies
- **User Story 3 (P1)**: Can start immediately — script changes are independent of Dockerfile
- **User Story 1 (P1)**: Integration story — verified when US2 + US3 are both complete
- **User Story 4 (P2)**: Depends on Foundational (SyncOnce export) and US3 (Termux detection in script)

### Within Each User Story

- Implementation tasks within a story are sequential (same file: `handlers_shortcode.go`)
- Tests can be written after implementation (per story checkpoint)
- Story complete before moving to next priority

### Parallel Opportunities

- **Phase 2 + Phase 3**: T001/T002 (Foundational) and T003/T004 (Dockerfile) can run in parallel — different files
- **Phase 5**: T012 and T013 are marked [P] — different test files
- **Cross-phase**: US2 (Dockerfile) and US3 (script adaptation) can proceed in parallel

---

## Parallel Example: Foundational + US2

```
# These can run in parallel (different files):
Task T001: Export SyncOnce() in internal/agent/agent.go
Task T003: Add build-linux-arm64 stage to Dockerfile
```

## Parallel Example: US4 Tests

```
# These can run in parallel (different test files):
Task T012: Test sync subcommand in internal/agent/agent_test.go
Task T013: Test cron setup in internal/hub/handlers_shortcode_test.go
```

---

## Implementation Strategy

### MVP First (User Stories 2 + 3 = US1 Verified)

1. Complete Phase 2: Foundational (export SyncOnce)
2. Complete Phase 3: US2 (Dockerfile — produces linux/arm64 binary)
3. Complete Phase 4: US3 (enrollment script Termux adaptation)
4. **STOP and VALIDATE**: Test enrollment end-to-end in Termux (verifies US1)
5. Deploy/demo if ready — Android enrollment works with manual agent restart

### Incremental Delivery

1. Foundational + US2 + US3 → Android enrollment works (MVP!)
2. Add US4 → Cron-based persistent sync on Android
3. Polish → Full validation and cleanup

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- US1 (Admin Enrolls) is an integration story verified by completing US2 + US3
- US4 modifies the same file as US3 (`handlers_shortcode.go`) — must be sequential
- All changes are modifications to existing files — no new packages or directories
- Constitution requires tests: table-driven tests using standard `testing` package
