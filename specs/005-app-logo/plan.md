# Implementation Plan: App Logo Integration

**Branch**: `005-app-logo` | **Date**: 2026-03-22 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/005-app-logo/spec.md`

## Summary

Add the existing `logo.svg` to the SSH Vault hub as both a browser favicon and a header logo. The SVG will be embedded in the binary alongside existing static assets (Pico CSS) and served via dedicated route handlers, following the established pattern.

## Technical Context

**Language/Version**: Go 1.26.1
**Primary Dependencies**: Standard library only (`embed`, `net/http`, `html/template`)
**Storage**: N/A (embedded static asset)
**Testing**: `go test` (standard testing package)
**Target Platform**: Web application (server-side rendered HTML)
**Project Type**: Web service (hub server)
**Performance Goals**: N/A (static asset serving)
**Constraints**: Must use `embed.FS` pattern consistent with existing `pico.min.css` serving
**Scale/Scope**: Single SVG file, 2 integration points (favicon link tag, header image)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Idiomatic Go | PASS | New handler follows `handleStaticCSS` pattern; standard library only |
| II. Testing | PASS | Handler testable with `httptest`; template changes verified manually |
| III. Simplicity | PASS | Minimal change: 1 new handler, 1 route, 2 HTML changes |
| Tech Stack | PASS | No new dependencies |
| Dev Workflow | PASS | Feature branch, PR-based |

No violations. No complexity tracking needed.

## Project Structure

### Documentation (this feature)

```text
specs/005-app-logo/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # N/A (no data entities)
├── quickstart.md        # Phase 1 output
└── contracts/           # N/A (no new external interfaces)
```

### Source Code (repository root)

```text
internal/hub/
├── server.go            # MODIFY: Add handleStaticLogo handler + route registration
└── templates/
    ├── layout.html      # MODIFY: Add favicon <link> and header logo <img>
    └── logo.svg         # ADD: Copy from repository root
```

**Structure Decision**: The logo.svg is placed in `internal/hub/templates/` to be captured by the existing `//go:embed templates/*` directive. No new directories or packages needed.

## Design Decisions

### 1. Logo Serving Approach

**Decision**: Dedicated handler at `/static/logo.svg` (mirrors `/static/pico.min.css`)

**Rationale**: Consistent with existing architecture. Each static asset has an explicit route and handler. No generic file server needed.

**Alternatives rejected**:
- Generic `http.FileServer` on `/static/` — violates simplicity principle; only 2 static files
- Inline SVG in HTML template — works for header but not for favicon; duplicates content

### 2. Favicon Format

**Decision**: SVG favicon via `<link rel="icon" type="image/svg+xml" href="/static/logo.svg">`

**Rationale**: SVG is natively supported by all modern browsers (Chrome 80+, Firefox 41+, Safari 16+, Edge 80+). No need for `.ico` conversion or multiple sizes. Graceful degradation — older browsers show default icon.

### 3. Header Logo Placement

**Decision**: Add `<img>` tag before the "SSH Vault" `<strong>` text inside the existing `<li>` element in the `<nav>`.

**Rationale**: Keeps the logo and app name together as a single branding unit. The Pico CSS framework handles nav layout. Size constrained via CSS height (matching nav bar height).

### 4. Logo File Location

**Decision**: Copy `logo.svg` to `internal/hub/templates/logo.svg`

**Rationale**: The `//go:embed templates/*` directive already embeds everything in that directory. Placing the SVG there requires zero embed configuration changes.

## Files to Modify

| File | Action | Change Description |
|------|--------|--------------------|
| `internal/hub/templates/logo.svg` | CREATE | Copy logo.svg from repository root |
| `internal/hub/server.go` | MODIFY | Add `handleStaticLogo` handler + register `/static/logo.svg` route |
| `internal/hub/templates/layout.html` | MODIFY | Add favicon `<link>` in `<head>` and logo `<img>` in header `<nav>` |

## Complexity Tracking

No violations to justify.
