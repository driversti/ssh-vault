# SSH Vault

Centralized SSH key distribution for your personal devices. One binary, zero dependencies, no database.

SSH Vault solves a simple problem: when you add a new laptop, server, or Raspberry Pi to your collection, you shouldn't have to manually copy SSH keys to every other device. Enroll a device once, approve it from the web dashboard, and its public key is automatically distributed to all your other devices within minutes.

```
┌──────────────┐     enroll     ┌──────────┐     sync       ┌─────────────┐
│  New Device  │───────────────▶│   Hub    │◀───────────────│  Device A   │
│  (agent)     │  token + key   │ :8080    │  GET /api/keys │  (agent)    │
└──────────────┘                │          │                └─────────────┘
                                │ dashboard│◀──────────────┐
                   approve ───▶ │ + API    │               │
                                └──────────┘     sync      │
                                                           │
                                                    ┌─────────────┐
                                                    │  Device B   │
                                                    │  (agent)    │
                                                    └─────────────┘
```

## Features

- **Single binary** — `ssh-vault` ships as one executable with three subcommands: `hub`, `agent`, `enroll`
- **Challenge-response enrollment** — devices prove private key ownership during enrollment via SSH signature verification
- **Web dashboard** — view devices, generate onboarding tokens, approve/revoke with a clean Pico CSS interface
- **Managed key block** — agent writes keys between `BEGIN/END SSH-VAULT MANAGED BLOCK` markers in `authorized_keys`, preserving your manual entries
- **Atomic file writes** — `authorized_keys` is never half-written; uses temp file + rename
- **Offline resilience** — if the hub goes down, agents retain the last synced keys and resume when it returns
- **Revocation propagation** — revoke a device from the dashboard and its key is removed from all devices within one sync cycle
- **Audit log** — enrollments, approvals, revocations, and failed auth attempts are logged
- **Optional TLS** — pass `--tls-cert` and `--tls-key`, or run behind a reverse proxy
- **No external dependencies** — standard library + `golang.org/x/crypto/ssh`, file-based JSON storage

## Quick Start

### Build

```bash
go build -o ssh-vault ./cmd/ssh-vault
```

### 1. Start the Hub

On your always-on server (home server, VPS, etc.):

```bash
export VAULT_PASSWORD="your-secret-passphrase"
./ssh-vault hub --addr :8080 --data ./data
```

### 2. Generate a Token

Open `http://your-hub:8080` in a browser, log in with your passphrase, go to **Tokens**, click **Generate Token**, and copy it.

### 3. Enroll a Device

On the device you want to add:

```bash
./ssh-vault enroll \
  --hub-url http://your-hub:8080 \
  --token <paste-token> \
  --key ~/.ssh/id_ed25519
```

### 4. Approve

Back on the dashboard, click **Approve** next to the new device.

### 5. Start the Agent

```bash
./ssh-vault agent \
  --hub-url http://your-hub:8080 \
  --interval 5m \
  --key ~/.ssh/id_ed25519
```

The agent syncs approved keys every 5 minutes into a managed block in `~/.ssh/authorized_keys`.

### 6. Verify

From another enrolled device:

```bash
ssh user@new-device  # no password prompt — key was distributed automatically
```

## CLI Reference

### `ssh-vault hub`

Start the hub server (dashboard + API).

| Flag | Default | Description |
|------|---------|-------------|
| `--addr` | `:8080` | Listen address |
| `--data` | `./data` | Data directory for `data.json` |
| `--password` | — | Dashboard passphrase (or `VAULT_PASSWORD` env) |
| `--tls-cert` | — | TLS certificate file (optional) |
| `--tls-key` | — | TLS private key file (optional) |

### `ssh-vault enroll`

Enroll this device with a hub.

| Flag | Default | Description |
|------|---------|-------------|
| `--hub-url` | — | Hub base URL (required) |
| `--token` | — | Onboarding token (required) |
| `--key` | `~/.ssh/id_ed25519` | SSH private key path |
| `--name` | hostname | Device display name |

### `ssh-vault agent`

