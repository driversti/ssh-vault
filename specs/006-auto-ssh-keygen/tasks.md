# Tasks: Auto SSH Key Generation During Enrollment

**Input**: Design documents from `/specs/006-auto-ssh-keygen/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md

**Tests**: Test tasks included — key generation logic involves filesystem operations and edge cases that benefit from automated testing.

**Organization**: Tasks are grouped by user story. US1 (enrollment script) and US2 (CLI command) are independent and can be implemented in either order, though US2 depends on the shared `EnsureSSHKey` function created in the Foundational phase.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2)
- Include exact file paths in descriptions

---

## Phase 1: Setup

**Purpose**: No setup needed — project structure and dependencies already exist.

(No tasks in this phase)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Create the shared `EnsureSSHKey` function used by US2 (CLI path). US1 (shell script) is self-contained and does not depend on this phase.

**⚠️ CRITICAL**: US2 cannot begin until this phase is complete.

- [x] T001 Create `EnsureSSHKey(keyPath string, promptFn func(string) bool) (string, error)` function in `internal/agent/keygen.go` — checks if key exists at keyPath, if not: ensures `~/.ssh/` dir exists (mode 700), verifies `ssh-keygen` is available, calls `ssh-keygen -t ed25519 -f <path> -N "" -q`, returns key path
- [x] T002 Create tests for `EnsureSSHKey` in `internal/agent/keygen_test.go` — table-driven tests covering: key already exists (no-op), key missing + prompt accepts (generates), key missing + prompt declines (returns error), ssh-keygen not found (error), directory creation, never overwrites existing keys

**Checkpoint**: `EnsureSSHKey` function passes all tests with `go test ./internal/agent/`

---

## Phase 3: User Story 1 — Automatic Key Generation in Enrollment Script (Priority: P1) 🎯 MVP

**Goal**: Replace the "no key found" error in the enrollment shell script with automatic `ssh-keygen` invocation.

**Independent Test**: Run enrollment script on a machine with no SSH keys — enrollment should complete without manual key generation.

### Implementation

- [x] T003 [US1] Modify `buildEnrollmentScript()` in `internal/hub/handlers_shortcode.go` — replace the error block (lines 284-290) with: `mkdir -p -m 700 ~/.ssh`, check `command -v ssh-keygen`, run `ssh-keygen -t ed25519 -f ~/.ssh/id_ed25519 -N "" -q`, echo feedback message, set `SSH_KEY=~/.ssh/id_ed25519.pub`
- [x] T004 [US1] Verify the existing "key found" path in the enrollment script remains unchanged — existing keys should be used as before (regression check)

**Checkpoint**: Enrollment script auto-generates keys and completes enrollment on a keyless machine.

---

## Phase 4: User Story 2 — Automatic Key Generation in CLI Enroll Command (Priority: P2)

**Goal**: Add interactive prompt to `ssh-vault enroll` CLI command when the specified key file is missing.

**Independent Test**: Run `ssh-vault enroll --key ~/.ssh/nonexistent` — should prompt "Generate SSH key? (y/n)" and proceed on "y".

### Implementation

- [x] T005 [US2] Modify `runEnroll()` in `cmd/ssh-vault/main.go` — after expanding the key path, check if private key file exists; if not, call `EnsureSSHKey()` with a stdin-based prompt function before calling `agent.Enroll()`
- [x] T006 [US2] Verify the "user declines" path — when user answers "n" to the prompt, the command should exit with a clear message suggesting manual key generation

**Checkpoint**: CLI `enroll` command prompts and generates keys when missing, proceeds with enrollment.

---

## Phase 5: Polish & Validation

- [x] T007 Run `go build ./...` and `go vet ./...` to verify no compilation errors
- [x] T008 Run `go test ./...` to verify all tests pass
- [x] T009 Run quickstart.md validation scenarios (both enrollment paths, regression with existing keys)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Foundational (Phase 2)**: No dependencies — can start immediately
- **US1 (Phase 3)**: No dependencies on Phase 2 — shell script is self-contained, can start immediately
- **US2 (Phase 4)**: Depends on Phase 2 (`EnsureSSHKey` function)
- **Polish (Phase 5)**: Depends on all previous phases

### Parallel Opportunities

- **T001 and T003**: Can run in parallel (different files: `keygen.go` vs `handlers_shortcode.go`)
- **T002**: Depends on T001 (tests need the function)
- **T005**: Depends on T001 (uses `EnsureSSHKey`)
- **T003 and T005**: Can run in parallel after T001 (different files)

### Within Each User Story

- US1: T003 → T004 (sequential, same file context)
- US2: T005 → T006 (sequential, same file context)

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Start T001 and T003 in parallel
2. Complete T002 (tests for EnsureSSHKey)
3. Complete T004 (US1 regression check)
4. **STOP and VALIDATE**: Test enrollment script on a keyless machine
5. This alone delivers the primary user value

### Full Feature

1. After MVP validation, proceed to T005-T006 (US2)
2. Complete T007-T009 (Polish)
3. All enrollment paths now handle missing keys gracefully
