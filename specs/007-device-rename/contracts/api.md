# API Contract: Device Rename

## POST /devices/{id}/rename

Renames a device. Requires session authentication (dashboard cookie).

### Request

**Content-Type**: `application/json`

```json
{
  "name": "my-new-device-name"
}
```

| Field | Type | Required | Constraints |
|-------|------|----------|-------------|
| name | string | Yes | Non-empty after trimming, max 255 characters |

### Response

**Success (200 OK)**:

```json
{
  "name": "my-new-device-name"
}
```

Returns the trimmed, saved name.

**Validation Error (400 Bad Request)**:

```json
{
  "error": "device name must not be empty"
}
```

Possible error messages:
- `"device name must not be empty"` — name is empty or whitespace-only after trimming
- `"device name must not exceed 255 characters, got N"` — name too long
- `"invalid request body"` — malformed JSON
- `"invalid device ID"` — missing or malformed ID in path

**Not Found (404 Not Found)**:

```json
{
  "error": "device not found"
}
```

**No Change (200 OK, no-op)**:

If the trimmed name is identical to the current name, returns 200 with the current name. No audit entry is created.

**Server Error (500 Internal Server Error)**:

```json
{
  "error": "internal error"
}
```

### Side Effects

- On successful rename (name actually changed): creates an `AuditEntry` with event `"device_renamed"` and details `"Device renamed from 'old' to 'new'"`.
- No audit entry for validation failures, no-ops, or server errors.

### Authentication

Requires a valid session cookie (same as all `/devices/*` dashboard routes). Unauthenticated requests are redirected to `/login`.
