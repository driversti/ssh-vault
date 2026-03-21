# Feature Specification: Short Enrollment URL

**Feature Branch**: `003-short-enrollment-url`
**Created**: 2026-03-21
**Status**: Draft
**Input**: User description: "As a User I want to simplify the enrollment process. Instead of writing long commands with 64-characters tokens (especially on mobile), I want to use a short command like `curl -sSL https://ssh-vault.yurii.live/e/123456 | sh` where 123456 is a short-living token allowing to pull the script or a command that will install and run the client."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - One-Command Device Enrollment (Priority: P1)

As a user, I want to enroll a new device by running a single short command (`curl -sSL https://hub/e/CODE | sh`) so that I don't have to type long tokens or remember complex CLI arguments, especially on mobile or constrained terminals.

**Why this priority**: This is the core value proposition — reducing enrollment friction from a multi-argument CLI command with a 64-character token to a single memorable URL with a short code.

**Independent Test**: Can be fully tested by generating a short code on the hub dashboard, running the curl command on a target device, and verifying the device appears as pending in the hub.

**Acceptance Scenarios**:

1. **Given** an admin has generated a short enrollment code on the hub dashboard, **When** a user runs `curl -sSL https://hub/e/CODE | sh` on a new device, **Then** the enrollment script is downloaded and executed, completing the enrollment handshake automatically.
2. **Given** a valid short enrollment code exists, **When** a user runs the curl command, **Then** the script detects the device's OS and architecture, downloads the appropriate ssh-vault binary, and initiates enrollment without further user input.
3. **Given** a short enrollment code has already been used, **When** another user tries to use the same code, **Then** the system returns a clear error message indicating the code is no longer valid.
4. **Given** a short enrollment code has expired, **When** a user tries to use it, **Then** the system returns a clear error message indicating the code has expired.

---

### User Story 2 - Short Code Generation (Priority: P1)

As a hub administrator, I want to generate short enrollment codes from the dashboard so that I can share them easily with users who need to enroll devices.

**Why this priority**: Without code generation, the enrollment URL feature cannot function. This is a prerequisite for User Story 1.

**Independent Test**: Can be tested by accessing the hub dashboard, generating a short code, and verifying the code appears in the active codes list with its expiration time.

**Acceptance Scenarios**:

1. **Given** an admin is logged into the hub dashboard, **When** they generate a new enrollment code, **Then** a short numeric code (6 digits) is created along with its underlying enrollment token, with a configurable expiration time.
2. **Given** an admin generates a short code, **When** the code is created, **Then** the dashboard displays the full enrollment URL (e.g., `curl -sSL https://hub/e/123456 | sh`) ready to copy and share.
3. **Given** multiple short codes exist, **When** an admin views the codes list, **Then** each code shows its status (active, used, expired), creation time, and remaining validity period.

---

### User Story 3 - Enrollment Script Delivery (Priority: P2)

As a user running the enrollment command, I want the system to automatically detect my platform and deliver the correct binary and enrollment script so that enrollment works seamlessly on any supported OS.

**Why this priority**: Platform detection enables a truly "just works" experience but can be initially limited to common platforms (Linux amd64/arm64, macOS) without blocking the core feature.

**Independent Test**: Can be tested by running the enrollment script on devices with different OS/architecture combinations (via `uname -s` / `uname -m`) and verifying the correct platform-specific binary is downloaded.

**Acceptance Scenarios**:

1. **Given** a user runs the curl command on a Linux amd64 machine, **When** the hub serves the enrollment script, **Then** the script downloads the Linux amd64 ssh-vault binary.
2. **Given** a user runs the curl command on a macOS arm64 machine, **When** the hub serves the enrollment script, **Then** the script downloads the macOS arm64 ssh-vault binary.
3. **Given** a user runs the curl command on an unsupported platform, **When** the hub serves the enrollment script, **Then** the script displays a clear error message listing supported platforms.

---

### Edge Cases

- What happens when the same short code is requested concurrently from two different devices? First request succeeds; second request receives "code already used" error.
- What happens if the enrollment script download is interrupted midway? The script should verify binary integrity (e.g., checksum) before executing enrollment.
- What happens if the hub is unreachable after the script is downloaded? The script should display a clear connection error and suggest retrying.
- What happens if a short code collides with an existing active code? The system must ensure uniqueness among active (non-expired, non-used) codes before issuing a new one.
- What happens if the device already has ssh-vault installed? The script should detect the existing installation and either use it or offer to update.
- What happens if no SSH key is found in `~/.ssh/`? The script should display a clear error asking the user to generate an SSH key first (e.g., `ssh-keygen`).

