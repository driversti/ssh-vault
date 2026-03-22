# Research: Device Rename

## R-001: Inline Edit UX Pattern (Vanilla JS)

**Decision**: Use vanilla JavaScript with `contenteditable` or dynamic `<input>` replacement on the device name element. On click, swap the static text for an `<input>` field. On `blur` or Enter keydown, POST the new name via `fetch()`. On Escape, restore original text.

**Rationale**: The project uses no JS frameworks (constitution: standard library only, simplicity). The existing dashboard already uses inline JS for `confirm()` dialogs (revoke/remove buttons). A small `<script>` block in `devices.html` is consistent with this pattern.

**Alternatives considered**:
- `contenteditable` attribute: Simpler markup, but harder to control (pasting HTML, no `maxlength`). Rejected.
- Modal dialog: Requires more JS and breaks the "click to edit" requirement. Rejected.
- Dynamic `<input>` replacement: Best control over validation, `maxlength`, focus behavior. **Chosen**.

## R-002: API Endpoint Design

**Decision**: `POST /devices/{id}/rename` with JSON body `{"name": "new name"}`. Returns JSON `{"name": "trimmed name"}` on success or `{"error": "message"}` on failure.

**Rationale**: Follows existing pattern (`POST /devices/{id}/approve`, `/revoke`, `/remove`). POST is appropriate because it's a state-changing operation. JSON response (not redirect) is needed because the caller is JavaScript `fetch()`, not a form submission.

**Alternatives considered**:
- `PATCH /devices/{id}`: More RESTful, but the project uses `POST` for all device actions. Consistency wins.
- `PUT /devices/{id}/name`: Overly granular. Rejected.

## R-003: Audit Event Naming

**Decision**: Add `EventDeviceRenamed = "device_renamed"` constant to `model/audit.go`. Audit detail format: `"Device renamed from 'old-name' to 'new-name'"`.

**Rationale**: Follows existing naming convention (`EventEnrolled`, `EventApproved`, `EventRevoked`). The `device_renamed` event string uses the same snake_case pattern as `auth_failed`, `token_used`, etc.

**Alternatives considered**:
- `"renamed"`: Too terse, could be ambiguous. Rejected.
- `"device_name_changed"`: Longer than necessary. Rejected.

## R-004: Audit Event Pill Color

**Decision**: Map `device_renamed` to the `"used"` CSS class (blue/neutral) in `eventPillClass()`.

**Rationale**: Rename is an informational event, not a security action (approved/revoked) or error (auth_failed). The `"used"` class provides a neutral visual treatment, similar to `token_used` and `shortcode_used`.

## R-005: Client-Side Validation

**Decision**: Validate on both client and server. Client-side: check non-empty and `maxlength="255"` on the input. Server-side: trim whitespace, check non-empty, check max 255, check unchanged name — all before saving.

**Rationale**: Client-side validation provides instant feedback. Server-side validation is the authoritative check (never trust client input).
