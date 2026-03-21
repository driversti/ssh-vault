# Tasks: SSH Key Sync Hub

**Input**: Design documents from `/specs/001-ssh-key-sync-hub/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, contracts/

**Tests**: Required by constitution (Principle II). Each phase includes test tasks alongside implementation.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **Single project**: `cmd/ssh-vault/` for entry point, `internal/` for packages
- Go conventions: `cmd/` + `internal/` at repository root

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization, Go module setup, and directory structure

- [x] T001 Create project directory structure: `cmd/ssh-vault/`, `internal/hub/`, `internal/hub/templates/`, `internal/agent/`, `internal/keyblock/`, `internal/model/`
- [x] T002 Initialize Go module with `go mod init` and add `golang.org/x/crypto` dependency in `go.mod`
- [x] T003 Create subcommand dispatch entry point with `flag.FlagSet` for `hub`, `agent`, and `enroll` subcommands in `cmd/ssh-vault/main.go`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core types, storage, and file manipulation that ALL user stories depend on

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T004 [P] Create Device struct with Status enum (pending/approved/revoked), JSON tags, state transition methods (Approve, Revoke), and validation in `internal/model/device.go`
- [x] T005 [P] Create Token struct with validation methods (IsValid, IsExpired), generation helper (NewToken with 32-byte hex value, configurable expiry), and JSON tags in `internal/model/token.go`
- [x] T006 [P] Create AuditEntry struct with Event enum (enrolled/approved/revoked/auth_failed), constructor helpers, and JSON tags in `internal/model/audit.go`
- [x] T007 [P] Create Store top-level struct (Devices, Tokens, AuditLog slices) with JSON schema matching `data-model.md` in `internal/model/store.go`
- [x] T008 [P] Implement atomic file write helper: `WriteFileAtomic(path string, data []byte, perm os.FileMode) error` using `os.CreateTemp` + `os.Chmod` + `os.Rename` pattern, with symlink resolution via `filepath.EvalSymlinks` in `internal/keyblock/atomic.go`
- [x] T009 [P] Implement managed key block operations: `ReadBlock(filePath) ([]string, error)`, `WriteBlock(filePath string, keys []string) error` — parse authorized_keys to find `BEGIN/END SSH-VAULT MANAGED BLOCK` markers, replace content between them (or append if absent), preserve all content outside the block, sort keys by fingerprint, use `WriteFileAtomic` for writing, handle missing file creation with 0600 permissions in `internal/keyblock/keyblock.go`
- [x] T010 Implement file-based JSON storage engine: `FileStore` struct wrapping `sync.RWMutex`, `Load(path) error`, `Save() error` (atomic write), CRUD methods for devices (`AddDevice`, `GetDevice`, `UpdateDevice`, `ListDevices`, `ListDevicesByStatus`), CRUD for tokens (`AddToken`, `GetToken`, `UseToken`, `ListTokens`), `AddAuditEntry`, and `ListAuditLog` in `internal/hub/store.go`

- [x] T010a [P] Write table-driven tests for Device: status transitions (Approve valid/invalid, Revoke valid/invalid), validation (empty name, invalid status), JSON round-trip in `internal/model/device_test.go`
- [x] T010b [P] Write table-driven tests for Token: IsValid/IsExpired with edge cases (just expired, just created, already used), NewToken generation, JSON round-trip in `internal/model/token_test.go`
- [x] T010c [P] Write table-driven tests for managed key block: ReadBlock/WriteBlock with empty file, existing content outside block, block replacement, missing markers (append), sorted output, atomic write failure in `internal/keyblock/keyblock_test.go`
- [x] T010d Write table-driven tests for FileStore: AddDevice/GetDevice/ListDevicesByStatus, AddToken/UseToken/GetToken, concurrent access (mutex), Load/Save round-trip with temp file in `internal/hub/store_test.go`

**Checkpoint**: Foundation ready — all types, storage, and file manipulation are implemented and tested. User story implementation can begin.

---

## Phase 3: User Story 1 — Enroll a New Device (Priority: P1) 🎯 MVP

**Goal**: Enable a device owner to enroll a new device using an onboarding token, approve it via the hub, and have its SSH public key distributed to all other enrolled devices within one sync cycle.

**Independent Test**: Generate a token (via store directly or a minimal CLI), run `ssh-vault enroll`, approve the device in the store, run `ssh-vault agent` once, verify the enrolled device's key appears in another device's authorized_keys managed block.

### Implementation for User Story 1

- [x] T011 [US1] Implement `POST /api/enroll` handler: accept JSON body `{token, public_key, name}`, validate token (not used, not expired), parse and validate SSH public key with `ssh.ParseAuthorizedKey`, create device record with status "pending", generate random challenge (32 bytes hex), store challenge on device record, mark token as used, add audit entry, return `{device_id, challenge}` in `internal/hub/handlers.go`
- [x] T012 [US1] Implement `POST /api/enroll/verify` handler: accept JSON body `{device_id, signature}`, look up pending device, verify SSH signature of stored challenge using `ssh.PublicKey.Verify`, on success set `Verified = true` and clear `Challenge` field (status remains "pending"), add audit entry, return `{status: "pending", message}` in `internal/hub/handlers.go`
- [x] T013 [US1] Implement `POST /devices/{id}/approve` handler: look up device by ID, verify status is "pending" AND `Verified == true` (reject unverified devices with 400), transition to "approved" via `device.Approve()`, generate API token (32-byte hex), store on device record, add audit entry, redirect to `/` in `internal/hub/handlers.go`
- [x] T014 [US1] Implement `GET /api/keys` handler: extract bearer token from `Authorization` header, look up device by API token, reject if not found or revoked (401), update `LastSyncAt` timestamp, return JSON `{keys: [...], updated_at}` with all approved devices' public keys excluding the requesting device's own key, sorted by fingerprint in `internal/hub/handlers.go`
- [x] T015 [US1] Implement bearer token authentication middleware: extract `Authorization: Bearer <token>` header, look up device by API token in store, reject revoked/unknown tokens with 401, attach device to request context in `internal/hub/server.go`
- [x] T016 [US1] Implement hub HTTP server: create `http.ServeMux`, register API routes (`/api/enroll`, `/api/enroll/verify`, `/api/keys`), register dashboard route (`/devices/{id}/approve`), configure `log/slog` structured logging, start listener on configurable address in `internal/hub/server.go`
- [x] T017 [US1] Implement enrollment client flow in agent: load SSH private key with `ssh.ParsePrivateKey`, read public key, `POST /api/enroll` with token + public key + hostname, receive challenge, sign challenge with `ssh.Signer.Sign`, `POST /api/enroll/verify` with device_id + signature, display result, save API token + device_id to local config file (`~/.ssh-vault/agent.json`) in `internal/agent/enroll.go`
- [x] T018 [US1] Implement agent configuration: define Config struct (HubURL, Interval, KeyPath, AuthKeysPath, APIToken, DeviceID), load from config file (`~/.ssh-vault/agent.json`), parse CLI flags with `flag.FlagSet` in `internal/agent/config.go`
- [x] T019 [US1] Implement agent sync loop: load config, start ticker at configured interval, on each tick: `GET /api/keys` with bearer token, parse response, call `keyblock.WriteBlock` to update authorized_keys managed block, log success/failure, handle graceful shutdown via signal in `internal/agent/agent.go`
- [x] T020 [US1] Wire `hub` subcommand: parse `--addr`, `--data`, `--password` flags (with `VAULT_PASSWORD` env fallback), initialize `FileStore` from data directory, start hub server in `cmd/ssh-vault/main.go`
- [x] T021 [US1] Wire `enroll` subcommand: parse `--hub-url`, `--token`, `--key`, `--name` flags, call enrollment flow, save returned config in `cmd/ssh-vault/main.go`
- [x] T022 [US1] Wire `agent` subcommand: parse `--hub-url`, `--interval`, `--key`, `--auth-keys` flags, load saved config from `~/.ssh-vault/agent.json`, start sync loop in `cmd/ssh-vault/main.go`

- [x] T022a Write table-driven tests for enrollment handlers: POST /api/enroll (valid token, expired token, used token, invalid SSH key, duplicate key), POST /api/enroll/verify (valid signature, invalid signature, wrong device ID) using `httptest.NewServer` in `internal/hub/handlers_test.go`
- [x] T022b Write table-driven tests for GET /api/keys: valid bearer token returns correct keys excluding own key, revoked token returns 401, unknown token returns 401 in `internal/hub/handlers_test.go`

**Checkpoint**: At this point, the full enrollment → approval → sync flow works end-to-end and is tested. A device can enroll, be approved, and receive other devices' keys in its authorized_keys file. This is the MVP.

---

## Phase 4: User Story 2 — Revoke a Compromised Device (Priority: P2)

**Goal**: Enable the owner to revoke a device from the hub, causing its key to be removed from all other devices' authorized_keys files within one sync cycle.

**Independent Test**: Enroll and approve three devices, revoke one via the hub, run agent sync on the remaining two, verify the revoked device's key is removed from both their authorized_keys managed blocks.

### Implementation for User Story 2

- [x] T023 [US2] Implement `POST /devices/{id}/revoke` handler: look up device by ID, verify status is "approved", transition to "revoked" via `device.Revoke()`, add audit entry (event: "revoked"), redirect to `/` in `internal/hub/handlers.go`. Note: API token is NOT cleared — the bearer auth middleware (T024) rejects revoked devices by status check, providing a clear "device revoked" error to the agent.
- [x] T024 [US2] Add revoked-device rejection to bearer auth middleware: when a device is found by API token but status is "revoked", respond with 401 and JSON error body `{error: "device revoked"}`, add audit entry (event: "auth_failed") in `internal/hub/server.go`
- [x] T025 [US2] Handle 401 "device revoked" response in agent sync loop: when `GET /api/keys` returns 401, log the rejection with `slog.Error`, stop the sync loop and exit with a non-zero status code indicating revocation in `internal/agent/agent.go`

- [x] T025a Write tests for revocation flow: revoke handler changes status, revoked device gets 401 on sync, other devices' GET /api/keys no longer includes revoked device's key in `internal/hub/handlers_test.go`

**Checkpoint**: Revocation propagates via the existing sync mechanism — revoking a device marks it as revoked, and the bearer auth middleware rejects its sync requests with a "device revoked" error. Other devices' next sync gets an updated key list excluding the revoked device.

---

## Phase 5: User Story 3 — Manage Devices via Web Dashboard (Priority: P3)

**Goal**: Provide a private web dashboard for viewing devices, their sync status, generating onboarding tokens, and performing approve/revoke actions.

**Independent Test**: Start the hub, access the dashboard in a browser, log in with the password, view the device list with correct status/fingerprint/last-sync data, generate a token, and copy it.

### Implementation for User Story 3

- [x] T026 [US3] Implement password session authentication: `POST /login` handler (compare password hash), `POST /logout` handler (clear session), in-memory session store (map + mutex), `requireSession` middleware that redirects to `/login` if no valid session cookie, cookie settings (HttpOnly, Secure, SameSite=Strict) in `internal/hub/auth.go`
- [x] T027 [US3] Create base HTML layout template with Pico CSS embedded via `//go:embed`, common header/nav, and template block structure in `internal/hub/templates/layout.html`
- [x] T028 [P] [US3] Create login page template with password form and error message display in `internal/hub/templates/login.html`
- [x] T029 [P] [US3] Create device list page template: table with columns (Name, Status, Fingerprint, Last Sync, Actions), status badges (pending/approved/revoked), stale device flag (>3x sync interval since last sync), approve button for pending devices, revoke button for approved devices, each action as a POST form in `internal/hub/templates/devices.html`
- [x] T030 [P] [US3] Create tokens page template: list of active (unused, unexpired) tokens with value and expiration, "Generate Token" POST form button, copiable token display in `internal/hub/templates/tokens.html`
- [x] T031 [US3] Download Pico CSS minified file and add to `internal/hub/templates/pico.min.css` for embedding
- [x] T032 [US3] Implement template rendering engine: parse all templates from embedded FS with `template.ParseFS`, define template functions (formatTime, formatFingerprint, isStale), render helper that executes template with data into ResponseWriter in `internal/hub/server.go`
- [x] T033 [US3] Implement dashboard route handlers: `GET /` (device list page with all devices from store), `GET /login` (login form), `GET /tokens` (token list + generate form), `POST /tokens` (generate new token, redirect to `/tokens`) in `internal/hub/handlers.go`
- [x] T034 [US3] Register all dashboard routes with `requireSession` middleware (except `/login` and `/api/*` routes), register static asset serving for Pico CSS in `internal/hub/server.go`
- [x] T035 [US3] Add audit log display to dashboard: `GET /audit` handler showing recent audit entries (enrollments, approvals, revocations, failed auth attempts) in reverse chronological order in `internal/hub/handlers.go` and new template `internal/hub/templates/audit.html`

