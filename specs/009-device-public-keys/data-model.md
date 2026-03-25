# Data Model: Device Public Keys Display

**Feature**: 009-device-public-keys | **Date**: 2026-03-22

## Existing Entities (No Changes)

### Device

The `Device` struct in `internal/model/device.go` already contains all required fields:

| Field | Type | Description |
|-------|------|-------------|
| ID | string | Unique device identifier |
| Name | string | User-friendly device name |
| PublicKey | string | Full SSH public key (OpenSSH format) |
| Fingerprint | string | SHA256 fingerprint |
| Status | string | "pending", "approved", or "revoked" |
| ... | ... | Other fields unchanged |

**No schema changes required.** All new display data is derived at render time.

## Derived Display Data (Template Functions)

These are computed from `Device.PublicKey` at template render time — not stored:

| Derived Field | Source | Logic |
|---------------|--------|-------|
| Key Type | `PublicKey` | Parsed via `ssh.ParseAuthorizedKey` → `Key.Type()` (e.g., "ssh-ed25519" → "ed25519") |
| Username | `PublicKey` comment | Split comment on `@` → first part. Empty if no comment or no `@`. |
| Host/IP | `PublicKey` comment | Split comment on `@` → second part. Empty if no comment or no `@`. |
| Full Key | `PublicKey` | Raw string, displayed as-is in expand view |

## State Transitions

No new state transitions. The `Device` lifecycle (pending → approved → revoked) is unchanged.

## Validation Rules

No new validation rules. The existing `Device.Validate()` already ensures `PublicKey` is non-empty.
