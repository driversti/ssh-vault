# Data Model: Device Rename

## Existing Entities (no schema changes)

### Device

The `Device` struct already has a `Name` field with validation (non-empty, max 255 chars) in `Validate()`. No schema changes needed — rename is a simple field update on an existing entity.

| Field | Type | Constraints | Changed? |
|-------|------|-------------|----------|
| ID | string | UUID format, immutable | No |
| Name | string | Non-empty, max 255 chars | **Updated by rename** |
| PublicKey | string | OpenSSH format | No |
| Fingerprint | string | SHA256 hash | No |
| Status | string | pending/approved/revoked | No |
| ... | ... | ... | No |

### AuditEntry

No schema changes. A new event constant is added:

| Constant | Value | Description |
|----------|-------|-------------|
| `EventDeviceRenamed` | `"device_renamed"` | **New** — logged when a device name is changed |

**Detail format**: `"Device renamed from 'old-name' to 'new-name'"`

## State Transitions

Rename does **not** affect device status. It is an orthogonal operation:

```
Any status (pending, approved, revoked) → Name updated → Same status
```

## Validation Rules

1. Trim leading/trailing whitespace from new name
2. New name must be non-empty after trimming
3. New name must not exceed 255 characters
4. New name must differ from current name (otherwise no-op)
