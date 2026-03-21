# Tasks: Self-Serve Binary Distribution

**Input**: Design documents from `/specs/004-self-serve-binaries/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, contracts/

**Tests**: Not explicitly requested. Tests are included because the existing codebase has thorough test coverage and the constitution mandates tests for non-trivial code.

**Organization**: Tasks are grouped by user story. US1 and US2 are both P1 and must ship together, but are structured as separate phases for traceability.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Foundational (Config & Struct Changes)

**Purpose**: Replace `githubRepo`/`releaseTag` with `distDir` in the Server struct, ServerConfig, and CLI flags. This is the prerequisite for all user stories.

**⚠️ CRITICAL**: No user story work can begin until this phase is complete.

- [x] T001 Replace `githubRepo` and `releaseTag` fields with `distDir string` in `Server` struct and `ServerConfig` struct in `internal/hub/server.go`. Update `NewServer` to assign `distDir` from config.
- [x] T002 Replace `--github-repo` and `--release-tag` flags with `--dist-dir` flag (env: `VAULT_DIST_DIR`) in `cmd/ssh-vault/main.go`. Pass `DistDir` to `ServerConfig` instead of `GithubRepo`/`ReleaseTag`.
- [x] T003 Register new route `GET /download/` in `registerRoutes()` in `internal/hub/server.go`, pointing to a download handler method on Server.

**Checkpoint**: Hub compiles with new `--dist-dir` flag. Old GitHub flags are removed. Route is registered but handler not yet implemented.

---

## Phase 2: User Story 1 — Self-Hosted Binary Distribution (Priority: P1) 🎯 MVP

**Goal**: Hub serves pre-built binaries from a local directory via `GET /download/{os}/{arch}`.

**Independent Test**: Place a binary in a temp dist directory, start the hub with `--dist-dir`, and `curl` the download endpoint.

### Implementation for User Story 1

- [x] T004 [US1] Create `internal/hub/handlers_download.go` with `handleDownload` method on Server. Implement: parse `{os}` and `{arch}` from URL path, validate against allowlist (`linux`/`darwin` × `amd64`/`arm64`), construct filename `ssh-vault_{os}_{arch}`, serve file from `s.distDir` using `http.ServeFile`. Return 501 if `distDir` is empty, 400 for invalid platform, 404 if file not found. Also handle `GET /download/checksums.txt` as a special case.
- [x] T005 [US1] Create `internal/hub/handlers_download_test.go` with table-driven tests covering: successful binary download (200 + correct headers), invalid OS/arch (400), missing binary (404), dist-dir not configured (501), checksums.txt serving, and directory traversal attempts.

**Checkpoint**: `GET /download/linux/amd64` serves a binary from the dist directory. All acceptance scenarios for US1 pass.

---

## Phase 3: User Story 2 — Enrollment Script Uses Hub for Downloads (Priority: P1)

**Goal**: The enrollment script downloads the agent binary from the hub instead of GitHub Releases.

**Independent Test**: Generate an enrollment link, curl the script, verify download URLs point to `{external-url}/download/{os}/{arch}`.

### Implementation for User Story 2

- [x] T006 [US2] Update `buildEnrollmentScript` in `internal/hub/handlers_shortcode.go`: change download URL template from GitHub Releases format to `{external-url}/download/{os}/{arch}`. Update `BINARY_URL` to use `${VAULT_DOWNLOAD_BASE}/${OS}/${ARCH}` and `CHECKSUMS_URL` to `${VAULT_DOWNLOAD_BASE}/checksums.txt`. The function signature changes: replace `downloadBaseURL` parameter with the hub's external URL.
- [x] T007 [US2] Update `handleShortCodeEnroll` in `internal/hub/handlers_shortcode.go`: remove the `downloadBaseURL` construction that uses `s.githubRepo` and `s.releaseTag`. Instead pass `s.externalURL + "/download"` to `buildEnrollmentScript`.
- [x] T008 [US2] Update `handleGenerateLink` in `internal/hub/handlers_shortcode.go`: change the configuration guard from checking `s.externalURL == "" || s.githubRepo == ""` to checking only `s.externalURL == ""`.
- [x] T009 [US2] Update tests in `internal/hub/handlers_shortcode_test.go`: update `testServerWithEnrollment` to use `DistDir` instead of `GithubRepo`/`ReleaseTag`. Update `TestHandleShortCodeEnroll_ScriptContainsTemplateVars` to assert the download URL points to the hub (`/download/`) instead of GitHub (`releases/download`). Update `TestHandleGenerateLink_NotConfigured` if needed.

**Checkpoint**: Enrollment script downloads binary from `{external-url}/download/{os}/{arch}`. All shortcode tests pass with new URL format.

---

## Phase 4: User Story 3 — Simplified Hub Configuration (Priority: P2)

**Goal**: Clean operator experience with a single `--dist-dir` flag, Docker/env updated, old GitHub config fully removed.

**Independent Test**: Start hub with only `--dist-dir`, run full enrollment flow, verify no GitHub references remain.

### Implementation for User Story 3

- [x] T010 [P] [US3] Update `docker-compose.yml`: remove `VAULT_GITHUB_REPO` and `VAULT_RELEASE_TAG` env vars, add `VAULT_DIST_DIR=/dist` env var and `./dist:/dist:ro` volume mount.
- [x] T011 [P] [US3] Update `Dockerfile`: add `RUN mkdir /dist && chown vault:vault /dist` and document the dist volume mount.
- [x] T012 [P] [US3] Update `.env.example`: remove `VAULT_GITHUB_REPO` and `VAULT_RELEASE_TAG` entries, add `VAULT_DIST_DIR` entry with documentation comment.

**Checkpoint**: Docker and env config reference only `--dist-dir`. No GitHub config references remain in infrastructure files.

---

## Phase 5: Polish & Cross-Cutting Concerns

**Purpose**: Documentation updates and final validation.

- [x] T013 Update `README.md`: replace `--github-repo` and `--release-tag` references with `--dist-dir` in CLI examples, Quick Start section, and CLI Reference section. Update the enrollment flow description to mention hub-served binaries instead of GitHub Releases.
- [x] T014 Run `go test ./...`, `go vet ./...`, and `golangci-lint run` to verify all tests pass, no vet issues, and no lint violations remain.
- [x] T015 Run quickstart.md validation: build binaries for the local platform into a `dist/` directory, start the hub with `--dist-dir ./dist`, generate an enrollment link, and verify the enrollment script serves the correct download URLs.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Foundational (Phase 1)**: No dependencies — can start immediately
- **US1 (Phase 2)**: Depends on Foundational (Phase 1) — needs `distDir` on Server struct and route registered
- **US2 (Phase 3)**: Depends on Foundational (Phase 1) — needs old fields removed from struct. Can run in parallel with US1 (different files)
- **US3 (Phase 4)**: Depends on Foundational (Phase 1) — infrastructure config cleanup. Can run in parallel with US1 and US2 (different files)
- **Polish (Phase 5)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Depends on Phase 1. Produces `handlers_download.go` (new file — no conflicts).
- **User Story 2 (P1)**: Depends on Phase 1. Modifies `handlers_shortcode.go` and `handlers_shortcode_test.go`.
- **User Story 3 (P2)**: Depends on Phase 1. Modifies Docker/env files only — no Go code conflicts.

### Within Each User Story

- Core handler implementation before tests (tests need the handler to exist for compilation)
- Config changes before handler changes

### Parallel Opportunities

- T010, T011, T012 (US3) are all [P] — different files, can run in parallel
- US1 (handlers_download.go) and US2 (handlers_shortcode.go) touch different files and can run in parallel after Phase 1
- US3 touches only infrastructure files (Docker, env) — can run in parallel with US1 and US2

---

## Parallel Example: After Phase 1

```text
# These can all run in parallel after Foundational phase:

