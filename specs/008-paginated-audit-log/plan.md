# Implementation Plan: Paginated Audit Log

**Branch**: `008-paginated-audit-log` | **Date**: 2026-03-22 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/008-paginated-audit-log/spec.md`

## Summary

Add server-side pagination to the existing audit log page. The handler will accept a `page` query parameter, slice the in-memory audit entry list into pages of 20, and pass pagination metadata to an updated HTML template that renders page navigation controls.

## Technical Context

**Language/Version**: Go 1.26.1
**Primary Dependencies**: Standard library (`net/http`, `html/template`, `strconv`, `math`), existing `golang.org/x/crypto/ssh`
**Storage**: JSON file via `FileStore` (no changes to persistence layer)
**Testing**: `go test` with table-driven tests
**Target Platform**: Linux server (self-hosted hub)
**Project Type**: Web service (single binary with embedded templates)
**Performance Goals**: Page loads remain instant for 1000+ audit entries (in-memory slicing, no I/O change)
**Constraints**: Standard library only; no third-party pagination libraries
**Scale/Scope**: Single handler + template change; ~50-80 lines of new/modified Go code

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Idiomatic Go | PASS | Uses standard library, explicit error handling, `gofmt` |
| II. Testing | PASS | Table-driven tests for pagination math |
| III. Simplicity | PASS | In-memory slicing of existing data; no new abstractions, packages, or store methods |

**Technology Stack**: PASS — Go standard library only, no new dependencies
**Development Workflow**: PASS — Feature branch, tests required

All gates pass. No violations to justify.

## Project Structure

### Documentation (this feature)

```text
specs/008-paginated-audit-log/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
└── tasks.md             # Phase 2 output (via /speckit.tasks)
```

### Source Code (files to modify)

```text
internal/hub/
├── server.go            # Modify handleAudit() — add page param parsing, slicing, pagination metadata
├── pagination.go        # New file — pagination helper (calcPage, PaginationData struct)
├── pagination_test.go   # New file — table-driven tests for pagination logic
└── templates/
    └── audit.html       # Modify — add pagination controls below the table
```

**Structure Decision**: All changes stay within `internal/hub/` — no new packages. A small `pagination.go` file isolates reusable pagination math from the handler, keeping `server.go` focused on HTTP concerns. This follows the existing pattern where `server.go` delegates to helpers.
