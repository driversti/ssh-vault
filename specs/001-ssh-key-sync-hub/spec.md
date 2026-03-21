# Feature Specification: SSH Key Sync Hub

**Feature Branch**: `001-ssh-key-sync-hub`
**Created**: 2026-03-21
**Status**: Draft
**Input**: User description: "Centralized management hub and automated syncing agent for SSH key distribution across personal devices"

## Clarifications

### Session 2026-03-21

- Q: How does the owner authenticate to the web dashboard? → A: Password/passphrase set at hub startup via config or environment variable.
- Q: Does the agent use the device's existing SSH key or a separate identity key? → A: Same key — the device's existing SSH public key is used for both hub authentication and distribution to other devices' authorized_keys.
- Q: When re-enrolling a previously revoked device, new record or reactivate? → A: New record — the revoked entry is preserved for audit; re-enrollment creates a fresh device entry.
- Q: Should the hub maintain a centralized audit log? → A: State-change log — log enrollments, approvals, revocations, and failed authentication attempts (not routine sync heartbeats).

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Enroll a New Device (Priority: P1)

As a device owner, I want to enroll a new device into my personal mesh so that it can immediately connect to—and be reachable from—all my other enrolled devices without manually copying SSH keys.

**Why this priority**: This is the foundational action. Without device enrollment, no other feature has value. A single enrolled device already proves the system works end-to-end.

**Independent Test**: Can be fully tested by installing the agent on a fresh machine, generating an onboarding token on the hub, running the enrollment command on the new device, approving it, and verifying the new device's public key appears in every other enrolled device's authorized_keys file within one sync cycle.

**Acceptance Scenarios**:

1. **Given** a running hub with zero enrolled devices, **When** I install the agent on a laptop and enroll it using a valid onboarding token, **Then** the hub shows the device as "pending approval" with its hostname, public key fingerprint, and enrollment timestamp.
2. **Given** a pending device on the hub dashboard, **When** I approve it, **Then** the device status changes to "approved" and its public key is included in the next sync distribution to all other approved devices.
3. **Given** two approved devices (A and B), **When** the agent on device B completes its next scheduled sync, **Then** device A's public key appears inside the managed block of device B's authorized_keys file, and vice versa.
4. **Given** an onboarding token, **When** it has already been used by one device, **Then** a second device attempting to use the same token is rejected.

---

### User Story 2 - Revoke a Compromised Device (Priority: P2)

As a device owner, I want to instantly revoke a lost or stolen device from the hub so that the compromised device's key is removed from every other device in my mesh within one sync cycle.

**Why this priority**: This is the primary security control. Without revocation, the system introduces risk by centralizing trust. Even with only enrollment working (US1), revocation makes the system safe to rely on.

**Independent Test**: Can be tested by enrolling three devices, revoking one from the dashboard, waiting for a sync cycle, and verifying the revoked device's key no longer appears in any remaining device's authorized_keys file.

**Acceptance Scenarios**:

1. **Given** three approved devices (A, B, C), **When** I press "Revoke" on device C from the hub dashboard, **Then** device C's status changes to "revoked" immediately on the dashboard.
2. **Given** device C has been revoked, **When** devices A and B complete their next sync, **Then** device C's public key is removed from the managed block of both A's and B's authorized_keys files.
3. **Given** a revoked device C, **When** device C's agent attempts to sync with the hub, **Then** the hub rejects the request and the agent logs the rejection locally.
4. **Given** a revoked device, **When** I view the dashboard, **Then** the device remains visible in a "revoked" state for audit purposes and is not re-enrollable without a new onboarding token.
5. **Given** a revoked device that is re-enrolled with a new token, **When** the enrollment is approved, **Then** a new device record is created (the old revoked record is preserved for audit) and the device receives a fresh unique ID.

---

### User Story 3 - Manage Devices via Web Dashboard (Priority: P3)

As a device owner, I want a private web dashboard to view all my devices, their sync status, and generate onboarding tokens so that I have full visibility and control over my device mesh from one place.

**Why this priority**: The dashboard is the management interface. While enrollment and revocation could theoretically work via a CLI, the dashboard provides the intended user experience and makes the system practical for daily use.

**Independent Test**: Can be tested by accessing the dashboard, verifying it displays accurate device information (name, status, last sync time, key fingerprint), generating a new onboarding token, and confirming it appears as a copiable value.

**Acceptance Scenarios**:

1. **Given** a running hub, **When** I access the dashboard in a browser, **Then** I must authenticate before seeing any data.
2. **Given** an authenticated session, **When** I view the device list, **Then** I see each device's name, status (pending/approved/revoked), public key fingerprint, and last successful sync timestamp.
3. **Given** an authenticated session, **When** I generate an onboarding token, **Then** the dashboard displays a single-use token that I can copy, along with its expiration time.
4. **Given** a device that has not synced for longer than 3x the expected sync interval, **When** I view the dashboard, **Then** the device is visually flagged as "stale" to alert me of potential connectivity issues.

---

### User Story 4 - Offline Resilience (Priority: P4)

As a device owner, I want my devices to retain the last known authorized_keys state when the hub is unreachable so that device-to-device SSH connections continue working during hub outages.

