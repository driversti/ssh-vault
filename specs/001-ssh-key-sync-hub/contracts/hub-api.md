# Hub API Contract

**Date**: 2026-03-21
**Protocol**: HTTP/HTTPS
**Base path**: Hub address (e.g., `https://hub.local:8080`)

## Authentication

### Dashboard Auth (Human)

Password-based session authentication.

- **Login**: `POST /login` with form data `password=<passphrase>`
- **Session**: `Set-Cookie: session=<random-token>; HttpOnly; Secure; SameSite=Strict`
- **Logout**: `POST /logout` (clears session cookie)
- All dashboard routes require valid session cookie; redirect to `/login` if missing/invalid

### Agent Auth (Machine)

Bearer token authentication for API endpoints.

- Header: `Authorization: Bearer <api-token>`
- The `api-token` is issued during enrollment approval and stored in the device record
- Invalid/revoked tokens receive `401 Unauthorized`

---

## Dashboard Endpoints (Session Auth)

### GET /

Dashboard home — device list view.

**Response**: HTML page showing all devices with status, name, fingerprint,
last sync time. Stale devices (no sync for >3x interval) visually flagged.

### GET /login

Login form.

### POST /login

Authenticate with password.

| Field    | Type   | Required | Description             |
| -------- | ------ | -------- | ----------------------- |
| password | string | yes      | Hub passphrase          |

**Success**: 303 redirect to `/`
**Failure**: 303 redirect to `/login?error=1`

### POST /logout

Clear session.

**Response**: 303 redirect to `/login`

### POST /devices/{id}/approve

Approve a pending device.

**Response**: 303 redirect to `/`
**Errors**: 404 if device not found; 400 if device not in "pending" status

### POST /devices/{id}/revoke

Revoke an approved device.

**Response**: 303 redirect to `/`
**Errors**: 404 if device not found; 400 if device not in "approved" status

### GET /tokens

Token management page — list active tokens, generate new ones.

### POST /tokens

Generate a new onboarding token.

**Response**: 303 redirect to `/tokens` (new token displayed on page)

---

## Agent API Endpoints

### POST /api/enroll

Begin device enrollment.

**Auth**: Onboarding token (not session or bearer).

**Request body** (JSON):
```json
{
  "token": "f47ac10b58cc4372...",
  "public_key": "ssh-ed25519 AAAA... user@device",
  "name": "my-laptop"
}
```

**Success response** (200):
```json
{
  "device_id": "550e8400-...",
  "challenge": "random-challenge-bytes-hex"
}
```

**Errors**:
- `400`: Invalid request body
- `401`: Token invalid, expired, or already used
- `409`: Public key already registered to an approved device

### POST /api/enroll/verify

Complete enrollment by proving private key ownership.

**Request body** (JSON):
```json
{
  "device_id": "550e8400-...",
  "signature": "base64-encoded-signature-of-challenge"
}
```

**Success response** (200):
```json
{
  "status": "pending",
  "message": "Device registered. Awaiting approval."
}
```

**Errors**:
- `400`: Invalid request body or device ID
- `401`: Signature verification failed

### GET /api/keys

Get the current list of approved device public keys.

**Auth**: Bearer token (agent's API token).

**Success response** (200):
```json
{
  "keys": [
    "ssh-ed25519 AAAA...key1... device-a",
    "ssh-ed25519 AAAA...key2... device-b"
  ],
  "updated_at": "2026-03-21T10:05:00Z"
}
```

**Notes**:
- The requesting device's own key is excluded from the list
- Keys are sorted by fingerprint for deterministic ordering

**Errors**:
- `401`: Invalid or revoked API token

---

## CLI Contract

### `ssh-vault hub`

Start the hub server.

| Flag         | Default        | Description                           |
| ------------ | -------------- | ------------------------------------- |
| `--addr`     | `:8080`        | Listen address                        |
| `--data`     | `./data`       | Data directory (for data.json)        |
| `--password` | (required)     | Dashboard passphrase (or env `VAULT_PASSWORD`) |
| `--tls-cert` | (optional)     | Path to TLS certificate file                  |
| `--tls-key`  | (optional)     | Path to TLS private key file                  |

### `ssh-vault agent`

Start the sync agent.

| Flag         | Default                  | Description                     |
| ------------ | ------------------------ | ------------------------------- |
| `--hub-url`  | (required)               | Hub base URL                    |
| `--interval` | `5m`                     | Sync interval                   |
| `--key`      | `~/.ssh/id_ed25519`      | Path to SSH private key         |
| `--auth-keys`| `~/.ssh/authorized_keys` | Path to authorized_keys file    |

### `ssh-vault enroll`

Enroll this device with the hub.

| Flag         | Default                  | Description                     |
| ------------ | ------------------------ | ------------------------------- |
| `--hub-url`  | (required)               | Hub base URL                    |
| `--token`    | (required)               | Onboarding token                |
| `--key`      | `~/.ssh/id_ed25519`      | Path to SSH private key         |
| `--name`     | (hostname)               | Device display name             |
