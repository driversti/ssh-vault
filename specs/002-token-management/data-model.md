# Data Model: Token Management Enhancements

**Feature**: 002-token-management | **Date**: 2026-03-21

## Entities

### Token (existing — no schema changes)

```
Token {
  Value     string     // 64-char hex, unique identifier, immutable after creation
  CreatedAt time.Time  // UTC, set on creation
  ExpiresAt time.Time  // UTC, set on creation (CreatedAt + 24h)
  UsedBy    string     // Device ID that consumed the token (empty if unused)
  Used      bool       // true once consumed during enrollment
}
```

**State transitions**:
```
  [created] ──use──► [used]     (enrollment consumes token; sets Used=true, UsedBy=deviceID)
  [created] ──remove──► [deleted]  (admin removes; hard delete from storage)
  [created] ──expire──► [expired]  (time passes ExpiresAt; auto-purged on next token list access)
```

**Validation rules**:
- Only tokens where `Used == false` can be removed (FR-002)
- Token value is used as the unique identifier for removal (no separate ID field needed)
- Expired and used tokens are purged during `PurgeExpiredTokens()` — they serve no further purpose

### AuditEntry (existing — new event type constants only)

```
AuditEntry {
  Timestamp time.Time  // UTC, set on creation
  Event     string     // Event type constant
  DeviceID  string     // Associated device (empty for admin-only actions)
  Details   string     // Human-readable context
}
```

**New event type constants**:

| Constant          | Value            | Trigger                          | DeviceID        | Details format                                         |
|-------------------|------------------|----------------------------------|-----------------|--------------------------------------------------------|
| EventTokenUsed    | "token_used"     | Device enrolls with a token      | enrolling device | "Token {prefix}... used by device '{name}'"            |
| EventTokenRemoved | "token_removed"  | Admin removes an unused token    | empty string     | "Token {prefix}... removed"                            |

**Existing event types** (unchanged):

| Constant        | Value         | Trigger                    |
|-----------------|---------------|----------------------------|
| EventEnrolled   | "enrolled"    | Device enrollment started  |
| EventApproved   | "approved"    | Admin approves device      |
| EventRevoked    | "revoked"     | Admin revokes device       |
| EventAuthFailed | "auth_failed" | Auth verification failure  |

### Store (existing — no schema changes)

```
Store {
  Devices  []Device      // All registered devices
  Tokens   []Token       // All tokens (active, used, expired — until purged)
  AuditLog []AuditEntry  // Append-only event log
}
```

**New store operations**:

| Operation             | Behavior                                                              |
|-----------------------|-----------------------------------------------------------------------|
| RemoveToken(value)    | Find token by value, verify unused, delete from slice, persist        |
| PurgeExpiredTokens()  | Remove all expired tokens from slice, persist if any removed, return count |

## Relationships

```
Token ──used_by──► Device (via UsedBy field, set during enrollment)
AuditEntry ──references──► Device (via DeviceID field)
AuditEntry ──references──► Token (via truncated prefix in Details string, not a foreign key)
```

## Storage Impact

- **No schema migration needed**: The JSON structure (`model.Store`) is unchanged. New constants are code-only additions.
- **Data reduction**: `PurgeExpiredTokens()` actively shrinks the token array, reducing file size over time.
- **Backward compatibility**: Existing data files work without modification. New event types appear only in new audit entries.
