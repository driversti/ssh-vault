# Tasks: Paginated Audit Log

**Input**: Design documents from `/specs/008-paginated-audit-log/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: Included — constitution requires tests for all non-trivial code.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2)
- Include exact file paths in descriptions

---

## Phase 1: Foundational (Pagination Infrastructure)

**Purpose**: Create the pagination helper that both user stories depend on

- [x] T001 [P] Create `PageItem` and `PaginationData` structs and `calcPagination(totalItems, page, pageSize int) PaginationData` function in `internal/hub/pagination.go`. Implements: page clamping (< 1 → 1, > totalPages → last), page count math (`ceil(total/pageSize)`), HasPrev/HasNext/PrevPage/NextPage computation, and `Pages` slice generation with ellipsis gaps (show all numbers when ≤ 7 pages; use `1 ... current-1 current current+1 ... last` pattern for more). Non-numeric or missing page values default to 1.
- [x] T002 [P] Write table-driven tests in `internal/hub/pagination_test.go` for `calcPagination`: cases include 0 items (0 pages), 1 item (1 page, no nav), 20 items (1 page, no nav), 21 items (2 pages), 45 items (3 pages), page clamping (page=0 → 1, page=999 → last), ellipsis generation for 8+ pages at various current positions (first, middle, last), and correct HasPrev/HasNext/PrevPage/NextPage values.

**Checkpoint**: `go test ./internal/hub/... -run TestCalcPagination` passes. Foundation ready for user story implementation.

---

## Phase 2: User Story 1 - View Audit Log with Pagination (Priority: P1) 🎯 MVP

**Goal**: Audit log page displays at most 20 entries with Previous/Next navigation and total entry count.

**Independent Test**: Open `/audit`, verify max 20 entries shown. Click Next/Previous to navigate. Verify total count displayed in header. Verify no controls when ≤ 20 entries.

### Implementation for User Story 1

- [x] T003 [US1] Update `handleAudit()` in `internal/hub/server.go`: parse `page` query parameter via `r.URL.Query().Get("page")` and `strconv.Atoi`, call `calcPagination(len(entries), page, 20)`, slice the reversed entries array to the current page window (`start = (page-1)*pageSize`, `end = min(start+pageSize, len(entries))`), and pass both the sliced `Entries` and `Pagination PaginationData` to the template data map.
- [x] T004 [US1] Update `internal/hub/templates/audit.html`: add total entry count to the header (e.g., `Audit Log ({{.Pagination.TotalItems}} entries)`), keep existing table rendering for `{{range .Entries}}` (now receives only the page slice), and add pagination controls below the table — a `<nav>` with Previous link (`/audit?page={{.Pagination.PrevPage}}`, disabled when `!.Pagination.HasPrev`) and Next link (`/audit?page={{.Pagination.NextPage}}`, disabled when `!.Pagination.HasNext`). Hide entire pagination `<nav>` when `TotalPages` ≤ 1. Style pagination controls to match existing UI (use existing CSS variables and pill/button patterns).
- [x] T005 [US1] Handle edge cases in `internal/hub/server.go`: empty audit log (0 entries → empty page, no pagination), and entry count display showing "0 entries" / "1 entry" / "N entries" (pluralization).

**Checkpoint**: `go build ./cmd/ssh-vault` succeeds. Manual test: `/audit` shows max 20 entries, Previous/Next work, total count shown, no controls for ≤ 20 entries.

---

## Phase 3: User Story 2 - Navigate to Specific Page (Priority: P2)

**Goal**: Users can click specific page numbers to jump directly to any page. Current page is visually highlighted.

**Independent Test**: With 100+ audit entries, verify page number links appear. Click page 3, verify correct entries and page 3 highlighted. Verify ellipsis appears for 8+ pages.

### Implementation for User Story 2

- [x] T006 [US2] Enhance pagination controls in `internal/hub/templates/audit.html`: between Previous and Next links, add `{{range .Pagination.Pages}}` rendering — for each `PageItem`, if `.IsGap` render `<span>…</span>`, else render `<a href="/audit?page={{.Number}}">{{.Number}}</a>` with active/current styling when `.IsActive` (e.g., bold, different background using existing CSS variables). Ensure the current page link is visually distinct (not clickable or styled as active state).

**Checkpoint**: `go build ./cmd/ssh-vault` succeeds. Manual test: page numbers visible, clicking any number navigates correctly, current page highlighted, ellipsis shows for many pages.

---

## Phase 4: Polish & Cross-Cutting Concerns

**Purpose**: Final verification and cleanup

- [x] T007 Run `go vet ./...`, `golangci-lint run`, and `go test ./...` — fix any issues
- [x] T008 Run `go build -o ssh-vault ./cmd/ssh-vault` and manually verify: empty audit log, 1 entry, 20 entries (no pagination), 21 entries (2 pages), 45 entries (3 pages), invalid page param (`?page=abc`, `?page=0`, `?page=999`)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Foundational (Phase 1)**: No dependencies — can start immediately
- **User Story 1 (Phase 2)**: Depends on Phase 1 completion (needs `calcPagination`)
- **User Story 2 (Phase 3)**: Depends on Phase 2 completion (needs handler + template base from US1)
- **Polish (Phase 4)**: Depends on all user stories being complete

### Within Each Phase

- T001 and T002 can run in parallel (different files)
- T003 → T004 → T005 are sequential (T004 needs template data from T003, T005 refines T003)
- T006 depends on T004 (extends the template)
- T007 and T008 are sequential (fix before verify)

### Parallel Opportunities

```bash
# Phase 1: Both tasks in parallel
Task: "Create pagination types and calcPagination in internal/hub/pagination.go"
Task: "Write table-driven tests in internal/hub/pagination_test.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Foundational (T001, T002)
2. Complete Phase 2: User Story 1 (T003, T004, T005)
3. **STOP and VALIDATE**: Test pagination with Previous/Next navigation
4. Deploy/demo if ready — core pagination is fully usable

### Incremental Delivery

1. Phase 1 → Foundation ready
2. Add User Story 1 → Test independently → Deploy (MVP! Prev/Next pagination works)
3. Add User Story 2 → Test independently → Deploy (Page number navigation added)
4. Phase 4 → Final verification

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Constitution requires tests (Principle II) — T002 covers pagination logic
- No store changes needed — handler slices the in-memory array (Research R1)
- Page size hardcoded to 20 per spec assumption
- Commit after each phase checkpoint
