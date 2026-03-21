# Implementation Plan: Token Management Enhancements

**Branch**: `002-token-management` | **Date**: 2026-03-21 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/002-token-management/spec.md`

## Summary

Enhance the token lifecycle with three capabilities: (1) administrators can remove unused tokens via the dashboard, (2) token usage and removal events are recorded in the audit log with truncated token prefixes, and (3) a copy-to-clipboard button is added next to each token. Additionally, expired tokens are automatically purged from storage during normal operations.

## Technical Context

**Language/Version**: Go 1.26.1
**Primary Dependencies**: `golang.org/x/crypto/ssh` (SSH key parsing), standard library only
**Storage**: JSON file via `FileStore` (atomic write with temp file + rename, `sync.RWMutex` for concurrency)
**Testing**: Standard `testing` package, table-driven tests, `t.TempDir()` for test isolation
**Target Platform**: Linux/macOS server (single binary)
**Project Type**: CLI + web-service (single binary, `cmd/ssh-vault/main.go`)
**Performance Goals**: N/A — admin tool, low concurrency
**Constraints**: No third-party dependencies beyond `golang.org/x/crypto`; embedded templates via `//go:embed`
**Scale/Scope**: Single administrator, handful of devices/tokens

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Idiomatic Go | ✅ Pass | All changes use standard library, explicit error handling, minimal exports |
| II. Testing | ✅ Pass | Table-driven tests planned for all new store methods; handler tests for new routes |
| III. Simplicity | ✅ Pass | No new abstractions — extends existing patterns (store methods, handler functions, template blocks). No new packages or dependencies |

**Gate result**: PASS — no violations.

## Project Structure

### Documentation (this feature)

```text
specs/002-token-management/
├── plan.md              # This file
├── spec.md              # Feature specification
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
└── tasks.md             # Phase 2 output (via /speckit.tasks)
```

### Source Code (repository root)

```text
cmd/ssh-vault/             # Entry point (no changes needed)
internal/
├── model/
│   ├── audit.go           # MODIFY: add EventTokenUsed, EventTokenRemoved constants
│   ├── token.go           # NO CHANGE (model is sufficient)
│   ├── token_test.go      # NO CHANGE
│   └── store.go           # NO CHANGE
├── hub/
│   ├── store.go           # MODIFY: add RemoveToken(), PurgeExpiredTokens()
│   ├── store_test.go      # MODIFY: add tests for new store methods
│   ├── handlers.go        # MODIFY: add handleRemoveToken(), add token_used audit in handleEnroll
│   ├── handlers_test.go   # MODIFY: add tests for POST /tokens/{value}/remove endpoint
│   ├── server.go          # MODIFY: register new route, add token purge call in handleTokens GET
│   └── templates/
│       ├── tokens.html    # MODIFY: add remove button + copy icon per token row
│       ├── audit.html     # NO CHANGE (badge-{{.Event}} already renders dynamically)
│       └── layout.html    # MODIFY: add CSS for copy button, badge-token_used, badge-token_removed
├── agent/                 # NO CHANGE
└── keyblock/              # NO CHANGE
```

**Structure Decision**: No new packages or directories. All changes extend existing files following established patterns.

## Implementation Phases

### Phase 1: Store Layer — RemoveToken & PurgeExpiredTokens

**Files**: `internal/hub/store.go`, `internal/hub/store_test.go`

**Changes to `store.go`**:

1. **`RemoveToken(value string) error`** — Acquires write lock, iterates `fs.data.Tokens`, finds token by value. If token is used (`t.Used == true`), returns error "cannot remove used token". If not found, returns error "token not found". Otherwise, removes from slice (swap with last + truncate pattern) and calls `Save()`.

2. **`PurgeExpiredTokens() (int, error)`** — Acquires write lock, filters `fs.data.Tokens` in-place to remove only expired tokens (used tokens are retained — they are already filtered from the UI and serve as historical records). Returns count of purged tokens. Calls `Save()` only if any were removed.

**Tests for `store_test.go`** (table-driven):