## Clarifications

### Session 2026-03-21

- Q: How should the enrollment script determine device name and SSH key? → A: Auto-detect both — use system hostname as device name, auto-discover first available SSH key in `~/.ssh/`.
- Q: Should generating a short code automatically create the underlying enrollment token? → A: Yes, auto-create. Generating a short code automatically creates and links a new enrollment token behind the scenes. The old token-only flow remains available for CLI users.
- Q: Where should the enrollment script download the ssh-vault binary from? → A: GitHub Releases. The hub provides the release tag/URL; the script downloads platform-specific binaries from GitHub.
- Q: Should the `/e/{code}` endpoint enforce rate limiting? → A: Yes, per-IP rate limit (e.g., 10 requests/minute per IP) with a clear error message on limit hit.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST generate short numeric enrollment codes (6 digits) that automatically create and link to a new enrollment token behind the scenes.
- **FR-002**: Short codes MUST expire after a configurable duration (default: 15 minutes).
- **FR-003**: Short codes MUST be single-use — once an enrollment begins with a code, it cannot be reused.
- **FR-004**: System MUST serve an enrollment shell script when the short code URL (`/e/{code}`) is accessed via HTTP GET.
- **FR-005**: The enrollment script MUST detect the target device's operating system and CPU architecture.
- **FR-006**: The enrollment script MUST download the appropriate ssh-vault binary for the detected platform from GitHub Releases.
- **FR-007**: The enrollment script MUST automatically initiate the enrollment process using the token associated with the short code.
- **FR-008**: The enrollment script MUST auto-detect the device name (system hostname) and SSH key (first available public key in `~/.ssh/`) without requiring user input.
- **FR-009**: System MUST return appropriate HTTP error responses for invalid, expired, or already-used codes.
- **FR-010**: The hub dashboard MUST provide a UI element to generate short enrollment codes.
- **FR-011**: The hub dashboard MUST display the full curl command for each active code, ready to copy.
- **FR-012**: System MUST ensure short code uniqueness among all active (non-expired, non-used) codes.
- **FR-013**: The enrollment script MUST verify the integrity of downloaded binaries before execution.
- **FR-014**: System MUST log short code generation and usage in the existing audit trail.
- **FR-015**: The `/e/{code}` endpoint MUST enforce per-IP rate limiting (e.g., 10 requests per minute) and return a clear error when the limit is exceeded.

### Key Entities

- **Short Code**: A 6-digit numeric code that maps to an enrollment token. Has a creation time, expiration duration, status (active/used/expired), and a reference to the underlying enrollment token.
- **Enrollment Script**: A dynamically generated shell script served by the hub that handles binary download, platform detection, and enrollment initiation for a specific short code.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can enroll a new device by typing a single command of fewer than 60 characters.
- **SC-002**: The enrollment process from running the command to device appearing as pending takes less than 30 seconds on a standard connection.
- **SC-003**: Short codes expire automatically and cannot be reused, maintaining the same security posture as the existing token system.
- **SC-004**: The enrollment command works without modification on Linux (amd64, arm64) and macOS (arm64).
- **SC-005**: 100% of enrollment attempts with expired or used codes are rejected with a clear user-facing error message.

## Assumptions

- The hub is accessible over HTTPS at a known public URL (e.g., `ssh-vault.yurii.live`).
- Pre-built ssh-vault binaries for supported platforms are published as GitHub Releases. The hub knows the repository and release tag to construct download URLs.
- The existing 3-step enrollment handshake (enroll, verify, approve) remains unchanged; the short code simplifies only the initiation step.
- The 6-digit code space (1,000,000 combinations) combined with per-IP rate limiting and short TTL makes brute-force attacks infeasible.
- The enrollment script targets POSIX-compatible shells (bash/sh) — Windows/PowerShell support is out of scope for this feature.

## Out of Scope

- QR code generation for enrollment URLs (potential future enhancement).
- Windows/PowerShell enrollment script support.
- Automatic admin approval of devices enrolled via short codes (existing approval flow remains).
- Binary auto-update mechanism for already-enrolled devices.
