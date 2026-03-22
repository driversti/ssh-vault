# Quickstart: Paginated Audit Log

**Branch**: `008-paginated-audit-log`

## What This Feature Does

Adds pagination to the audit log page so that each page shows at most 20 entries. Users navigate between pages using Previous/Next links and page number buttons.

## Files to Create

| File | Purpose |
|------|---------|
| `internal/hub/pagination.go` | `PaginationData` struct, `PageItem` struct, `calcPagination()` helper |
| `internal/hub/pagination_test.go` | Table-driven tests for pagination math |

## Files to Modify

| File | Change |
|------|--------|
| `internal/hub/server.go` | Update `handleAudit()` to parse `page` query param, slice entries, pass `PaginationData` to template |
| `internal/hub/templates/audit.html` | Add pagination controls below the table; show total entry count in header |

## Build & Test

```bash
go build -o ssh-vault ./cmd/ssh-vault    # Build
go test ./internal/hub/...               # Run hub tests
go vet ./...                             # Static analysis
```

## Key Design Decisions

1. **Paginate in handler, not store** — `ListAuditLog()` unchanged; handler slices the reversed array
2. **Pure function for math** — `calcPagination(total, page, pageSize)` returns all template data
3. **Page clamping** — Invalid/out-of-range page numbers silently clamped (no redirects or errors)
4. **Ellipsis for many pages** — Show all page numbers ≤ 7 pages; use `1 ... N-1 N N+1 ... last` pattern otherwise
