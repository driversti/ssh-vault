# Feature Specification: Token Management Enhancements

**Feature Branch**: `002-token-management`
**Created**: 2026-03-21
**Status**: Draft
**Input**: User description: "As a User, I want to be able to remove a generated token if it wasn't used. When a token was used, a new record should appear in the audit log. I also want a copy icon next to the token for easy copying."

## Clarifications

### Session 2026-03-21

- Q: Should token removal be audited? → A: Yes, create a `token_removed` audit entry when a token is deleted.
- Q: How should expired tokens be handled? → A: Automatically purge expired tokens from storage during normal operations.
- Q: How much token detail should audit entries include? → A: Include a truncated prefix (first 8 characters) for identification without full credential exposure.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Remove Unused Token (Priority: P1)

As a hub administrator, I want to remove enrollment tokens that I generated but no longer need, so that I can keep my token list clean and prevent unauthorized devices from enrolling with tokens I no longer intend to distribute.

**Why this priority**: Token removal is the core security and housekeeping capability requested. Without it, unused tokens persist until expiry, creating an unnecessary enrollment window and cluttering the dashboard. This is the highest-value story because it directly addresses a gap in token lifecycle management.

**Independent Test**: Can be fully tested by generating a token, verifying it appears in the token list, clicking remove, and confirming it disappears from the list and can no longer be used for enrollment.

**Acceptance Scenarios**:

1. **Given** an active (unused, not expired) token exists, **When** the administrator clicks the remove action for that token, **Then** the token is permanently deleted and no longer appears in the token list.
2. **Given** an active token exists, **When** the administrator removes it and a device later attempts to enroll with that token value, **Then** the enrollment is rejected as if the token never existed.
3. **Given** multiple active tokens exist, **When** the administrator removes one token, **Then** all other tokens remain unaffected and still valid for enrollment.

---

### User Story 2 - Token Usage Audit Trail (Priority: P2)

As a hub administrator, I want to see a record in the audit log whenever a token is used for device enrollment, so that I have a complete trail of how enrollment tokens are consumed.

**Why this priority**: Audit visibility is essential for security oversight. Currently, device enrollment creates an audit entry, but there is no explicit record tying a specific token to its usage event. This story closes the observability gap by adding a token-specific audit event.

**Independent Test**: Can be fully tested by generating a token, using it to enroll a device, and then checking the audit log for a new entry that identifies the token and the enrolling device.

**Acceptance Scenarios**:

1. **Given** an active token exists, **When** a device successfully enrolls using that token, **Then** a new audit log entry appears recording the token usage event, including which device consumed the token.
2. **Given** the audit log page is open, **When** a token usage event is recorded, **Then** the event is visually distinguishable from other event types (e.g., distinct badge color or label).
3. **Given** multiple enrollment events have occurred, **When** the administrator views the audit log, **Then** token usage events appear in reverse chronological order alongside other audit events.

---

### User Story 3 - Copy Token to Clipboard (Priority: P3)

As a hub administrator, I want a copy icon next to each token on the dashboard so I can quickly copy the token value to my clipboard with a single click, making it easy to share with devices that need to enroll.

**Why this priority**: This is a usability improvement. The current UI already supports selecting the full token text, but a dedicated copy button reduces friction — especially on mobile or when tokens are long. It is the lowest priority because the feature works without it, just with slightly more effort.

**Independent Test**: Can be fully tested by generating a token, clicking the copy icon next to it, pasting into a text editor, and verifying the pasted value matches the token exactly.

**Acceptance Scenarios**:

1. **Given** an active token is displayed on the tokens page, **When** the administrator clicks the copy icon next to it, **Then** the full token value is copied to the system clipboard.
2. **Given** an active token is displayed, **When** the administrator clicks the copy icon, **Then** visual feedback is shown confirming the copy succeeded (e.g., icon change, tooltip, or brief confirmation text).
3. **Given** multiple tokens are listed, **When** the administrator clicks the copy icon for a specific token, **Then** only that token's value is copied, not others.

---

### Edge Cases

- What happens when an administrator tries to remove a token that was just used by a device moments ago (race condition)?
  - The system should check the token's current state before deletion. If the token was used between page load and the delete action, the removal should be rejected with an appropriate message.
- What happens when the clipboard API is unavailable (e.g., non-HTTPS context or unsupported browser)?
  - The copy icon should degrade gracefully — either hide the icon or show an error tooltip explaining clipboard access is not available. The existing text-selection behavior remains as a fallback.
- What happens if the administrator removes a token and then refreshes the page?
  - The removed token should not reappear. Deletion is permanent and persisted.
- What happens to expired tokens over time?
  - Expired tokens are automatically purged from storage during normal operations. They do not accumulate indefinitely.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST allow authenticated administrators to permanently remove any unused, non-expired token from the dashboard.
- **FR-002**: System MUST prevent removal of tokens that have already been used for device enrollment.
- **FR-003**: System MUST reject enrollment attempts that use a previously removed token.
- **FR-004**: System MUST create an audit log entry whenever a token is successfully consumed during device enrollment, recording the event type, the enrolling device identity, a truncated token prefix (first 8 characters), and a timestamp.
- **FR-005**: System MUST display a copy-to-clipboard control adjacent to each active token on the tokens page.
- **FR-006**: System MUST provide visual feedback to the administrator after a successful clipboard copy action.
- **FR-007**: System MUST handle concurrent token removal and usage gracefully, ensuring data consistency (a token cannot be both used and removed).
- **FR-008**: System MUST create an audit log entry whenever an administrator removes a token, recording the event type, a truncated token prefix (first 8 characters), and a timestamp.
- **FR-009**: System MUST automatically purge expired tokens from storage during normal operations (e.g., on token list access or periodic cleanup), without requiring administrator action.

### Key Entities

- **Token**: Represents an enrollment credential. Key attributes: unique value, creation time, expiration time, usage status, consuming device identity. Tokens can be hard-deleted (permanently removed from storage) by an administrator before use, or automatically purged after expiration.
- **Audit Entry**: Records a significant system event. Key attributes: timestamp, event type, associated device, human-readable details. New event types: "token_used" (records when a token is consumed during enrollment) and "token_removed" (records when an administrator deletes an unused token).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Administrators can remove an unused token in a single action (one click + confirmation) from the tokens page.
- **SC-002**: 100% of token usage events during enrollment produce a corresponding audit log entry visible on the audit page.
- **SC-003**: Copying a token value to the clipboard requires exactly one click on the copy control.
- **SC-004**: Removed tokens are immediately unavailable for enrollment — zero delay between removal and enforcement.
- **SC-005**: The audit log correctly attributes each token usage event to the specific device that consumed it.

## Assumptions

- The administrator is the only user role interacting with the dashboard (consistent with the existing single-password authentication model).
- Token removal is a hard delete — removed tokens are permanently discarded, not soft-deleted or archived.
- The copy-to-clipboard feature relies on the browser's Clipboard API, which requires a secure context (HTTPS or localhost). The hub is expected to run in such a context.
- The "token_used" audit event is a new event type added alongside the existing event types (enrolled, approved, revoked, auth_failed).
