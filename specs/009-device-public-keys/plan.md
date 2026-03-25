# Implementation Plan: Device Public Keys Display

**Branch**: `009-device-public-keys` | **Date**: 2026-03-22 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/009-device-public-keys/spec.md`

## Summary

Replace the truncated fingerprint display on the Devices page with parsed SSH public key metadata (key type, username, host/IP) extracted from the key comment field at display time. Add an expand/collapse mechanism to view and copy the full public key. No new persistence — all metadata is derived from the existing `Device.PublicKey` field.

## Technical Context

**Language/Version**: Go 1.26.1
**Primary Dependencies**: Standard library (`html/template`, `strings`), existing `golang.org/x/crypto/ssh`
**Storage**: JSON file via `FileStore` (no changes to persistence layer)
**Testing**: `go test` with standard `testing` package, table-driven tests
**Target Platform**: Linux server (hub binary)
**Project Type**: Web service (single binary with embedded templates)
**Performance Goals**: Same rendering performance as current page (< 100ms for ~50 devices)
**Constraints**: No new dependencies, standard library only for new code
**Scale/Scope**: Single-user hub, typically < 50 devices

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Idiomatic Go | ✅ Pass | New template functions follow existing patterns; `strings.Cut` for comment parsing |
| II. Testing | ✅ Pass | Table-driven tests for comment parsing logic; template rendering coverage via existing patterns |
| III. Simplicity | ✅ Pass | No new packages, no new persistence, pure display-time derivation; expand/collapse via vanilla JS |

**Gate result**: PASS — no violations.

## Project Structure

### Documentation (this feature)

```text
specs/009-device-public-keys/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
└── tasks.md             # Phase 2 output (/speckit.tasks)
```

### Source Code (repository root)

```text
cmd/ssh-vault/         # Entry point (no changes)
internal/hub/
├── server.go          # Add new template functions (keyUser, keyHost, keyType)
├── server_test.go     # New: tests for keyUser, keyHost, keyType functions
├── templates/
│   ├── devices.html   # Update: replace fingerprint with key metadata + expand/collapse
│   └── theme.css      # Update: add styles for key metadata display and expand/collapse
└── ...
internal/model/        # No changes
```

**Structure Decision**: Minimal changes within existing hub package. New template functions in `server.go` alongside existing `formatFingerprint`, `isStale`, etc. New test file `server_test.go` for the parsing logic. Template and CSS updates in-place.

## Complexity Tracking

No violations to justify — plan uses existing patterns throughout.