- [x] T035a Write tests for session auth: login with correct/incorrect password, session cookie required for dashboard routes, logout clears session, API routes bypass session auth in `internal/hub/auth_test.go`

**Checkpoint**: The full dashboard experience works — login, view devices with live status, generate tokens, approve/revoke devices, and view audit log. All from the browser.

---

## Phase 6: User Story 4 — Offline Resilience (Priority: P4)

**Goal**: Ensure that when the hub is unreachable, agents retain the last synced authorized_keys state and resume normal operation when the hub returns.

**Independent Test**: Enroll two devices, verify SSH works between them, stop the hub, wait several sync intervals, verify SSH still works, restart the hub, verify agents resume syncing.

### Implementation for User Story 4

- [x] T036 [US4] Add graceful hub-unreachable handling to agent sync loop: on HTTP connection error or timeout, log with `slog.Warn` (not Error — this is expected), do NOT modify authorized_keys file, continue ticker for next retry in `internal/agent/agent.go`
- [x] T037 [US4] Add "never synced" state handling: if agent has never completed a successful sync (no keys in managed block and no cached response), log `slog.Info` indicating first sync pending, distinguish from "previously synced but hub is down" state in `internal/agent/agent.go`
- [x] T038 [US4] Add hub recovery handling: after one or more failed sync attempts, when hub becomes reachable again, log `slog.Info` indicating connection restored, apply the key set from the response as normal in `internal/agent/agent.go`

