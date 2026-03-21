# Tasks: Token Management Enhancements

**Input**: Design documents from `/specs/002-token-management/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, quickstart.md

**Tests**: Included — constitution principle II mandates tests for all non-trivial code.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Foundational (Blocking Prerequisites)

**Purpose**: Shared infrastructure that ALL user stories depend on — new audit event constants and store methods.

**⚠️ CRITICAL**: No user story work can begin until this phase is complete.

- [x] T001 [P] Add `EventTokenUsed` and `EventTokenRemoved` constants to `internal/model/audit.go`
- [x] T002 [P] Add `RemoveToken(value string) error` method to `internal/hub/store.go` — acquire write lock, find token by value, reject if used (`"cannot remove used token"`), reject if not found, remove via swap-with-last pattern, call Save()
- [x] T003 Add `PurgeExpiredTokens() (int, error)` method to `internal/hub/store.go` — acquire write lock, filter tokens in-place to retain only non-expired tokens (keep used tokens as they are already filtered from the UI), call Save() only if count > 0, return purged count
- [x] T004 Add tests for `RemoveToken` and `PurgeExpiredTokens` in `internal/hub/store_test.go` — table-driven tests: remove valid token, remove used token (error), remove not found (error), remove preserves others, purge expired mix, purge when none expired

**Checkpoint**: Store layer complete — `RemoveToken`, `PurgeExpiredTokens` tested and working.

---

## Phase 2: User Story 1 — Remove Unused Token (Priority: P1) 🎯 MVP

**Goal**: Administrators can remove unused tokens from the dashboard with one click + confirmation, and the removal is audited and enforced immediately.

**Independent Test**: Generate a token → click Remove → confirm → verify token gone from list → attempt enrollment with that token → verify rejection.

### Implementation for User Story 1

- [x] T005 [US1] Add `handleRemoveToken` handler in `internal/hub/handlers.go` — extract token value from URL path using `extractPathParam(path, "/tokens/", "/remove")`, validate token exists and is unused via `store.GetToken()`, call `store.RemoveToken(value)`, create `token_removed` audit entry with `value[:8]` prefix and empty DeviceID, redirect to `/tokens`
- [x] T006 [US1] Add `handleTokenAction` router and register `/tokens/` route in `internal/hub/server.go` — route paths ending in `/remove` to `handleRemoveToken`; add `s.store.PurgeExpiredTokens()` call at the start of `handleTokens` GET handler (before listing tokens), log purged count via `slog.Info` if > 0
- [x] T007 [US1] Add remove button with confirmation dialog to each token row in `internal/hub/templates/tokens.html` — add "Actions" column header; per row add `<form method="post" action="/tokens/{{.Value}}/remove">` with delete button and `onclick="return confirm('Remove this token?')"`
- [x] T008 [P] [US1] Add CSS for `.badge-token_removed` (orange `#fd7e14`) in `internal/hub/templates/layout.html`
- [x] T009 [US1] Add handler tests for token removal in `internal/hub/handlers_test.go` — test cases: POST remove success (verify redirect + token gone), POST remove used token (verify error), POST remove not found (verify error), GET /tokens purges expired tokens, POST /api/enroll with a previously removed token value (verify 401 rejection — validates FR-003)

**Checkpoint**: User Story 1 fully functional — tokens can be removed, removal is audited, expired tokens are purged, enrollment with removed tokens is rejected.

---

## Phase 3: User Story 2 — Token Usage Audit Trail (Priority: P2)

**Goal**: A `token_used` audit entry appears in the audit log whenever a token is consumed during device enrollment, with a truncated token prefix and the enrolling device identity.

**Independent Test**: Generate a token → enroll a device with it → check audit log → verify `TOKEN_USED` event with correct device and token prefix.

### Implementation for User Story 2

- [x] T010 [US2] Add `token_used` audit entry in `handleEnroll` in `internal/hub/handlers.go` — after existing `EventEnrolled` audit entry (line ~131), add `s.store.AddAuditEntry(model.NewAuditEntry(model.EventTokenUsed, deviceID, fmt.Sprintf("Token %s... used by device '%s'", req.Token[:8], req.Name)))`
- [x] T011 [P] [US2] Add CSS for `.badge-token_used` (blue `#17a2b8`) in `internal/hub/templates/layout.html`
- [x] T012 [US2] Add handler test verifying `token_used` audit entry in `internal/hub/handlers_test.go` — enroll a device via POST /api/enroll with valid token, verify audit log contains entry with Event == "token_used" and correct device ID

**Checkpoint**: User Stories 1 AND 2 both work independently. Audit log shows both `token_used` and `token_removed` events with distinct badge colors.

---

