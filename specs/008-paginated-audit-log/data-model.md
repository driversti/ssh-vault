# Data Model: Paginated Audit Log

**Date**: 2026-03-22 | **Branch**: `008-paginated-audit-log`

## Existing Entities (No Changes)

### AuditEntry

The audit entry model is **unchanged**. No fields added, removed, or modified.

| Field     | Type      | Description                          |
|-----------|-----------|--------------------------------------|
| Timestamp | time.Time | UTC timestamp of the event           |
| Event     | string    | Event type constant (e.g., "enrolled") |
| DeviceID  | string    | Associated device ID (may be empty)  |
| Details   | string    | Human-readable event description     |

### Store.AuditLog

The `Store.AuditLog` field (slice of `AuditEntry`) and the `FileStore` methods (`AddAuditEntry`, `ListAuditLog`) remain unchanged.

## New Types (View Layer Only)

### PageItem

Represents a single item in the pagination control bar.

| Field    | Type   | Description                                      |
|----------|--------|--------------------------------------------------|
| Number   | int    | Page number (0 for ellipsis sentinel)             |
| IsActive | bool   | Whether this is the current page                  |
| IsGap    | bool   | Whether this represents an ellipsis ("...") gap   |

### PaginationData

Pre-computed pagination metadata passed to the template.

| Field       | Type       | Description                                  |
|-------------|------------|----------------------------------------------|
| CurrentPage | int        | Current page number (1-based)                |
| TotalPages  | int        | Total number of pages                        |
| TotalItems  | int        | Total audit entry count                      |
| Pages       | []PageItem | Ordered list of page items for rendering     |
| HasPrev     | bool       | Whether a previous page exists               |
| HasNext     | bool       | Whether a next page exists                   |
| PrevPage    | int        | Previous page number (0 if none)             |
| NextPage    | int        | Next page number (0 if none)                 |

## State Transitions

No state transitions. `PaginationData` is a computed view-model — it is created fresh on each request from the in-memory audit log slice and is never persisted.

## Relationships

```
AuditEntry[]  --(slice in handler)-->  Page of AuditEntry[] (max 20)
AuditEntry[]  --(count + math)-->      PaginationData
PaginationData + Page entries  ------> Template rendering
```
