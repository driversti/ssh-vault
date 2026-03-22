# Implementation Plan: Device Rename

**Branch**: `007-device-rename` | **Date**: 2026-03-22 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/007-device-rename/spec.md`

## Summary

Add inline device renaming to the dashboard. Users click a device name to edit it in-place; blur or Enter saves the new name via a JSON API endpoint; Escape cancels. Every successful rename creates an audit log entry recording old and new names.

## Technical Context

**Language/Version**: Go 1.26.1
**Primary Dependencies**: `golang.org/x/crypto/ssh` (existing), standard library only
**Storage**: JSON file via `FileStore` (atomic write with temp file + rename, `sync.RWMutex` for concurrency)
**Testing**: `go test` with `net/http/httptest`, table-driven tests
**Target Platform**: Linux server (self-hosted)
**Project Type**: Web service (single binary, embedded templates)
**Performance Goals**: N/A — single-user admin dashboard
**Constraints**: Standard library only, no JavaScript frameworks

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Idiomatic Go | PASS | Handler follows established `extractPathParam` → `GetDevice` → `UpdateDevice` → audit pattern. `gofmt`/`goimports` enforced. |
| II. Testing | PASS | Table-driven handler test with `httptest`. Tests alongside code in `*_test.go`. |
| III. Simplicity | PASS | No new packages, no abstractions. Inline edit via vanilla JS. Single new endpoint + event constant. |

**Post-Phase 1 re-check**: PASS — no new dependencies, no new packages, no abstraction layers introduced.

## Project Structure

### Documentation (this feature)

```text
specs/007-device-rename/
├── spec.md
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
└── contracts/
    └── api.md           # Phase 1 output
```

### Source Code (repository root)

```text
internal/
├── model/
│   └── audit.go          # Add EventDeviceRenamed constant
├── hub/
│   ├── server.go         # Add /rename route to handleDeviceAction
│   ├── handlers.go       # Add handleRename handler (or new file)
│   ├── handlers_test.go  # Add rename handler tests
│   └── templates/
│       └── devices.html  # Add inline edit JS + editable name markup
```

**Structure Decision**: All changes fit within existing files. No new packages or directories needed (except spec artifacts).

## Complexity Tracking

No violations — no entries needed.
