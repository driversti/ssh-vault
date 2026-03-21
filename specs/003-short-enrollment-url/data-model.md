# Data Model: Short Enrollment URL

**Feature**: 003-short-enrollment-url
**Date**: 2026-03-21

## New Entity: ShortCode

| Field | Type | Description |
|-------|------|-------------|
| Code | string | 6-digit numeric code (e.g., "482917") — unique among active codes |
| TokenValue | string | Reference to the linked enrollment Token.Value (64-char hex) |
| CreatedAt | time.Time | UTC timestamp of creation |
| ExpiresAt | time.Time | UTC timestamp of expiration (default: CreatedAt + 15 minutes) |
| Used | bool | Set to true when the code is consumed by an enrollment request |
| UsedAt | *time.Time | UTC timestamp when used (nil if unused) |
| UsedByIP | string | IP address that consumed the code (empty if unused) |

### Validation Rules

- `Code` must be exactly 6 digits (100000–999999)
- `Code` must be unique among all active (non-expired, non-used) codes at creation time
- `TokenValue` must reference a valid, unused Token in the store
- `ExpiresAt` must be after `CreatedAt`

### State Transitions

```
[new] → active (Used=false, not expired)
      → used (Used=true, UsedAt set, linked token consumed by enrollment)
      → expired (time.Now() >= ExpiresAt, Used=false — linked token also cleaned up)
```

### Methods

- `NewShortCode(code string, tokenValue string, ttl time.Duration) *ShortCode`
- `IsValid() bool` — `!Used && !IsExpired()`
- `IsExpired() bool` — `!time.Now().UTC().Before(ExpiresAt)`
- `MarkUsed(ip string)` — sets `Used=true`, `UsedAt=now`, `UsedByIP=ip`

## Modified Entity: Store (model/store.go)

Add `ShortCodes []ShortCode` to the existing `Store` struct:

```
Store {
    Devices    []Device      (existing)
    Tokens     []Token       (existing)
    ShortCodes []ShortCode   (NEW)
    AuditLog   []AuditEntry  (existing)
}
```

## Modified Entity: AuditEntry (model/audit.go)

New event constants:

| Event | Details Format |
|-------|---------------|
| `shortcode_created` | `"Short code {code} created (expires {ExpiresAt})"` |
| `shortcode_used` | `"Short code {code} used from {IP}"` |
| `shortcode_expired` | `"Short code {code} expired (token cleaned up)"` |

## New Component: Rate Limiter

In-memory structure (not persisted):

| Field | Type | Description |
|-------|------|-------------|
| mu | sync.Mutex | Protects concurrent access |
| requests | map[string][]time.Time | IP → list of request timestamps within the window |
| window | time.Duration | Sliding window size (1 minute) |
| limit | int | Max requests per window per IP (10) |

### Cleanup

Stale IPs (no requests in >5 minutes) are removed lazily on each `Allow(ip)` check or periodically via a background goroutine.

## Relationships

```
ShortCode.TokenValue ──references──▶ Token.Value
ShortCode ──creates──▶ Token (on short code generation)
ShortCode ──uses──▶ Token (on enrollment, token marked as used)
ShortCode ──cleanup──▶ Token (on short code expiry, orphaned token removed)
```

## FileStore Operations (hub/store.go)

New methods following existing patterns:

| Method | Lock | Description |
|--------|------|-------------|
| `AddShortCode(sc ShortCode) error` | Write | Append to ShortCodes slice, save |
| `GetShortCode(code string) (*ShortCode, error)` | Read | Find by Code field |
| `UseShortCode(code string, ip string) (*ShortCode, error)` | Write | Validate, mark used, save |
| `ListShortCodes() []ShortCode` | Read | Return all short codes |
| `PurgeExpiredShortCodes() int` | Write | Remove expired+unused codes, clean up orphaned tokens |
