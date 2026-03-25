# Contract: Audit Log Page

**Date**: 2026-03-22 | **Type**: Server-rendered HTML endpoint

## Endpoint

`GET /audit`

**Authentication**: Session cookie (existing — unchanged)

## Query Parameters

| Parameter | Type   | Default | Description                    |
|-----------|--------|---------|--------------------------------|
| page      | int    | 1       | Page number (1-based)          |

### Behavior

- `page` < 1 or non-numeric → treated as page 1
- `page` > total pages → clamped to last page
- Omitted → page 1

## Response

Server-rendered HTML page containing:

1. **Header**: "Audit Log" title with total entry count (e.g., "Audit Log (45 entries)")
2. **Table**: Up to 20 audit entries for the requested page, columns: Time | Event | Device | Details
3. **Pagination controls** (only if total entries > 20):
   - "Previous" link (disabled on page 1)
   - Page number links with ellipsis for large page counts
   - "Next" link (disabled on last page)
   - Current page visually highlighted

## URL Examples

```
/audit           → Page 1 (default)
/audit?page=3    → Page 3
/audit?page=999  → Last available page (clamped)
/audit?page=abc  → Page 1 (invalid input fallback)
```
