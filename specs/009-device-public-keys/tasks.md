# Tasks: Device Public Keys Display

**Input**: Design documents from `/specs/009-device-public-keys/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md

**Tests**: Included — table-driven tests for new parsing functions per Constitution principle II.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2)
- Include exact file paths in descriptions

---

## Phase 1: Foundational (Template Functions + Tests)

**Purpose**: Add SSH key parsing template functions that both user stories depend on

**⚠️ CRITICAL**: Both user stories need these template functions before UI work can begin

- [x] T001 Add `keyUser(publicKey string) string` template function to funcMap in `internal/hub/server.go` — parse SSH public key using `ssh.ParseAuthorizedKey` to get the comment, split comment on `@` via `strings.Cut`, return the username portion (before `@`). Return empty string if no comment. If comment has no `@`, return the full comment as the "user" display value (per FR-003).
- [x] T002 Add `keyHost(publicKey string) string` template function to funcMap in `internal/hub/server.go` — extract host portion (after `@`) from the SSH public key comment. Return empty string if no comment or no `@`.
- [x] T003 Add `keyType(publicKey string) string` template function to funcMap in `internal/hub/server.go` — parse key via `ssh.ParseAuthorizedKey`, return `Key.Type()` with the `ssh-` prefix stripped (e.g., `"ssh-ed25519"` → `"ed25519"`). Return empty string on parse failure.
- [x] T004 Add table-driven tests for `keyUser`, `keyHost`, `keyType` in `internal/hub/server_test.go` — cover: standard `user@host` comment, comment without `@` (freeform label), empty comment, no comment field at all, RSA vs ed25519 key types, malformed key string.

**Checkpoint**: Template functions tested and ready. UI work can now begin.

---

## Phase 2: User Story 1 — View Key Metadata on Devices Page (Priority: P1) 🎯 MVP

**Goal**: Replace truncated fingerprint with parsed key metadata (key type, username, host/IP) on the Devices page

**Independent Test**: Enroll devices with various SSH keys (with/without comments) and verify the Devices page shows key type + username + host instead of fingerprint. Verify fallback to fingerprint for keys without comments.

### Implementation for User Story 1

- [x] T005 [US1] Update device cell in `internal/hub/templates/devices.html` — replace `{{formatFingerprint .Fingerprint}}` in the `.device-info-fp` div with key metadata display: show key type badge + `user@host` when available, fall back to `{{formatFingerprint .Fingerprint}}` when `keyUser` returns empty. Use conditional template logic: `{{if keyUser .PublicKey}}` for the metadata path, `{{else}}` for fingerprint fallback.
- [x] T006 [US1] Add CSS styles for key metadata display in `internal/hub/templates/theme.css` — add `.key-type` badge style (small pill/tag, monospace font, muted background), `.key-identity` style for the `user@host` text (monospace, same size as existing `.device-info-fp`). Maintain existing responsive behavior.

**Checkpoint**: Devices page shows key metadata instead of fingerprint. MVP complete.

---

## Phase 3: User Story 2 — View Full Public Key on Demand (Priority: P2)

**Goal**: Add expand/collapse to reveal and copy the full SSH public key per device

**Independent Test**: Click expand on any device row, verify full key appears in monospace pre block, click copy button and verify clipboard contains valid SSH key, click collapse to hide.

### Implementation for User Story 2

- [x] T007 [US2] Add expand/collapse toggle button and collapsible key container in `internal/hub/templates/devices.html` — add a small "key" icon button in the device cell that toggles a hidden `<div>` below the device row containing a `<pre>` block with `{{.PublicKey}}` and a "Copy" button. Use vanilla JavaScript matching existing `editable-name` interaction pattern (event delegation on `document`).
- [x] T008 [US2] Add CSS styles for expand/collapse and full key display in `internal/hub/templates/theme.css` — add `.key-expand-btn` toggle button style, `.key-detail` collapsible container (hidden by default, border-top separator), `.key-detail pre` monospace block style, `.key-copy-btn` copy button style. Ensure the expanded key container spans the full table row width.
- [x] T009 [US2] Add vanilla JavaScript for copy-to-clipboard in `internal/hub/templates/devices.html` — implement click handler on copy button using `navigator.clipboard.writeText()`, show brief "Copied!" feedback (swap button text for 2 seconds), with fallback for older browsers using `document.execCommand('copy')`.

**Checkpoint**: Full key expand/collapse works. Both user stories complete.

---

## Phase 4: Polish & Cross-Cutting Concerns

**Purpose**: Responsive behavior and edge case validation

- [x] T010 Verify responsive layout in `internal/hub/templates/theme.css` — ensure key metadata and expand/collapse work on narrow viewports (mobile-width). Adjust truncation/ellipsis if needed for long hostnames. Test that expanded key container scrolls horizontally for long keys.
- [x] T011 Run `go vet ./...`, `golangci-lint run`, and `go test ./...` to verify all tests pass, no lint violations, and no static analysis issues

---

## Dependencies & Execution Order

### Phase Dependencies

- **Foundational (Phase 1)**: No dependencies — start immediately
- **User Story 1 (Phase 2)**: Depends on Phase 1 (template functions must exist)
- **User Story 2 (Phase 3)**: Depends on Phase 1 (template functions). Can run in parallel with US1 since they touch different parts of the template, but sequential is safer to avoid merge conflicts in devices.html.
- **Polish (Phase 4)**: Depends on Phase 2 + Phase 3

### User Story Dependencies

- **User Story 1 (P1)**: Depends only on Foundational — no dependency on US2
- **User Story 2 (P2)**: Depends only on Foundational — no dependency on US1 (but recommended to do after US1 for cleaner diffs)

### Within Each User Story

- Template HTML before CSS (styling references HTML classes)
- Copy JS after expand/collapse structure exists

### Parallel Opportunities

- T001, T002, T003 can be written together (same file, but logically independent functions)
- T005 and T006 can be developed in parallel [P] (different files)
- T007 and T008 can be developed in parallel [P] (different files)

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Template functions + tests (T001–T004)
2. Complete Phase 2: Key metadata display (T005–T006)
3. **STOP and VALIDATE**: Enroll devices, verify metadata shows on Devices page
4. Deploy if ready — users immediately get the key identification improvement

### Incremental Delivery

1. Phase 1 → Template functions ready
2. Phase 2 (US1) → Key metadata visible → Deploy (MVP!)
3. Phase 3 (US2) → Expand/collapse for full key → Deploy
4. Phase 4 → Polish and verify → Final deploy

---

## Notes

- All new Go code goes in `internal/hub/server.go` (functions) and `internal/hub/server_test.go` (tests)
- All UI changes in `internal/hub/templates/devices.html` and `internal/hub/templates/theme.css`
- No changes to `internal/model/` — no new persistence
- Uses existing dependency `golang.org/x/crypto/ssh` for key parsing (already imported elsewhere)
- Commit after each phase completion