- [x] T038a Write tests for offline resilience: agent sync with unreachable hub does not modify authorized_keys file, agent resumes sync when hub returns, never-synced agent retries without error in `internal/agent/agent_test.go`

**Checkpoint**: The agent is resilient to hub outages — it never corrupts or clears the managed block when the hub is down, and seamlessly resumes when connectivity returns.

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [x] T039 [P] Add structured logging throughout: ensure all hub handlers and agent operations use `log/slog` with consistent fields (device_id, event, endpoint) across all files in `internal/hub/` and `internal/agent/`
- [x] T040 [P] Add graceful shutdown to hub server: listen for SIGINT/SIGTERM, call `http.Server.Shutdown` with timeout, flush pending store writes in `internal/hub/server.go`
- [x] T040a [P] Add optional TLS support to hub server: `--tls-cert` and `--tls-key` flags, if provided use `http.ListenAndServeTLS`, if omitted start plain HTTP with a `slog.Warn` log line stating "running without TLS — use a reverse proxy or SSH tunnel for encrypted connections" in `internal/hub/server.go`
- [x] T041 Validate quickstart.md end-to-end: follow each step in `specs/001-ssh-key-sync-hub/quickstart.md`, verify the build → hub start → token generate → enroll → approve → agent sync → SSH access flow works completely
- [x] T042 Run `go vet ./...` and `golangci-lint run`, fix any issues across all source files

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion — BLOCKS all user stories
- **User Story 1 (Phase 3)**: Depends on Foundational — delivers MVP
- **User Story 2 (Phase 4)**: Depends on Foundational — can start after Phase 2 (parallel with US1 if desired, but builds on US1's handlers)
- **User Story 3 (Phase 5)**: Depends on Foundational — can start after Phase 2 (parallel with US1/US2, adds dashboard layer)
- **User Story 4 (Phase 6)**: Depends on US1 (agent sync loop must exist) — adds resilience behavior
- **Polish (Phase 7)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) — no dependencies on other stories
- **User Story 2 (P2)**: Can start after Foundational, but practically extends US1's handlers and auth middleware — recommended after US1
- **User Story 3 (P3)**: Can start after Foundational — adds dashboard UI layer on top of US1/US2 endpoints
- **User Story 4 (P4)**: Depends on US1 (requires agent sync loop) — adds error handling behavior