**Why this priority**: This is a resilience guarantee. It prevents the system from becoming a single point of failure. It is lower priority because the default behavior of "do nothing if the hub is down" already partially achieves this—the spec formalizes it as a guarantee.

**Independent Test**: Can be tested by enrolling two devices, confirming they can SSH to each other, shutting down the hub, waiting several sync intervals, and verifying the devices can still SSH to each other with the previously synced keys.

**Acceptance Scenarios**:

1. **Given** two approved devices with previously synced keys, **When** the hub becomes unreachable, **Then** the agents log the connectivity failure but do NOT modify the authorized_keys file.
2. **Given** the hub has been unreachable for multiple sync cycles, **When** the hub comes back online, **Then** the agents resume normal sync behavior on the next scheduled check-in and apply any pending changes.
3. **Given** a device that has never successfully synced (freshly enrolled, hub went down before first sync), **When** the hub is unreachable, **Then** the agent retries on each scheduled interval and logs that no initial sync has been completed.

---

### Edge Cases

- What happens when two devices are enrolled simultaneously with different tokens? Both MUST appear as independent pending approvals.
- What happens when the agent process is killed mid-write of the authorized_keys file? The agent MUST use atomic file operations (write to temp file, then rename) to prevent corruption.
- What happens when the authorized_keys file has been manually edited outside the managed block? The agent MUST preserve all content outside its managed block markers unchanged.
- What happens when a device's hostname changes after enrollment? The device identity is tied to its cryptographic key pair, not its hostname. The hostname is display-only metadata that can be updated.
- What happens when the hub's storage is corrupted or lost? Agents retain their last synced state until a new hub is configured.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The hub MUST maintain a registry of all devices with their public keys, status (pending/approved/revoked), and metadata (hostname, enrollment date, last sync time).
- **FR-002**: The hub MUST expose an authenticated web dashboard for device management, token generation, and status monitoring.
- **FR-003**: The hub MUST generate single-use, time-limited onboarding tokens for device enrollment.
- **FR-004**: The hub MUST provide a device list endpoint that agents call to retrieve the current set of approved device public keys.
- **FR-005**: The agent MUST run as a background service and check in with the hub on a configurable interval (default: 5 minutes).
- **FR-006**: The agent MUST write approved public keys into a clearly delimited managed block within the authorized_keys file, preserving all content outside that block.
- **FR-007**: The agent MUST use atomic file operations when modifying the authorized_keys file to prevent corruption.
- **FR-008**: The agent MUST authenticate to the hub using the device's existing SSH key pair (e.g., `~/.ssh/id_ed25519`). The same public key is used for both hub authentication and distribution to other devices' authorized_keys files.
- **FR-009**: The hub MUST reject sync requests from revoked devices.
- **FR-010**: The system MUST support the standard OpenSSH authorized_keys format.
- **FR-011**: All hub-agent communication MUST occur over encrypted connections.
- **FR-012**: The hub MUST require authentication for all dashboard and management operations via a password/passphrase configured at hub startup (e.g., via environment variable or config file).
- **FR-013**: The agent MUST NOT modify the authorized_keys file when the hub is unreachable; it MUST retain the last successfully synced state.
- **FR-014**: Onboarding tokens MUST be single-use and expire after a configurable duration (default: 24 hours).
- **FR-015**: Re-enrollment of a previously revoked device MUST create a new device record. The revoked record MUST be preserved for audit purposes. Device records are never reactivated.
- **FR-016**: The hub MUST log all state-change events: device enrollments, approvals, revocations, and failed authentication attempts. Routine sync heartbeats are NOT logged.

### Key Entities

- **Device**: Represents a machine in the mesh. Attributes: unique ID, display name/hostname, SSH public key (the device's existing key, used for both hub auth and SSH access), status (pending/approved/revoked), enrollment timestamp, last sync timestamp.
- **Onboarding Token**: A short-lived, single-use credential for enrolling a new device. Attributes: token value, creation timestamp, expiration timestamp, used-by device (once consumed).
- **Managed Key Block**: The demarcated section within a device's authorized_keys file that the agent owns. Contains only hub-distributed keys. Delimited by start/end markers.

### Assumptions

- This is a single-user system (one owner managing their personal devices). Multi-user/multi-tenant support is out of scope.
- The owner has root/admin access on all enrolled devices to install and run the agent.
- All devices have network connectivity to the hub (directly or via VPN). NAT traversal and relay are out of scope.
- The hub runs on a device that is always-on (e.g., a home server or VPS).
- Devices run operating systems that support the standard OpenSSH authorized_keys file format.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A new device can be enrolled and able to SSH to all other enrolled devices within 10 minutes of starting the enrollment process (including approval).
- **SC-002**: A revoked device's key is removed from all enrolled devices' authorized_keys files within one sync interval (default: 5 minutes) of pressing "Revoke."
- **SC-003**: The agent's managed block in authorized_keys never corrupts or removes manually added keys, verified across 1000 consecutive sync cycles.
- **SC-004**: When the hub is offline, existing device-to-device SSH connections continue working indefinitely using the last synced key set.
- **SC-005**: The onboarding flow (generate token → enroll device → approve → first sync) can be completed by a technically proficient user in under 5 minutes.
- **SC-006**: The dashboard displays accurate real-time status for all devices, with last-sync timestamps updating within one sync interval.
