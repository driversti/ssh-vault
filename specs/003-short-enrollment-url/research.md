# Research: Short Enrollment URL

**Feature**: 003-short-enrollment-url
**Date**: 2026-03-21

## R1: Short Code Generation Strategy

**Decision**: Use `crypto/rand` to generate a random 6-digit numeric code (100000–999999), checking uniqueness against active codes before accepting.

**Rationale**: The existing token generation already uses `crypto/rand` for 32 random bytes. For 6-digit codes, we generate a random uint32, take modulo 900000, and add 100000 to ensure exactly 6 digits. Uniqueness is enforced by checking against all active (non-expired, non-used) codes in the store before accepting — with a maximum of ~10 concurrent active codes and 900,000 possible values, collisions are astronomically unlikely but must still be guarded against.

**Alternatives considered**:
- Sequential codes (000001, 000002...): Predictable, enables enumeration attacks
- Alphanumeric codes (e.g., `a3f7b2`): Harder to type on mobile, case-sensitivity confusion
- UUID-based short codes: Too long for the "easy to type" goal

## R2: Rate Limiter Design

**Decision**: In-memory per-IP sliding window rate limiter using `sync.Mutex`, `map[string][]time.Time`, and periodic cleanup of stale entries.

**Rationale**: Standard library only (per constitution). A sliding window of timestamps per IP is simple to implement and reason about. Entries older than the window (1 minute) are pruned on each check. A background goroutine or lazy cleanup on access removes IPs that haven't been seen in >5 minutes to prevent memory growth.

**Alternatives considered**:
- Token bucket: More complex, no advantage at this scale
- `golang.org/x/time/rate`: External dependency, violates constitution preference
- Global rate limit: Would block legitimate users during concurrent enrollment attempts from different IPs

## R3: Enrollment Script Architecture

**Decision**: The hub serves a dynamically generated POSIX shell script via `text/template`. The script is parameterized with the hub URL, enrollment token, and GitHub Release download URL pattern. Platform detection happens client-side using `uname -s` and `uname -m`.

**Rationale**: Client-side platform detection (via `uname`) is more reliable than User-Agent parsing — `curl` doesn't send a meaningful User-Agent by default. The script uses `set -euo pipefail` for safety and includes a checksum verification step. The script is generated from a Go `text/template` embedded in the handler, not stored as a separate file.

**Alternatives considered**:
- Static script with code passed as argument: Would require two requests (get script + pass code)
- Server-side platform detection via User-Agent: Unreliable with `curl`, would need custom header
- Separate script file in repository: Adds deployment complexity for a single template

## R4: Short Code ↔ Token Lifecycle

**Decision**: Generating a short code automatically creates and links a new enrollment token (24h TTL). The short code has its own shorter TTL (default 15 min). When the short code is used, the linked token is consumed by the enrollment flow. If the short code expires unused, its linked token is also cleaned up.

**Rationale**: Per clarification, this is a one-step admin workflow. The token is an internal implementation detail — the admin only interacts with short codes. The token's 24h TTL is longer than the short code's 15min TTL, giving the enrollment process time to complete after the short code is consumed. Cleanup of orphaned tokens (where short code expired but token wasn't used) happens during `PurgeExpiredShortCodes`.

**Alternatives considered**:
- Shared TTL (short code and token expire together): The 3-step enrollment may take time if the user doesn't complete immediately
- No linked token (short code IS the token): Would require changing the enrollment API contract, higher blast radius

## R5: Binary Download URL Pattern

**Decision**: The hub knows the GitHub repository owner/name and the latest release tag. The download URL follows the standard Go release pattern: `https://github.com/{owner}/{repo}/releases/download/{tag}/ssh-vault_{os}_{arch}`.

**Rationale**: This is the standard convention for Go CLI releases (used by goreleaser, etc.). The hub stores the GitHub repo path and release tag in its configuration (CLI flags or environment variables). The enrollment script constructs the full URL from OS/arch detection + the base URL provided by the hub in the script template.

**Alternatives considered**:
- Hub proxies binary downloads: Adds bandwidth/storage burden to the hub server
- Hardcoded URL in script: Not configurable, breaks on repo rename or tag change
- GitHub API for latest release: Adds complexity, rate limits, may require auth for private repos

## R6: Dashboard Integration

**Decision**: Extend the existing `tokens.html` template to add a "Quick Enrollment" section above the existing token management. This section has a "Generate Enrollment Link" button that creates a short code and displays the `curl | sh` command with a copy button.

**Rationale**: Keeps the enrollment experience centralized on the tokens page. The admin flow becomes: visit tokens page → click "Generate Enrollment Link" → copy the command → share. The existing token management section remains unchanged for advanced/CLI users. A separate page would fragment the admin experience.

**Alternatives considered**:
- New dedicated page: Adds a nav item for a simple feature, fragmenting the admin UX
- Modal dialog: More complex JS, harder to copy the command
- Separate "enrollment" page combining devices + codes: Too much rework