## Phase 4: User Story 3 — Copy Token to Clipboard (Priority: P3)

**Goal**: A copy icon next to each token allows one-click clipboard copy with visual feedback.

**Independent Test**: Generate a token → click copy icon → paste into text editor → verify full 64-char token value matches.

### Implementation for User Story 3

- [x] T013 [US3] Add copy-to-clipboard icon button with inline JavaScript in `internal/hub/templates/tokens.html` — add clipboard SVG icon button per token row with `onclick` calling `navigator.clipboard.writeText('{{.Value}}')`, on success change icon to checkmark for 1.5 seconds, on failure show alert with fallback message
- [x] T014 [P] [US3] Add CSS for `.copy-btn` styling in `internal/hub/templates/layout.html` — inline button, no background/border, cursor pointer, subtle opacity hover effect, position adjacent to token value

**Checkpoint**: All three user stories fully functional and independently testable.

---

## Phase 5: Polish & Cross-Cutting Concerns

**Purpose**: Verification across all stories, regression testing, and build validation.

- [x] T015 Run `go build ./...` to verify compilation of all changes
- [x] T016 Run `go vet ./...` for static analysis
- [x] T017 Run `go test ./...` full test suite — verify all new and existing tests pass
- [x] T018 Validate quickstart.md scenarios: token removal, audit trail, copy-to-clipboard, expired purge

---

## Dependencies & Execution Order

### Phase Dependencies

- **Foundational (Phase 1)**: No dependencies — can start immediately. BLOCKS all user stories.
- **US1 (Phase 2)**: Depends on Phase 1 completion. Core MVP.
- **US2 (Phase 3)**: Depends on Phase 1 completion. Independent of US1 (different handler, different file sections).
- **US3 (Phase 4)**: Depends on Phase 1 completion. Fully independent of US1 and US2 (UI-only, no backend changes).
- **Polish (Phase 5)**: Depends on all user stories being complete.

### User Story Dependencies

- **US1 (Remove Token)**: Requires `RemoveToken()`, `PurgeExpiredTokens()` from Phase 1. No dependency on US2 or US3.
- **US2 (Token Usage Audit)**: Requires `EventTokenUsed` constant from Phase 1. No dependency on US1 or US3.
- **US3 (Copy to Clipboard)**: Requires only template changes. No dependency on US1 or US2.

### Within Each User Story

- Backend handler before template changes (handler defines the route the template targets)
- Route registration before template (template forms POST to the route)
- CSS can run in parallel with implementation (different file)
- Tests after implementation (tests validate the handler behavior)

### Parallel Opportunities

- **Phase 1**: T001 (audit.go) and T002 (store.go) can run in parallel — different files
- **Phase 2–4**: After Phase 1, all three user stories can run in parallel — they touch different sections of files and have no cross-dependencies
- **Within each story**: CSS tasks marked [P] can run in parallel with other story tasks

---

## Parallel Example: After Phase 1 Completion

```text
# All three user stories can start simultaneously:

# Agent A: User Story 1 (Remove Token)
Task: "T005 Add handleRemoveToken handler in internal/hub/handlers.go"
Task: "T006 Add handleTokenAction router in internal/hub/server.go"
Task: "T007 Add remove button in internal/hub/templates/tokens.html"

# Agent B: User Story 2 (Token Usage Audit)
Task: "T010 Add token_used audit entry in handleEnroll in internal/hub/handlers.go"
Task: "T011 Add CSS for badge-token_used in internal/hub/templates/layout.html"

# Agent C: User Story 3 (Copy to Clipboard)
Task: "T013 Add copy icon button in internal/hub/templates/tokens.html"
Task: "T014 Add CSS for copy-btn in internal/hub/templates/layout.html"
```

**⚠️ Note**: US1 and US2 both modify `handlers.go` and US1/US3 both modify `tokens.html`, so true parallel execution requires careful merge. Sequential execution within a single agent is recommended for safety.

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Foundational (T001–T004)
2. Complete Phase 2: User Story 1 (T005–T009)
3. **STOP and VALIDATE**: Test token removal end-to-end
4. The hub now supports removing unused tokens with audit trail and expired token cleanup

### Incremental Delivery

1. Phase 1 → Foundation ready
2. Add US1 (Remove Token) → Test → Deploy (MVP!)
3. Add US2 (Token Usage Audit) → Test → Deploy
4. Add US3 (Copy to Clipboard) → Test → Deploy
5. Each story adds value without breaking previous stories

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Constitution mandates tests (Principle II) — test tasks included
- All store methods use existing `sync.RWMutex` for thread safety
- Templates use `//go:embed` — changes compile into the binary
- Commit after each phase completion for clean git history
