# Tasks: App Logo Integration

**Input**: Design documents from `/specs/005-app-logo/`
**Prerequisites**: plan.md (required), spec.md (required for user stories)

**Tests**: No tests explicitly requested. The handler follows an existing proven pattern (`handleStaticCSS`), so manual browser verification per quickstart.md is sufficient.

**Organization**: Both user stories share the same foundational asset (logo.svg), so they are grouped sequentially.

## Phase 1: Setup

**Purpose**: Place the logo asset in the correct location

- [x] T001 [US1,US2] Copy `logo.svg` from repository root to `internal/hub/templates/logo.svg`

**Checkpoint**: Logo file exists in the embedded templates directory

---

## Phase 2: User Story 1 - Logo in Browser Tab (Priority: P1)

**Goal**: Display logo.svg as the browser favicon on all pages

**Independent Test**: Open any page and verify the logo appears in the browser tab

### Implementation

- [x] T002 [US1] Add `handleStaticLogo` handler in `internal/hub/server.go` (mirrors `handleStaticCSS`)
- [x] T003 [US1] Register `/static/logo.svg` route in `registerRoutes()` in `internal/hub/server.go`
- [x] T004 [US1] Add `<link rel="icon" type="image/svg+xml" href="/static/logo.svg">` to `<head>` in `internal/hub/templates/layout.html`

**Checkpoint**: Favicon visible in browser tab on all pages

---

## Phase 3: User Story 2 - Logo in Header (Priority: P1)

**Goal**: Display logo in the left corner of the header navigation bar

**Independent Test**: Open any page and verify logo appears next to "SSH Vault" text in the header

### Implementation

- [x] T005 [US2] Add `<img src="/static/logo.svg">` with appropriate sizing before "SSH Vault" text in the `<nav>` section of `internal/hub/templates/layout.html`

**Checkpoint**: Logo visible in header on all pages, layout balanced

---

## Phase 4: Polish

- [x] T006 Run `go build ./...` and `go vet ./...` to verify no compilation errors
- [x] T007 Run quickstart.md validation steps

---

## Dependencies & Execution Order

- **T001**: No dependencies — must complete first (shared asset)
- **T002, T003**: Depend on T001 — sequential (same file: server.go)
- **T004**: Depends on T001 — can run in parallel with T002/T003 (different file: layout.html)
- **T005**: Depends on T001 — can run in parallel with T002/T003 (different file: layout.html), but sequential with T004 (same file)
- **T006, T007**: Depend on all previous tasks
