# Tasks: Short Enrollment URL

**Input**: Design documents from `/specs/003-short-enrollment-url/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, contracts/

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: No new project setup needed — existing Go project structure is sufficient. This phase adds shared configuration that all user stories depend on.

- [x] T001 Add hub configuration flags (`-external-url`, `-github-repo`, `-release-tag`) and corresponding env var support (`VAULT_EXTERNAL_URL`, `VAULT_GITHUB_REPO`, `VAULT_RELEASE_TAG`) to `cmd/ssh-vault/main.go`, passing them to `hub.Server`
- [x] T002 Add `ExternalURL`, `GithubRepo`, `ReleaseTag` fields to the `Server` struct in `internal/hub/server.go`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core model, storage, and rate limiting infrastructure that MUST be complete before ANY user story can be implemented

**CRITICAL**: No user story work can begin until this phase is complete

- [x] T003 [P] Create `ShortCode` model struct with fields (`Code`, `TokenValue`, `CreatedAt`, `ExpiresAt`, `Used`, `UsedAt`, `UsedByIP`), `NewShortCode()` constructor, `IsValid()`, `IsExpired()`, and `MarkUsed()` methods in `internal/model/shortcode.go`
- [x] T004 [P] Create table-driven tests for `ShortCode` model (constructor, validity, expiration, mark-used) in `internal/model/shortcode_test.go`
- [x] T005 [P] Add new audit event constants `EventShortCodeCreated`, `EventShortCodeUsed`, `EventShortCodeExpired` to `internal/model/audit.go`
- [x] T006 Add `ShortCodes []ShortCode` field to the `Store` struct in `internal/model/store.go`
- [x] T007 Implement `FileStore` methods for ShortCode: `AddShortCode()`, `GetShortCode()`, `UseShortCode()`, `ListShortCodes()`, `PurgeExpiredShortCodes()` following existing mutex and atomic-write patterns in `internal/hub/store.go`
- [x] T008 [P] Create per-IP sliding window rate limiter with `Allow(ip string) bool`, configurable window (1 min) and limit (10 req), lazy cleanup of stale entries in `internal/hub/ratelimit.go`
- [x] T009 [P] Create table-driven tests for rate limiter (allow, deny, window expiry, IP isolation, cleanup) in `internal/hub/ratelimit_test.go`

**Checkpoint**: Foundation ready — ShortCode model, store operations, and rate limiter are available for user story implementation

---

## Phase 3: User Story 2 - Short Code Generation (Priority: P1) 🎯 MVP

**Goal**: Admin can generate 6-digit short enrollment codes from the hub dashboard and see the full `curl | sh` command ready to copy

**Independent Test**: Log into the hub dashboard, click "Generate Enrollment Link", verify a 6-digit code appears with a copyable curl command showing the hub's external URL

### Implementation for User Story 2

- [x] T010 [US2] Create short code generation handler: generate 6-digit code, auto-create linked enrollment token (24h TTL), create ShortCode (15min TTL), add audit entries, redirect to `/tokens` in `internal/hub/handlers_shortcode.go`
- [x] T011 [US2] Add helper function to generate unique 6-digit code (100000–999999 via `crypto/rand`, retry on collision with active codes) in `internal/hub/handlers_shortcode.go`
- [x] T012 [US2] Update `tokens.html` template to add "Quick Enrollment" section above existing token list: "Generate Enrollment Link" button (POST form), active short codes table with code, status, curl command, expiry countdown, and copy button in `internal/hub/templates/tokens.html`
- [x] T013 [US2] Register route `POST /tokens/generate-link` with `requireSession` middleware in `internal/hub/server.go`
- [x] T014 [US2] Update `handleTokens` GET handler to pass active short codes and hub external URL to the template in `internal/hub/handlers.go`
- [x] T015 [US2] Create tests for short code generation handler (successful generation, uniqueness, audit logging, session required) in `internal/hub/handlers_shortcode_test.go`

**Checkpoint**: Admin can generate short codes and see curl commands on the dashboard. The enrollment URL doesn't serve scripts yet.

---

## Phase 4: User Story 1 - One-Command Device Enrollment (Priority: P1) 🎯 MVP

**Goal**: User runs `curl -sSL https://hub/e/CODE | sh` and the device is enrolled automatically (appears as pending in hub)

**Independent Test**: Generate a short code on the dashboard, run the curl command on a target device, verify the device appears as "pending" in the hub's device list

### Implementation for User Story 1

- [x] T016 [US1] Create enrollment script shell template as a Go `text/template` constant with variables for `HubURL`, `Token`, `DownloadBaseURL` — include: `set -euo pipefail`, platform detection (`uname -s`/`uname -m`), binary download, hostname detection, SSH key discovery (`~/.ssh/id_*.pub`), enrollment execution, cleanup trap in `internal/hub/handlers_shortcode.go`
- [x] T017 [US1] Implement `handleShortCodeEnroll` handler for `GET /e/{code}`: validate code via store, apply rate limiter, resolve linked token value, render enrollment script template with hub config values, return as `text/plain` in `internal/hub/handlers_shortcode.go`
- [x] T018 [US1] Implement error responses for `GET /e/{code}` as valid shell scripts (`echo "Error: ..."; exit 1`) for invalid/expired/used codes (404/410) and rate limit exceeded (429) in `internal/hub/handlers_shortcode.go`
- [x] T019 [US1] Register route `GET /e/` (prefix pattern) with rate limiter middleware in `internal/hub/server.go`
- [x] T020 [US1] Create tests for enrollment script handler (valid code returns script, expired code returns 410 shell error, used code returns 404 shell error, rate limit returns 429, script contains correct template variables) in `internal/hub/handlers_shortcode_test.go`