### Within Each User Story

- API handlers before client code (hub before agent)
- Server/routing before handlers
- Core implementation before integration
- Story complete before moving to next priority

### Parallel Opportunities

- All Foundational tasks marked [P] can run in parallel (T004–T009)
- Within US3: template files (T028, T029, T030) can run in parallel
- Polish tasks T039, T040 can run in parallel

---

## Parallel Example: Foundational Phase

```bash
# Launch all model definitions together:
Task: "Create Device struct in internal/model/device.go"        # T004
Task: "Create Token struct in internal/model/token.go"          # T005
Task: "Create AuditEntry struct in internal/model/audit.go"     # T006
Task: "Create Store struct in internal/model/store.go"          # T007

# Launch keyblock operations together:
Task: "Atomic file write helper in internal/keyblock/atomic.go" # T008
Task: "Managed block operations in internal/keyblock/keyblock.go" # T009
```

## Parallel Example: User Story 3 Templates

```bash
# Launch all template files together:
Task: "Login page template in internal/hub/templates/login.html"    # T028
Task: "Device list template in internal/hub/templates/devices.html" # T029
Task: "Tokens page template in internal/hub/templates/tokens.html"  # T030
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL — blocks all stories)
3. Complete Phase 3: User Story 1 (Enrollment + Sync)
4. **STOP and VALIDATE**: Test full enrollment → approval → key sync flow
5. Deploy/demo if ready — this alone solves the "copy keys manually" problem

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 → Test enrollment flow → Deploy (MVP!)
3. Add User Story 2 → Test revocation → Deploy (MVP + security)
4. Add User Story 3 → Test dashboard → Deploy (full UX)
5. Add User Story 4 → Test offline resilience → Deploy (production-ready)
6. Each story adds value without breaking previous stories

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- The spec does not request tests — add them if desired but they are not in the task list
- Avoid: vague tasks, same file conflicts, cross-story dependencies that break independence