Start the sync agent.

| Flag | Default | Description |
|------|---------|-------------|
| `--hub-url` | — | Hub base URL |
| `--interval` | `5m` | Sync interval |
| `--key` | `~/.ssh/id_ed25519` | SSH private key path |
| `--auth-keys` | `~/.ssh/authorized_keys` | authorized_keys file path |

## How It Works

### Enrollment Flow

1. Hub generates a single-use onboarding token (valid 24h)
2. Agent sends the token + SSH public key to `POST /api/enroll`
3. Hub returns a random challenge
4. Agent signs the challenge with its SSH private key
5. Agent sends the signature to `POST /api/enroll/verify`
6. Hub verifies the signature, marks device as "pending"
7. Owner approves via dashboard → hub generates an API bearer token
8. Agent uses the bearer token for all subsequent sync requests

### Sync Loop

Every interval (default 5 minutes), the agent:

1. Calls `GET /api/keys` with its bearer token
2. Receives the list of all approved devices' public keys (excluding its own)
3. Writes them into the managed block in `authorized_keys`
4. Keys outside the managed block are never touched

### Managed Block

```
# existing manual keys are preserved
ssh-rsa AAAA... admin@jumpbox

# BEGIN SSH-VAULT MANAGED BLOCK — DO NOT EDIT
ssh-ed25519 AAAA... laptop
ssh-ed25519 BBBB... desktop
ssh-ed25519 CCCC... raspberry-pi
# END SSH-VAULT MANAGED BLOCK
```

### Revocation

Click **Revoke** on the dashboard. The revoked device:
- Gets a `401 "device revoked"` on its next sync and stops
- Is excluded from other devices' key lists on their next sync
- Record is preserved in the audit log

## Architecture

```
cmd/ssh-vault/         # Single binary entry point
internal/
├── hub/               # Hub server — HTTP handlers, auth, storage, templates
├── agent/             # Sync agent — enrollment, config, sync loop
├── keyblock/          # authorized_keys file manipulation (atomic writes)
└── model/             # Shared types — Device, Token, AuditEntry
```

- **Storage**: Single `data.json` file (human-readable, easy to back up)
- **Dashboard**: Server-rendered HTML with [Pico CSS](https://picocss.com), embedded in the binary via `//go:embed`
- **Auth**: Bearer tokens for agents, password sessions for dashboard
- **Concurrency**: `sync.RWMutex` protects the in-memory store

## Design Decisions

| Decision | Rationale |
|----------|-----------|
| Single binary | Simplest distribution — copy one file |
| File-based JSON storage | No database dependency; sufficient for <50 devices |
| `golang.org/x/crypto/ssh` only | Quasi-stdlib; no third-party dependencies |
| Challenge-response enrollment | Proves the agent holds the private key, not just a copy of the public key |
| Bearer tokens for sync | Simpler than per-request SSH signatures; identity verified once at enrollment |
| Server-rendered HTML | No frontend build step; the dashboard is embedded in the binary |
| Managed block pattern | Coexists with manually managed keys; `sshd` sees one coherent file |

## Security Considerations

- **TLS**: Use `--tls-cert`/`--tls-key` or run behind a reverse proxy. Without TLS, tokens and keys traverse the network in cleartext.
- **Password**: The dashboard password is compared with constant-time comparison (`crypto/subtle`). Set a strong passphrase.
- **Tokens**: Onboarding tokens are 32 bytes of `crypto/rand`, single-use, 24h expiry.
- **File permissions**: `authorized_keys` is written with `0600` permissions via atomic rename.
- **Single-user system**: Designed for one owner managing their personal devices. Not suited for multi-tenant or team use.

## Development

```bash
go build -o ssh-vault ./cmd/ssh-vault   # Build
go test ./...                            # Run tests (42 test cases)
go vet ./...                             # Static analysis
```

## Requirements

- Go 1.22+ (uses `log/slog`, `embed`)
- Linux or macOS (hub and agents)
- SSH key pair on each device (e.g., `ssh-keygen -t ed25519`)

## License

MIT
