# Tasks: Device Rename

**Input**: Design documents from `/specs/007-device-rename/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Foundational (Blocking Prerequisites)

**Purpose**: Add the model-layer constant and server-side routing needed by all user stories

- [x] T001 Add `EventDeviceRenamed = "device_renamed"` constant in `internal/model/audit.go`
- [x] T002 Add `"/rename"` case to `handleDeviceAction` switch in `internal/hub/server.go` (route to `s.handleRename`)
- [x] T003 Add `"device_renamed"` → `"used"` mapping in `eventPillClass` func in `internal/hub/server.go`

**Checkpoint**: New event constant exists, route is wired (handler not yet implemented)

---

## Phase 2: User Story 1 — Inline Device Rename (Priority: P1) 🎯 MVP

**Goal**: Users can click a device name, edit it inline, and save via blur/Enter. Escape cancels. Server validates and persists the new name.

**Independent Test**: Click a device name on the dashboard, type a new name, click away, refresh the page — the new name persists.

### Implementation for User Story 1

- [x] T004 [US1] Implement `handleRename` handler in `internal/hub/handlers.go` — accept `POST /devices/{id}/rename` with JSON body `{"name": "..."}`, trim whitespace, validate non-empty and max 255 chars, skip save if unchanged, call `store.UpdateDevice`, return JSON `{"name": "..."}` on success or `{"error": "..."}` on failure
- [x] T005 [US1] Update device name markup in `internal/hub/templates/devices.html` — wrap `{{.Name}}` in a clickable `<span>` with `data-device-id="{{.ID}}"` and a CSS class for the editable name element
- [x] T006 [US1] Add inline edit JavaScript in `internal/hub/templates/devices.html` — on click: replace `<span>` text with `<input>` pre-filled with current name (`maxlength="255"`), auto-focus; on blur/Enter: POST to `/devices/{id}/rename` via `fetch()`, update DOM on success, restore on error; on Escape: restore original text; skip POST if trimmed value equals original; client-side reject empty name

**Checkpoint**: Inline rename is fully functional — click, edit, save, cancel all work. Audit not yet wired.

---

## Phase 3: User Story 2 — Audit Trail for Renames (Priority: P1)

**Goal**: Every successful rename creates an audit log entry with old and new name details. Failed/cancelled renames produce no audit entry.

**Independent Test**: Rename a device, navigate to the audit log page — a `device_renamed` event appears with `"Device renamed from 'old' to 'new'"` in the details column.

### Implementation for User Story 2

- [x] T007 [US2] Add audit logging to `handleRename` in `internal/hub/handlers.go` — after successful `UpdateDevice`, call `s.store.AddAuditEntry(model.NewAuditEntry(model.EventDeviceRenamed, deviceID, fmt.Sprintf("Device renamed from '%s' to '%s'", oldName, newName)))`. Ensure no audit entry is created for validation failures or no-op (unchanged name).

**Checkpoint**: Rename + audit trail fully functional. All P1 stories complete.

---

## Phase 4: User Story 3 — Visual Feedback During Rename (Priority: P2)

**Goal**: Users get clear visual cues that names are editable (hover state), see progress during save, and get error messages on failure.

**Independent Test**: Hover over a device name — cursor/style changes. Rename a device — brief saving indicator appears. Simulate a server error — error message shown, original name restored.

### Implementation for User Story 3

- [x] T008 [US3] Add hover styles for editable device names in `internal/hub/templates/devices.html` `<style>` block — cursor pointer, subtle underline or edit icon on hover to indicate editability (FR-009)
- [x] T009 [US3] Add saving/error feedback to inline edit JavaScript in `internal/hub/templates/devices.html` — show brief "Saving..." text or opacity change during `fetch()`, show error message on non-OK response or network failure, restore original name on any error (FR-010)

**Checkpoint**: All user stories complete with full visual feedback.

---

## Phase 5: Polish & Cross-Cutting Concerns

**Purpose**: Tests and final validation

- [x] T010 Add table-driven tests for `handleRename` in `internal/hub/handlers_test.go` — cover: successful rename, empty name rejected, name too long rejected, whitespace-only name rejected, unchanged name no-op, rename revoked device (succeeds — name is a display label), device not found 404, invalid JSON 400, method not allowed 405. Verify audit entry created on success and absent on failure/no-op.
- [x] T011 Run `go test ./...`, `go vet ./...` and verify all tests pass
- [x] T012 Manual smoke test per `specs/007-device-rename/quickstart.md` — rename a device, verify audit log, test edge cases (empty, Escape, unchanged)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Foundational)**: No dependencies — can start immediately
- **Phase 2 (US1)**: Depends on T001 (event constant) and T002 (route wiring)
- **Phase 3 (US2)**: Depends on T004 (handler must exist to add audit call)
- **Phase 4 (US3)**: Depends on T005, T006 (markup and JS must exist to add styles/feedback)
- **Phase 5 (Polish)**: Depends on all implementation phases

### Within Each Phase

- T001, T002, T003 can all run in parallel (different locations in different files or same file different functions)
- T005, T006 are sequential (T006 JS depends on T005 markup)
- T008, T009 can run in parallel (CSS vs JS, but same file — sequential is safer)

### Parallel Opportunities

```
Phase 1: T001 ─┐
         T002 ─┼─ All parallel (different files/functions)
         T003 ─┘
                ↓
Phase 2: T004 → T005 → T006 (sequential chain)
                ↓
Phase 3: T007 (single task, depends on T004)
                ↓
Phase 4: T008 → T009 (sequential, same file)
                ↓
Phase 5: T010 → T011 → T012 (sequential)
```

---

## Implementation Strategy

### MVP First (User Stories 1 + 2)

1. Complete Phase 1: Add constant + route + pill mapping
2. Complete Phase 2: Handler + inline edit UI
3. Complete Phase 3: Add audit logging to handler
4. **STOP and VALIDATE**: Test rename + audit independently
5. Deploy if ready — visual polish can follow

### Full Delivery

1. MVP (above)
2. Add Phase 4: Hover styles + saving/error feedback
3. Add Phase 5: Automated tests + smoke test
4. Feature complete

---

## Notes

- All changes are in existing files — no new files created (except spec artifacts)
- The handler follows the established pattern: `extractPathParam` → `GetDevice` → modify → `UpdateDevice` → audit
- The `eventPillClass` mapping ensures the new event renders correctly on the audit page without any HTML template changes to `audit.html`
- Client-side and server-side validation are both required (FR-004, FR-005, R-005)