- `TestFileStore_RemoveToken` — add token, remove it, verify `GetToken` returns not found
- `TestFileStore_RemoveToken_Used` — add token, use it, attempt remove, verify error
- `TestFileStore_RemoveToken_NotFound` — attempt to remove nonexistent token, verify error
- `TestFileStore_RemoveToken_PreservesOthers` — add 3 tokens, remove middle, verify other 2 intact
- `TestFileStore_PurgeExpiredTokens` — add mix of expired/valid tokens, purge, verify only valid remain
- `TestFileStore_PurgeExpiredTokens_NoneExpired` — all valid, purge, verify no save called (count=0)

### Phase 2: Audit Events — New Constants & Token Usage Audit

**Files**: `internal/model/audit.go`, `internal/hub/handlers.go`

**Changes to `audit.go`**:

Add two new event constants:
```go
EventTokenUsed    = "token_used"
EventTokenRemoved = "token_removed"
```

**Changes to `handlers.go`**:

1. **In `handleEnroll`** (after line 131): Add a `token_used` audit entry alongside the existing `enrolled` entry. The detail string includes the truncated token prefix (first 8 chars) and the device name. Note: line 133 already uses `req.Token[:8]` — the `enrolled` event already references the token. The new `token_used` event uses `EventTokenUsed` as event type with the same truncated prefix pattern.

2. **New `handleRemoveToken(w, r)`**: Handles `POST /tokens/{value}/remove`. Extracts token value from URL path using `extractPathParam`. Validates token exists and is not used. Calls `store.RemoveToken(value)`. Creates `token_removed` audit entry with truncated prefix (`value[:8]`). Redirects to `/tokens`. Note: The `DeviceID` field in the audit entry will be empty (admin action, not device action).

### Phase 3: Routes & Token Purge Integration

**Files**: `internal/hub/server.go`

**Changes**:

1. **Register new route** in `registerRoutes()`:
   ```go
   s.mux.HandleFunc("/tokens/", s.requireSession(s.handleTokenAction))
   ```
   Where `handleTokenAction` routes to `handleRemoveToken` for paths ending in `/remove`.

2. **Add purge call** in `handleTokens` GET handler (around line 135): Before listing tokens, call `s.store.PurgeExpiredTokens()`. This is the "during normal operations" trigger specified in FR-009. Log the count if any were purged.

### Phase 4: UI — Copy Icon, Remove Button, Audit Badge Styles

**Files**: `internal/hub/templates/tokens.html`, `internal/hub/templates/layout.html`

**Changes to `tokens.html`**:

1. Add "Actions" column to the table header.
2. For each token row, add:
   - A copy button (using a clipboard SVG icon inline or Unicode character like 📋) with `onclick` handler calling `navigator.clipboard.writeText()`. On success, briefly change the icon/text to a checkmark as visual feedback.
   - A remove form: `<form method="post" action="/tokens/{{.Value}}/remove">` with a styled delete button and `onclick="return confirm('Remove this token?')"` for confirmation (SC-001: "one click + confirmation").

**Changes to `layout.html`** (CSS):

1. Add `.copy-btn` styles — inline button, no border, cursor pointer, subtle hover effect.
2. Add `.badge-token_used` — distinct color (e.g., blue/info: `#17a2b8`) for token usage events.
3. Add `.badge-token_removed` — distinct color (e.g., orange/warning: `#fd7e14`) for token removal events.

**No changes to `audit.html`** — it already renders badges dynamically using `badge-{{.Event}}`, so the new event types will automatically pick up the CSS classes.

### Phase 5: Handler Tests

**Files**: `internal/hub/handlers_test.go`

**Tests**:
- `TestHandleRemoveToken_Success` — POST to `/tokens/{value}/remove` with valid unused token, verify redirect and token removed from store
- `TestHandleRemoveToken_UsedToken` — attempt to remove a used token, verify error response
- `TestHandleRemoveToken_NotFound` — attempt to remove nonexistent token, verify error
- `TestHandleEnroll_AuditTokenUsed` — enroll a device, verify `token_used` audit entry exists
- `TestHandleTokens_PurgesExpired` — add expired tokens, GET `/tokens`, verify they're purged

### Build & Verification Sequence

1. `go build ./...` — verify compilation
2. `go test ./internal/model/...` — model tests (unchanged, regression check)
3. `go test ./internal/hub/...` — store + handler tests (new + existing)
4. `go vet ./...` — static analysis
5. `go test ./...` — full suite

## Complexity Tracking

No constitution violations to justify — all changes extend existing patterns.
