# Data Model: SSH Key Sync Hub

**Date**: 2026-03-21
**Source**: [spec.md](spec.md), [research.md](research.md)

## Entities

### Device

Represents a machine enrolled in the mesh.

| Field           | Type      | Description                                          |
| --------------- | --------- | ---------------------------------------------------- |
| ID              | string    | Unique identifier (UUID v4, assigned at enrollment)  |
| Name            | string    | Display name / hostname (user-editable metadata)     |
| PublicKey       | string    | SSH public key in OpenSSH authorized_keys format     |
| Fingerprint     | string    | SSH key fingerprint (SHA256), derived from PublicKey  |
| Status          | enum      | pending / approved / revoked                         |
| APIToken        | string    | Bearer token for sync requests (set after approval)  |
| EnrolledAt      | timestamp | When the device was enrolled (UTC)                   |
| ApprovedAt      | timestamp | When the device was approved (UTC, nullable)         |
| RevokedAt       | timestamp | When the device was revoked (UTC, nullable)          |
| LastSyncAt      | timestamp | Last successful sync check-in (UTC, nullable)        |
| Challenge       | string    | Enrollment challenge (hex, transient — cleared after verification, nullable) |
| Verified        | bool      | Whether the device proved private key ownership during enrollment (default: false) |

**Identity rule**: A device is uniquely identified by its `ID`. The same
public key MAY appear in multiple records (e.g., a revoked record and a
new enrollment of the same physical device).

**Validation rules**:
- `PublicKey` MUST be a valid OpenSSH public key (parseable by
  `ssh.ParseAuthorizedKey()`)
- `Name` MUST be non-empty, max 255 characters
- `APIToken` is set only after approval; retained on revocation for lookup (rejected by status check in auth middleware)

### State Transitions

```text
           enroll
  (none) ────────► pending
                      │
              approve │
                      ▼
                   approved ◄─── (sync updates LastSyncAt)
                      │
              revoke  │
                      ▼
                   revoked (terminal — no transitions out)
```

- **pending → approved**: Owner approves via dashboard. Hub sets
  `ApprovedAt` and generates `APIToken`.
- **approved → revoked**: Owner revokes via dashboard. Hub sets
  `RevokedAt`. `APIToken` is retained (auth middleware rejects by status).
  Device record preserved for audit.
- **revoked → (none)**: No reactivation. Re-enrollment creates a new
  Device record with a new ID (per FR-015).

### Token (Onboarding Token)

Short-lived credential for enrolling a new device.

| Field      | Type      | Description                                       |
| ---------- | --------- | ------------------------------------------------- |
| Value      | string    | Random token string (32-byte hex)                 |
| CreatedAt  | timestamp | When the token was generated (UTC)                |
| ExpiresAt  | timestamp | Expiration time (default: CreatedAt + 24 hours)   |
| UsedBy     | string    | Device ID that consumed this token (nullable)     |
| Used       | bool      | Whether the token has been consumed                |

**Validation rules**:
- A token is valid if: `Used == false` AND `now < ExpiresAt`
- Once used, `UsedBy` is set to the enrolling device's ID and `Used`
  is set to `true`
- Expired tokens are never automatically deleted (retained for audit);
  they can be cleaned up manually

### AuditEntry

Log entry for state-change events (per FR-016).

| Field     | Type      | Description                                        |
| --------- | --------- | -------------------------------------------------- |
| Timestamp | timestamp | When the event occurred (UTC)                      |
| Event     | enum      | enrolled / approved / revoked / auth_failed         |
| DeviceID  | string    | Related device ID (nullable for auth_failed)       |
| Details   | string    | Human-readable description of the event            |

**Storage**: Append-only list within the same `data.json` file.

## Storage Schema

Top-level JSON structure persisted in `data.json`:

```json
{
  "devices": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "macbook-pro",
      "public_key": "ssh-ed25519 AAAA... user@macbook",
      "fingerprint": "SHA256:abc123...",
      "status": "approved",
      "api_token": "a1b2c3d4...",
      "enrolled_at": "2026-03-21T10:00:00Z",
      "approved_at": "2026-03-21T10:05:00Z",
      "revoked_at": null,
      "last_sync_at": "2026-03-21T10:10:00Z"
    }
  ],
  "tokens": [
    {
      "value": "f47ac10b58cc4372...",
      "created_at": "2026-03-21T09:00:00Z",
      "expires_at": "2026-03-22T09:00:00Z",
      "used_by": null,
      "used": false
    }
  ],
  "audit_log": [
    {
      "timestamp": "2026-03-21T10:00:00Z",
      "event": "enrolled",
      "device_id": "550e8400-e29b-41d4-a716-446655440000",
      "details": "Device 'macbook-pro' enrolled via token f47ac..."
    }
  ]
}
```

## Managed Key Block Format

The agent writes this block into `~/.ssh/authorized_keys`:

```text
# BEGIN SSH-VAULT MANAGED BLOCK — DO NOT EDIT
ssh-ed25519 AAAA...key1... device-a
ssh-ed25519 AAAA...key2... device-b
ssh-rsa AAAA...key3... device-c
# END SSH-VAULT MANAGED BLOCK
```

**Rules**:
- Keys within the block are sorted alphabetically by fingerprint for
  deterministic output
- The requesting device's own key is excluded from its own block
  (a device doesn't need to authorize itself)
- The block comment includes "DO NOT EDIT" to warn manual editors
- If the block doesn't exist in the file, it is appended at the end
- If the block exists, only the content between markers is replaced