Agent A (US1): "Create handlers_download.go with download handler"
Agent B (US2): "Update handlers_shortcode.go enrollment script URLs"
Agent C (US3): "Update docker-compose.yml, Dockerfile, .env.example"
```

---

## Implementation Strategy

### MVP First (US1 + US2 Together)

1. Complete Phase 1: Foundational (struct + CLI changes)
2. Complete Phase 2: US1 (download handler)
3. Complete Phase 3: US2 (enrollment script)
4. **STOP and VALIDATE**: `go test ./...` passes, enrollment flow works end-to-end
5. Complete Phase 4: US3 (Docker/env cleanup)
6. Complete Phase 5: Polish (README, final validation)

### Incremental Delivery

1. Phase 1 → Hub compiles with `--dist-dir`
2. Phase 2 → Binary downloads work via `/download/{os}/{arch}`
3. Phase 3 → Enrollment scripts use hub instead of GitHub
4. Phase 4 → Docker and env fully updated
5. Phase 5 → Documentation current, all tests green

---

## Notes

- US1 and US2 are both P1 and must both be complete for the feature to be functional (the enrollment script needs the download endpoint, and the download endpoint needs the enrollment script to point to it).
- The feature is a net simplification: removes 2 config fields, adds 1. Removes external dependency on GitHub.
- No new Go dependencies. No new packages. No new data model entities.
