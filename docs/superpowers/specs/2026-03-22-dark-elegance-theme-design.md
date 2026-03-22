# Dark Elegance Theme — Design Spec

**Date:** 2026-03-22
**Status:** Approved
**Scope:** Replace Pico CSS with custom dual-theme (light/dark) stylesheet following Apple design language

## Overview

Replace the current Pico CSS framework and inline styles with a single custom `theme.css` file that supports both light and dark themes via `prefers-color-scheme` media query. The design follows Apple's "Dark Elegance" aesthetic — clean typography, pill-shaped controls, subtle glass effects, and intentionally different color palettes per theme.

## File Changes

| Action   | File                    | Purpose                                                        |
|----------|-------------------------|----------------------------------------------------------------|
| Create   | `templates/theme.css`   | All styles — CSS variables, light/dark palettes, components    |
| Delete   | `templates/pico.min.css`| No longer needed                                              |
| Edit     | `templates/layout.html` | Remove inline `<style>`, add `color-scheme` meta, update `<link>` href, rework nav to segmented control |
| Edit     | `templates/devices.html`| Device avatars, pill badges, stats bar, stale dot. Remove standalone Fingerprint column — move `formatFingerprint` output into device cell as `.device-info-fp` |
| Edit     | `templates/tokens.html` | cmd-block, token-chip, updated buttons, confirm dialog restyled |
| Edit     | `templates/audit.html`  | Pill badges with event-to-pill mapping, secondary text styling |
| Edit     | `templates/login.html`  | Centered card with gradient heading. Replace `var(--pico-del-color)` on error paragraph with `var(--red)` |
| Edit     | `server.go`             | Rename static route path AND `ReadFile` call inside `handleStaticCSS` body (line 244: `templates/pico.min.css` → `templates/theme.css`), add `countByStatus` helper, add `eventPillClass` template function |

No new Go dependencies. Existing template functions (`formatTime`, `formatTimePtr`, `formatFingerprint`, `isStale`, `upper`) remain unchanged. One new template function added: `eventPillClass`.

## Theme System

CSS custom properties on `:root` define the light palette as default. A `@media (prefers-color-scheme: dark)` block overrides all variables for dark mode. No JavaScript required — the browser handles switching automatically.

### Variable Groups

- **Surfaces:** `--bg`, `--bg-elevated`, `--bg-secondary`, `--bg-tertiary`
- **Borders:** `--border`, `--border-strong`
- **Text:** `--text`, `--text-secondary`, `--text-tertiary`
- **Accent:** `--accent`, `--accent-glow`
- **Status colors:** `--green`, `--yellow`, `--red`, `--teal` + matching `*-bg` variants
- **Contextual:** `--nav-bg`, `--nav-active-bg`, `--row-hover`, `--heading-gradient-from/to`, `--card-shadow`, `--login-shadow`

### Color Strategy

Follows Apple's dual-theme approach — different values per theme, not inversions:

| Token      | Light        | Dark         |
|------------|--------------|--------------|
| `--bg`     | `#ffffff`    | `#000000`    |
| `--bg-elevated` | `#f5f5f7` | `#1c1c1e` |
| `--text`   | `#1d1d1f`    | `#f5f5f7`    |
| `--accent`  | `#0071e3`   | `#0a84ff`    |
| `--green`  | `#248a3d`    | `#30d158`    |
| `--yellow` | `#9a6700`    | `#ffd60a`    |
| `--red`    | `#d70015`    | `#ff453a`    |
| `--teal`   | `#0071e3`    | `#64d2ff`    |

### HTML Head Changes

- Add: `<meta name="color-scheme" content="light dark">`
- Remove: `data-theme="light"` attribute from `<html>` (Pico-specific)
- Change: `<link>` href from `/static/pico.min.css` to `/static/theme.css`

## Component Design

### Navigation

Sticky top bar with `backdrop-filter: saturate(180%) blur(20px)`. Three-section layout:
- Left: logo SVG + "SSH Vault" brand text
- Center: segmented control (iOS-style pill switcher) with page links
- Right: "Log out" button

Active nav link gets `--nav-active-bg` background + subtle shadow.

### Cards

- Background: `--bg-elevated`
- Border: `0.5px solid var(--border)`
- Border-radius: `12px`
- Shadow: subtle in light mode (`0 1px 3px rgba(0,0,0,0.04)`), none in dark
- Used for: table wrappers, login form

### Tables

- Full-width inside cards
- Headers: uppercase, small font, `--text-tertiary`
- Rows: subtle hover (`--row-hover`), `0.5px` border between rows, none on last
- No outer border (card provides the container)

### Status Pills

Pill-shaped badges (`border-radius: 100px`) with a 5px glowing dot (`box-shadow: 0 0 6px currentColor`).

Variants for device status (used in `devices.html` via `pill-{{.Status}}`):
- `pill-approved` — green
- `pill-pending` — yellow
- `pill-revoked` — red

Variants for audit events (used in `audit.html` via `pill-{{eventPillClass .Event}}`):
- `pill-approved` — green (event: `approved`)
- `pill-enrolled` — green (event: `enrolled`)
- `pill-revoked` — red (events: `revoked`, `auth_failed`)
- `pill-used` — teal (events: `token_used`, `shortcode_used`, `shortcode_created`)
- `pill-expired` — muted gray (events: `token_removed`, `shortcode_expired`)

