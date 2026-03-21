# Research: Token Management Enhancements

**Feature**: 002-token-management | **Date**: 2026-03-21

## Research Tasks

### 1. Token Deletion Pattern in Go Slices

**Decision**: Use swap-with-last + truncate pattern for O(1) removal from `[]model.Token`.

**Rationale**: The token list is unordered (display order is irrelevant for storage). The swap-with-last pattern (`tokens[i] = tokens[len-1]; tokens = tokens[:len-1]`) avoids the cost of shifting elements that `append(tokens[:i], tokens[i+1:]...)` incurs. For small slices this is negligible, but it's the idiomatic Go approach for unordered slice removal.

**Alternatives considered**:
- `append(s[:i], s[i+1:]...)`: Preserves order but O(n). Acceptable for small slices but unnecessary since token storage order has no semantic meaning.
- `slices.Delete()` (Go 1.21+): Cleaner API but preserves order (shifts elements). No benefit over manual swap.

### 2. Clipboard API Browser Support

**Decision**: Use `navigator.clipboard.writeText()` with graceful degradation.

**Rationale**: The Clipboard API is supported in all modern browsers (Chrome 66+, Firefox 63+, Safari 13.1+, Edge 79+). It requires a secure context (HTTPS or localhost), which aligns with the hub's expected deployment. The existing Pico CSS framework doesn't include clipboard utilities, so inline JavaScript in the template is the simplest approach.

**Alternatives considered**:
- `document.execCommand('copy')`: Deprecated, but wider legacy support. Rejected because the hub targets modern admin browsers and the spec assumes secure context.
- Third-party clipboard library (e.g., clipboard.js): Rejected per constitution principle III (simplicity) and the no-external-dependencies constraint. Inline JS is ~10 lines.

### 3. Token Removal Confirmation UX

**Decision**: Use native browser `confirm()` dialog triggered by the form's `onclick` handler.

**Rationale**: The existing codebase has no custom modal/dialog components. Adding one would violate simplicity. The native `confirm()` dialog is universally supported, requires zero additional code, and provides the "one click + confirmation" pattern specified in SC-001.

**Alternatives considered**:
- Custom modal with Pico CSS: More polished UX but requires additional HTML/CSS/JS. Over-engineered for an admin tool with a single user.
- Two-step button (click to reveal "Are you sure?" then click again): More modern UX but more complex template logic. Not justified for this scope.

### 4. Expired Token Purge Trigger Strategy

**Decision**: Purge expired tokens lazily on `GET /tokens` (token list page load).

**Rationale**: The admin typically views the token page before and after generating tokens. Purging on page load means expired tokens are cleaned up whenever the admin interacts with tokens, without needing a background goroutine or cron job. This is the simplest approach that satisfies FR-009 ("during normal operations").

**Alternatives considered**:
- Background ticker goroutine: Ensures regular cleanup regardless of admin activity. Over-engineered — the hub is a small single-admin tool; orphaned expired tokens cost nothing until the page is loaded.
- Purge on every store operation: Too aggressive; adds overhead to unrelated operations (device updates, key sync).
- Purge on token creation: Only triggers when generating new tokens, so expired tokens could accumulate between creation events.

### 5. Audit Entry DeviceID for Admin Actions

**Decision**: Use empty string for `DeviceID` in `token_removed` audit entries (admin actions have no associated device).

**Rationale**: The existing `AuditEntry` struct uses `DeviceID` to identify the device involved in an event. Token removal is an admin action, not a device action. Using an empty string is consistent with the struct's semantics — the field is optional context, not a required foreign key. The audit template renders `<code>{{.DeviceID}}</code>`, which will render as an empty `<code>` element — visually acceptable.

**Alternatives considered**:
- Use "admin" as a sentinel value: Potentially confusing — could be mistaken for a device ID. Also inconsistent with the existing model where DeviceID is always a UUID-formatted string.
- Add a separate `Actor` field to `AuditEntry`: Clean design but modifies the data model beyond what this feature requires. Violates YAGNI.