**Checkpoint**: Full enrollment flow works end-to-end: admin generates link → user runs curl | sh → device enrolled as pending

---

## Phase 5: User Story 3 - Enrollment Script Delivery (Priority: P2)

**Goal**: The enrollment script correctly detects Linux (amd64, arm64) and macOS (arm64), downloads the right binary, and shows a clear error on unsupported platforms

**Independent Test**: Request the enrollment URL from devices with different OS/arch combinations and verify the script handles each correctly

### Implementation for User Story 3

- [x] T021 [US3] Enhance enrollment script template with robust platform mapping: `uname -s` → `linux`/`darwin`, `uname -m` → `amd64`/`arm64` (with `x86_64`→`amd64` and `aarch64`→`arm64` aliases), unsupported platform error message listing supported combinations in `internal/hub/handlers_shortcode.go`
- [x] T022 [US3] Add binary checksum verification to the enrollment script template: download `checksums.txt` from GitHub Release, verify binary SHA-256 before executing in `internal/hub/handlers_shortcode.go`
- [x] T023 [US3] Add detection of existing ssh-vault installation in enrollment script: check PATH and common locations, use existing binary if version matches, download if missing or outdated in `internal/hub/handlers_shortcode.go`
- [x] T024 [US3] Add "no SSH key found" error handling to enrollment script: check for `~/.ssh/id_*.pub`, display clear message suggesting `ssh-keygen` if none found in `internal/hub/handlers_shortcode.go`
- [x] T025 [US3] Create tests verifying enrollment script content includes platform detection, checksum verification, SSH key discovery, and error handling sections in `internal/hub/handlers_shortcode_test.go`

**Checkpoint**: All user stories are independently functional. Enrollment works seamlessly across Linux (amd64, arm64) and macOS (arm64).

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Cleanup, expiry management, and validation

- [x] T026 [P] Wire `PurgeExpiredShortCodes()` into the `handleTokens` GET handler (alongside existing `PurgeExpiredTokens()`) to clean up expired short codes and their orphaned tokens on dashboard visit in `internal/hub/handlers.go`
- [x] T027 [P] Add short code expiry cleanup for orphaned tokens: when purging an expired short code, also remove its linked token if the token hasn't been used by another flow in `internal/hub/store.go`
- [x] T028 Run `go vet ./...` and `go test ./...` to verify all tests pass and no vet issues exist
- [x] T029 Run quickstart.md validation: build binary, start hub with new flags, generate enrollment link, verify curl command works

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion — BLOCKS all user stories
- **US2 - Short Code Generation (Phase 3)**: Depends on Foundational — generates codes on dashboard
- **US1 - One-Command Enrollment (Phase 4)**: Depends on Foundational — can run in parallel with US2 (different files), but end-to-end testing requires US2 for code generation
- **US3 - Script Delivery (Phase 5)**: Depends on US1 — enhances the enrollment script created in US1
- **Polish (Phase 6)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 2 (P1)**: Can start after Foundational (Phase 2) — no dependencies on other stories
- **User Story 1 (P1)**: Can start after Foundational (Phase 2) — independent from US2 at code level, but integration testing needs US2
- **User Story 3 (P2)**: Depends on US1 (enhances the enrollment script template created there)

### Within Each User Story

- Models before services
- Services before handlers
- Core implementation before integration
- Story complete before moving to next priority

### Parallel Opportunities

- T003, T004, T005 can run in parallel (different files in `internal/model/`)
- T008, T009 can run in parallel with T003–T007 (different package: `internal/hub/ratelimit.go`)
- US1 and US2 implementation tasks are in different handler functions within the same file, but can be developed in parallel by different agents
- T026, T027 can run in parallel (different files)

---

## Parallel Example: Foundational Phase

```bash
# Launch model tasks in parallel:
Task: "Create ShortCode model in internal/model/shortcode.go"
Task: "Create ShortCode tests in internal/model/shortcode_test.go"
Task: "Add audit event constants in internal/model/audit.go"

# Launch rate limiter in parallel with model tasks:
Task: "Create rate limiter in internal/hub/ratelimit.go"
Task: "Create rate limiter tests in internal/hub/ratelimit_test.go"
```

---

## Implementation Strategy

### MVP First (User Stories 1 + 2)

1. Complete Phase 1: Setup (config flags)
2. Complete Phase 2: Foundational (model, store, rate limiter)
3. Complete Phase 3: User Story 2 (admin generates codes on dashboard)
4. Complete Phase 4: User Story 1 (curl | sh enrollment works)
5. **STOP and VALIDATE**: Generate a code, run curl on a device, verify enrollment
6. Deploy if ready — platform-specific hardening (US3) can follow

### Incremental Delivery

1. Setup + Foundational → Foundation ready
2. Add US2 → Admin can generate enrollment links (dashboard value)
3. Add US1 → Full enrollment flow works (core value delivered!)
4. Add US3 → Multi-platform hardening and checksum verification
5. Polish → Cleanup, expiry management, quickstart validation

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- The enrollment script template (T016) is the most complex single task — it produces a ~80-line shell script via Go text/template