A new `eventPillClass` template function in `server.go` maps event strings to pill class suffixes:

```go
"eventPillClass": func(event string) string {
    switch event {
    case "approved":
        return "approved"
    case "enrolled":
        return "enrolled"
    case "revoked", "auth_failed":
        return "revoked"
    case "token_used", "shortcode_used", "shortcode_created":
        return "used"
    case "token_removed", "shortcode_expired":
        return "expired"
    default:
        return "expired"
    }
},
```

Variants for token/short-code status (used in `tokens.html`):
- `pill-active` — green
- `pill-used` — teal
- `pill-expired` — muted gray

### Buttons

Pill-shaped (`border-radius: 100px`). Three variants:
- **Primary:** accent background + glow shadow, `scale(1.02)` on hover
- **Secondary:** `--bg-secondary` background + border
- **Danger:** red-tinted background + red text + red border

Size modifier: `.btn-sm` for table action buttons.

### Confirm Dialog (Tokens Page)

The existing `<dialog>` for token removal confirmation is preserved and restyled:
- Dialog backdrop: `rgba(0,0,0,0.25)` with `backdrop-filter: blur(2px)`
- Dialog card: `--bg-elevated` background, `--border`, 16px radius, max-width 360px
- Cancel button: `.btn-secondary` styling
- Confirm button: `.btn-danger` styling
- CSS classes: `.confirm-dialog`, `.btn-cancel` → `.btn-secondary`, `.btn-confirm` → `.btn-danger`

### Copy Button

`.copy-btn`: borderless icon button, `color: var(--text-tertiary)`, hover changes to `var(--accent)`. This replaces the current Pico-dependent definition in the inline `<style>` block — both definitions must not coexist.

### Device Rows

- 36px avatar box: gradient background (`--bg-secondary` to `--bg-tertiary`), monitor SVG icon
- Name (font-weight 500) + fingerprint in monospace below (fingerprint moves from its own column into the device cell as `.device-info-fp`)
- Stale rows: `opacity: 0.45` + glowing red dot (`box-shadow: 0 0 6px var(--red)`) after name
- "Awaiting verification" text for unverified pending devices: styled with `color: var(--text-secondary); font-size: 0.82rem`

### Stats Bar (Devices Page)

Four flex cards showing: Total, Approved, Pending, Revoked.
- Large numbers (1.6rem, weight 700) colored with `--green`, `--yellow`, `--red`
- Small uppercase labels in `--text-tertiary`
- Same card styling as table wrappers

### Login Page

- Centered card (max-width 360px) with logo SVG
- Gradient text heading (`--heading-gradient-from` to `--heading-gradient-to`)
- Input with accent focus ring (`box-shadow: 0 0 0 3px var(--accent-glow)`)
- Full-width primary button

### Page Headings

- Gradient text: `linear-gradient(135deg, var(--heading-gradient-from), var(--heading-gradient-to))`
- Font-weight 700, letter-spacing `-0.03em`

## Template Data Changes

One addition to `handleDashboard` in `server.go`:

```go
s.renderTemplate(w, "devices.html", map[string]any{
    "Devices":       devices,
    "TotalCount":    len(devices),
    "ApprovedCount": countByStatus(devices, model.StatusApproved),
    "PendingCount":  countByStatus(devices, model.StatusPending),
    "RevokedCount":  countByStatus(devices, model.StatusRevoked),
})
```

New helper function:

```go
func countByStatus(devices []model.Device, status string) int {
    n := 0
    for _, d := range devices {
        if d.Status == status {
            n++
        }
    }
    return n
}
```

All other handlers pass the same data as before.

## Route Change

| Before                       | After                       |
|------------------------------|-----------------------------|
| `/static/pico.min.css`       | `/static/theme.css`         |
| reads `templates/pico.min.css` | reads `templates/theme.css` |

Two changes inside `handleStaticCSS`:
1. Route registration: `"/static/pico.min.css"` → `"/static/theme.css"`
2. Function body `ReadFile` call: `templateFS.ReadFile("templates/pico.min.css")` → `templateFS.ReadFile("templates/theme.css")`

Caching headers remain identical.

## Migration & Compatibility

**Removed:**
- `pico.min.css` (~83KB embedded file) — must be `git rm`'d since `//go:embed templates/*` would otherwise silently embed the dead file
- Inline `<style>` block in `layout.html` — all styles (including `.copy-btn`, `.confirm-dialog`, `.btn-cancel`, `.btn-confirm`) move to `theme.css`. Both must not coexist to avoid cascade conflicts.
- `data-theme="light"` attribute (Pico-specific)
- `var(--pico-del-color)` in `login.html` error message → replaced with `var(--red)`

**Preserved:**
- All Go handlers, template functions, form actions, URL paths
- `logo.svg` and its handler
- Timestamp localization `<script>` in `layout.html`
- Confirm dialog logic in `tokens.html`

**Risk:** Low. Pure frontend change — no data model, API, or business logic modifications. Only Go code changes are the static route path and a small `countByStatus` helper.

**Testing:** Visual verification in browser under both light and dark OS settings. Existing Go handler tests are unaffected.

## Reference

Design prototype: `designs/design-3-dark-elegance.html` (interactive, includes theme toggle for preview)
