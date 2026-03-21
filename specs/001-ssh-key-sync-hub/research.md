# Research: SSH Key Sync Hub

**Date**: 2026-03-21
**Feature**: 001-ssh-key-sync-hub

## R1: Agent-Hub Authentication (SSH Key-based HTTP Auth)

**Decision**: Token-based authentication issued during enrollment.

**Rationale**: Rather than signing every HTTP request with the SSH private
key (which requires implementing a custom HTTP signature scheme), the agent
authenticates once during enrollment by signing a challenge with its SSH
private key. Upon successful verification, the hub issues an opaque API
token tied to the device record. Subsequent sync requests use this token
as a `Bearer` token in the `Authorization` header.

This approach is simpler, well-understood, and avoids the complexity of
per-request SSH signature verification while still proving device identity
at enrollment time.

**Implementation details**:
- Enrollment flow:
  1. Agent sends onboarding token + SSH public key to hub
  2. Hub stores device as "pending" and returns a random challenge
  3. Agent signs the challenge with its SSH private key using
     `golang.org/x/crypto/ssh.Signer`
  4. Agent sends signed challenge back to hub
  5. Hub verifies signature using stored public key via
     `ssh.PublicKey.Verify()`
  6. Hub issues an API token (random 32-byte hex string) stored alongside
     the device record
- Sync requests: `GET /api/keys` with `Authorization: Bearer <api-token>`
- Key package: `golang.org/x/crypto/ssh` — `ssh.ParsePrivateKey()`,
  `ssh.NewPublicKeys()`, `ssh.PublicKey.Verify()`

**Alternatives considered**:
- **Per-request SSH signatures** (HTTP Signatures RFC 9421 style): Too
  complex for a personal tool; no mature Go library; would require
  implementing custom signing/verification middleware.
- **Mutual TLS with SSH-derived certificates**: Conceptually elegant but
  requires a CA and certificate management — overkill.
- **Pre-shared API key (no SSH verification)**: Simpler but doesn't prove
  the agent actually holds the private key corresponding to the registered
  public key. An attacker who obtains the API token could impersonate the
  device without having its private key. Acceptable tradeoff for a personal
  system, but the challenge-response approach is only marginally more complex
  and provides real key verification.

## R2: Atomic File Writes for authorized_keys

**Decision**: Write to temp file in same directory, then `os.Rename()`.

**Rationale**: This is the standard POSIX atomic write pattern. `os.Rename()`
is atomic on the same filesystem. Writing to a temp file in the same
directory as `authorized_keys` guarantees same-filesystem semantics.

**Implementation details**:
- Create temp file with `os.CreateTemp(filepath.Dir(target), ".tmp-*")`
- Set permissions to match original file (default `0600` for authorized_keys)
- Write full content to temp file
- `os.Rename(tempPath, targetPath)` — atomic on POSIX
- If `authorized_keys` doesn't exist yet, create it with `0600` permissions
- If `authorized_keys` is a symlink, resolve it first with
  `filepath.EvalSymlinks()` and operate on the resolved path

**Managed block pattern**:
```text
# BEGIN SSH-VAULT MANAGED BLOCK
ssh-ed25519 AAAA... device-a
ssh-ed25519 AAAA... device-b
# END SSH-VAULT MANAGED BLOCK
```
- Read entire file
- Find start/end markers
- Replace content between markers (or append block if not found)
- Preserve everything outside the markers verbatim
- Sort keys within the block for deterministic output (easier diffing)

**Alternatives considered**:
- **In-place editing with file locking**: Risk of corruption if process
  dies mid-write; `flock()` not portable across all platforms.
- **Using `os.WriteFile()` directly**: Not atomic — partial writes visible
  to other processes (like `sshd`).

## R3: Web Dashboard Approach

**Decision**: Server-rendered HTML with Go `html/template`, embedded via
`//go:embed`, with Pico CSS for minimal styling. No JavaScript framework.

**Rationale**: The constitution mandates simplicity and minimal dependencies.
A server-rendered dashboard with no frontend build step is the simplest
approach. Pico CSS (~10KB) provides a clean, classless/semantic CSS framework
that makes HTML forms and tables look professional without writing custom CSS.
For the small number of interactive actions (approve, revoke, generate token),
standard HTML forms with POST requests and redirects are sufficient.

**Implementation details**:
- `//go:embed templates/*` to embed all HTML templates into the binary
- `html/template` for server-side rendering with template inheritance
  (`layout.html` base + page templates)
- Pico CSS embedded as a single file (`pico.min.css`) in the templates
  directory
- Session auth: generate random session token on login, store in-memory
  (single-user, no persistence needed), set as `HttpOnly`/`Secure` cookie
- Actions (approve/revoke/generate token) as standard HTML form POSTs
  with redirect-after-POST pattern (PRG)
- No JavaScript required for core functionality

**Alternatives considered**:
- **HTMX for partial page updates**: Adds a dependency and complexity for
  marginal UX improvement. With <50 devices, full page reloads are instant.
  Can be added later if needed.
- **React/Vue SPA**: Violates simplicity principle. Requires frontend build
  toolchain, separate dev server, CORS configuration. Completely overkill.
- **Raw HTML without CSS framework**: Functional but ugly. Pico CSS is a
  single file with no build step — minimal cost for significant UX
  improvement.

## R4: Storage Approach

**Decision**: Single JSON file (`data.json`) with file-level mutex locking.

**Rationale**: For a single-user system with <50 devices, a JSON file is the
simplest storage that meets all requirements. The hub is single-process, so
a `sync.RWMutex` in Go provides safe concurrent access from HTTP handlers.
The entire dataset (devices + tokens) will be well under 100KB.

**Implementation details**:
- Single `data.json` file in configurable data directory
- Go struct serialized/deserialized with `encoding/json`
- `sync.RWMutex` for concurrent handler access (read-lock for queries,
  write-lock for mutations)
- Atomic write (same pattern as R2) when persisting changes
- Load into memory at startup; write to disk after every mutation
- Backup: users can copy the file; `data.json` is human-readable

**Data structure**:
```go
type Store struct {
    Devices []Device `json:"devices"`
    Tokens  []Token  `json:"tokens"`
}
```

**Alternatives considered**:
- **SQLite**: More robust but adds CGo dependency (or pure-Go driver which
  is slower). Overkill for <50 records. Violates "prefer standard library."
- **BoltDB/bbolt**: Embedded KV store. Better than SQLite for Go purity but
  still unnecessary complexity for this scale.
- **Multiple JSON files** (one per device): More complex file management;
  atomicity harder to guarantee for operations that touch multiple devices.

## R5: Subcommand Structure

**Decision**: Single binary with `os.Args[1]` subcommand dispatch.

**Rationale**: Go's `flag` package doesn't natively support subcommands well,
but for three simple subcommands (`hub`, `agent`, `enroll`), manual dispatch
on `os.Args[1]` with per-subcommand `flag.FlagSet` is the simplest approach.
No need for `cobra` or `urfave/cli`.

**Implementation**:
```
ssh-vault hub    --addr :8080 --data ./data --password <pass>
ssh-vault agent  --hub-url https://hub:8080 --interval 5m --key ~/.ssh/id_ed25519
ssh-vault enroll --hub-url https://hub:8080 --token <token> --key ~/.ssh/id_ed25519
```

**Alternatives considered**:
- **cobra**: Popular but heavy dependency for 3 subcommands. Violates
  simplicity and "prefer standard library."
- **Separate binaries** (`ssh-vault-hub`, `ssh-vault-agent`): Constitution
  says "one binary" unless explicitly required. Single binary simplifies
  distribution.
