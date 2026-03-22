# Research: Paginated Audit Log

**Date**: 2026-03-22 | **Branch**: `008-paginated-audit-log`

## R1: Pagination Strategy — Handler-Level vs Store-Level

**Decision**: Paginate in the handler by slicing the reversed entry list.

**Rationale**: The `FileStore` loads all data into memory from a JSON file. `ListAuditLog()` already returns a full copy of the slice. Adding store-level pagination (offset/limit parameters) would add complexity without any I/O or memory benefit — the entire dataset is already in memory. Slicing a Go slice is O(1) and trivially correct.

**Alternatives considered**:
- **Store-level pagination** (`ListAuditLogPaginated(offset, limit int)`): Rejected because the store loads all data into memory regardless. Would add interface complexity for zero performance gain.
- **Cursor-based pagination** (timestamp-based): Rejected as overkill for an in-memory dataset viewed through server-rendered HTML. Cursor pagination shines for API-driven infinite scroll, not page-number navigation.

## R2: Pagination Helper — Separate File vs Inline

**Decision**: Create `internal/hub/pagination.go` with a `PaginationData` struct and a `calcPagination(totalItems, page, pageSize int)` function.

**Rationale**: Isolating pagination math into a pure function makes it trivially testable with table-driven tests. The handler stays focused on HTTP concerns (parsing query params, calling store, rendering template). If other pages need pagination in the future, the helper is ready — but we're not over-engineering it; it's a single struct and function.

**Alternatives considered**:
- **Inline in handler**: Rejected because mixing math with HTTP handling makes the handler harder to test and read.
- **Separate `internal/pagination` package**: Rejected per constitution (Simplicity — no new packages when a file in the existing package suffices).

## R3: Page Number Controls — Display Strategy

**Decision**: Show all page numbers when total pages ≤ 7. For more pages, use ellipsis truncation: `1 ... (current-1) current (current+1) ... last`.

**Rationale**: The audit log for a small self-hosted hub is unlikely to exceed a few hundred entries (< 50 pages). A simple ellipsis strategy covers the edge case without complex windowing logic. The template can iterate over a pre-computed slice of page numbers (with sentinel values for ellipsis gaps).

**Alternatives considered**:
- **Always show all page numbers**: Works for small counts but degrades for 50+ pages.
- **Only Previous/Next (no page numbers)**: Rejected because the spec explicitly requires page number navigation (FR-003, User Story 2).

## R4: Invalid Page Handling

**Decision**: Clamp the page number — values < 1 become page 1; values > totalPages become the last page. Non-numeric values default to page 1.

**Rationale**: Spec FR-007 says "display the last valid page" for out-of-range. Clamping is the simplest correct behavior and avoids error pages or redirects.

**Alternatives considered**:
- **HTTP redirect to valid page**: Adds a round-trip for no user benefit. The URL still updates via the query param in the rendered links.
- **Return 400 error**: Poor UX for a bookmarked URL that becomes stale as entries are deleted.

## R5: Template Data Structure

**Decision**: Pass a `PaginationData` struct to the template alongside entries:

```
PaginationData {
    CurrentPage  int
    TotalPages   int
    TotalItems   int
    Pages        []PageItem  // pre-computed page numbers with ellipsis markers
    HasPrev      bool
    HasNext      bool
    PrevPage     int
    NextPage     int
}
```

**Rationale**: Pre-computing navigation state in Go keeps the template logic minimal (just `{{range}}` and `{{if}}`). Go templates don't support arithmetic, so computing prev/next page numbers server-side is necessary.

**Alternatives considered**:
- **Passing raw numbers and computing in template**: Go templates lack arithmetic operators; would require adding custom template functions for basic math. Pre-computing is cleaner.
