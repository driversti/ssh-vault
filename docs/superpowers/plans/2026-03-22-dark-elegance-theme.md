# Dark Elegance Theme Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace Pico CSS with a custom dual-theme (light/dark) stylesheet following Apple's Dark Elegance design language.

**Architecture:** Single `theme.css` file with CSS custom properties. Light palette on `:root`, dark palette via `@media (prefers-color-scheme: dark)`. No JS needed for theme switching. One new Go template function (`eventPillClass`) and one helper (`countByStatus`).

**Tech Stack:** Go `html/template`, CSS custom properties, `prefers-color-scheme` media query

**Spec:** `docs/superpowers/specs/2026-03-22-dark-elegance-theme-design.md`
**Prototype:** `designs/design-3-dark-elegance.html`

---

## File Structure

| Action | File | Responsibility |
|--------|------|----------------|
| Create | `internal/hub/templates/theme.css` | All styles — variables, light/dark palettes, components |
| Delete | `internal/hub/templates/pico.min.css` | No longer needed |
| Rewrite | `internal/hub/templates/layout.html` | Shell: head, nav, footer script |
| Rewrite | `internal/hub/templates/devices.html` | Device list with stats bar |
| Rewrite | `internal/hub/templates/tokens.html` | Enrollment links + onboarding tokens |
| Rewrite | `internal/hub/templates/audit.html` | Audit log |
| Rewrite | `internal/hub/templates/login.html` | Login form |
| Edit | `internal/hub/server.go` | Route rename, new template funcs, dashboard data |

---

### Task 1: Create theme.css

Extract CSS from the design prototype into a production stylesheet, removing demo-only styles.

**Files:**
- Create: `internal/hub/templates/theme.css`

- [ ] **Step 1: Create theme.css**

Copy the CSS from `designs/design-3-dark-elegance.html` lines 9–557 into `internal/hub/templates/theme.css`. Remove these demo-only rules: `.demo-label`, `.demo-page`, `.theme-toggle`, `.theme-toggle button`. Add the `pill-enrolled` variant (same as `pill-approved` — green). Add confirm dialog styles.

The full file contents:

```css
/* ── Reset ── */
*, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }

/* ── Light theme (default) ── */
:root {
  --bg: #ffffff;
  --bg-elevated: #f5f5f7;
  --bg-secondary: #ebebed;
  --bg-tertiary: #e0e0e2;
  --border: rgba(0,0,0,0.06);
  --border-strong: rgba(0,0,0,0.12);
  --text: #1d1d1f;
  --text-secondary: #6e6e73;
  --text-tertiary: #98989d;
  --accent: #0071e3;
  --accent-glow: rgba(0,113,227,0.15);
  --green: #248a3d;
  --green-bg: rgba(36,138,61,0.1);
  --yellow: #9a6700;
  --yellow-bg: rgba(154,103,0,0.08);
  --red: #d70015;
  --red-bg: rgba(215,0,21,0.08);
  --teal: #0071e3;
  --teal-bg: rgba(0,113,227,0.08);
  --expired-bg: rgba(0,0,0,0.04);
  --nav-bg: rgba(255,255,255,0.72);
  --nav-active-bg: rgba(0,0,0,0.06);
  --nav-active-shadow: 0 1px 3px rgba(0,0,0,0.06);
  --row-hover: rgba(0,0,0,0.015);
  --heading-gradient-from: #1d1d1f;
  --heading-gradient-to: #6e6e73;
  --login-shadow: 0 0 80px rgba(0,113,227,0.05);
  --card-shadow: 0 1px 3px rgba(0,0,0,0.04);
  --logout-hover-border: rgba(0,0,0,0.25);
}

/* ── Dark theme ── */
@media (prefers-color-scheme: dark) {
  :root {
    --bg: #000000;
    --bg-elevated: #1c1c1e;
    --bg-secondary: #2c2c2e;
    --bg-tertiary: #3a3a3c;
    --border: rgba(255,255,255,0.08);
    --border-strong: rgba(255,255,255,0.14);
    --text: #f5f5f7;
    --text-secondary: #98989d;
    --text-tertiary: #636366;
    --accent: #0a84ff;
    --accent-glow: rgba(10,132,255,0.2);
    --green: #30d158;
    --green-bg: rgba(48,209,88,0.12);
    --yellow: #ffd60a;
    --yellow-bg: rgba(255,214,10,0.1);
    --red: #ff453a;
    --red-bg: rgba(255,69,58,0.1);
    --teal: #64d2ff;
    --teal-bg: rgba(100,210,255,0.1);
    --expired-bg: rgba(142,142,147,0.1);
    --nav-bg: rgba(0,0,0,0.75);
    --nav-active-bg: var(--bg-secondary);
    --nav-active-shadow: 0 1px 4px rgba(0,0,0,0.3);
    --row-hover: rgba(255,255,255,0.02);
    --heading-gradient-from: #f5f5f7;
    --heading-gradient-to: #a1a1a6;
    --login-shadow: 0 0 80px rgba(10,132,255,0.06);
    --card-shadow: none;
    --logout-hover-border: rgba(255,255,255,0.25);
  }
}

body {
  font-family: -apple-system, BlinkMacSystemFont, 'SF Pro Text', 'Helvetica Neue', sans-serif;
  background: var(--bg);
  color: var(--text);
  line-height: 1.5;
  -webkit-font-smoothing: antialiased;
}

/* ── Top Bar ── */
.top-bar {
  position: sticky;
  top: 0;
  z-index: 100;
  background: var(--nav-bg);
  backdrop-filter: saturate(180%) blur(20px);
  -webkit-backdrop-filter: saturate(180%) blur(20px);
  border-bottom: 0.5px solid var(--border);
}

.top-inner {
  max-width: 1000px;
  margin: 0 auto;
  padding: 0 2rem;
  height: 48px;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.top-brand {
  display: flex;
  align-items: center;
  gap: 0.55rem;
  font-weight: 600;
  font-size: 0.95rem;
  color: var(--text);
  text-decoration: none;
  letter-spacing: -0.01em;
}

.top-brand svg { width: 26px; height: 26px; border-radius: 6px; }

.top-nav {
  display: flex;
  align-items: center;
  gap: 0.15rem;
  list-style: none;
  background: var(--bg-elevated);
  padding: 3px;
  border-radius: 10px;
  border: 0.5px solid var(--border);
}

.top-nav a {
  padding: 0.3rem 0.75rem;
  border-radius: 7px;
  font-size: 0.8rem;
  font-weight: 500;
  color: var(--text-secondary);
  text-decoration: none;
  transition: all 0.2s cubic-bezier(0.25, 0.1, 0.25, 1);
}

.top-nav a:hover { color: var(--text); }

.top-nav a.active {
  color: var(--text);
  background: var(--nav-active-bg);
  box-shadow: var(--nav-active-shadow);
}

.top-right {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.btn-logout {
  font-size: 0.78rem;
  padding: 0.3rem 0.7rem;
  border-radius: 7px;
  border: 0.5px solid var(--border-strong);
  background: transparent;
  color: var(--text-secondary);
  cursor: pointer;
  font-weight: 500;
  transition: all 0.2s cubic-bezier(0.25, 0.1, 0.25, 1);
}

.btn-logout:hover { color: var(--text); border-color: var(--logout-hover-border); }

/* ── Container ── */
.container {
  max-width: 1000px;
  margin: 0 auto;
  padding: 2rem;
}

/* ── Page Header ── */
.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 1.5rem;
}

.page-header h1 {
  font-size: 1.65rem;
  font-weight: 700;
  letter-spacing: -0.03em;
  background: linear-gradient(135deg, var(--heading-gradient-from), var(--heading-gradient-to));
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

/* ── Cards ── */
.card {
  background: var(--bg-elevated);
  border: 0.5px solid var(--border);
  border-radius: 12px;
  overflow: hidden;
  box-shadow: var(--card-shadow);
}

/* ── Buttons ── */
.btn {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.45rem 0.95rem;
  border-radius: 100px;
  font-size: 0.8rem;
  font-weight: 500;
  border: none;
  cursor: pointer;
  transition: all 0.2s cubic-bezier(0.25, 0.1, 0.25, 1);
  text-decoration: none;
}

.btn-primary {
  background: var(--accent);
  color: #fff;
  box-shadow: 0 0 20px var(--accent-glow);
}

.btn-primary:hover { filter: brightness(1.15); transform: scale(1.02); }

.btn-secondary {
  background: var(--bg-secondary);
  color: var(--text);
  border: 0.5px solid var(--border-strong);
}

.btn-secondary:hover { background: var(--bg-tertiary); }

.btn-danger {
  background: var(--red-bg);
  color: var(--red);
  border: 0.5px solid rgba(215,0,21,0.15);
}

.btn-danger:hover { background: var(--red-bg); filter: brightness(0.92); }

.btn-sm { padding: 0.3rem 0.65rem; font-size: 0.76rem; }

/* ── Table ── */
table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.86rem;
}

thead th {
  text-align: left;
  padding: 0.65rem 1.15rem;
  font-weight: 500;
  font-size: 0.74rem;
  color: var(--text-tertiary);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  border-bottom: 0.5px solid var(--border);
}

tbody td {
  padding: 0.8rem 1.15rem;
  border-bottom: 0.5px solid var(--border);
  vertical-align: middle;
}

tbody tr:last-child td { border-bottom: none; }
tbody tr { transition: background 0.2s cubic-bezier(0.25, 0.1, 0.25, 1); }
tbody tr:hover { background: var(--row-hover); }

/* ── Status Pills ── */
.pill {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
  padding: 0.18rem 0.55rem;
  border-radius: 100px;
  font-size: 0.72rem;
  font-weight: 600;
  letter-spacing: 0.02em;
  text-transform: uppercase;
}

.pill::before {
  content: '';
  width: 5px;
  height: 5px;
  border-radius: 50%;
  box-shadow: 0 0 6px currentColor;
}

.pill-approved { background: var(--green-bg); color: var(--green); }
.pill-approved::before { background: var(--green); }

.pill-enrolled { background: var(--green-bg); color: var(--green); }
.pill-enrolled::before { background: var(--green); }

.pill-pending { background: var(--yellow-bg); color: var(--yellow); }
.pill-pending::before { background: var(--yellow); }

.pill-revoked { background: var(--red-bg); color: var(--red); }
.pill-revoked::before { background: var(--red); }

.pill-active { background: var(--green-bg); color: var(--green); }
.pill-active::before { background: var(--green); }

.pill-used { background: var(--teal-bg); color: var(--teal); }
.pill-used::before { background: var(--teal); }

.pill-expired { background: var(--expired-bg); color: var(--text-tertiary); }
.pill-expired::before { background: var(--text-tertiary); }

/* ── Device Row ── */
.device-cell {
  display: flex;
  align-items: center;
  gap: 0.65rem;
}

.device-avatar {
  width: 36px;
  height: 36px;
  border-radius: 8px;
  background: linear-gradient(135deg, var(--bg-secondary), var(--bg-tertiary));
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  border: 0.5px solid var(--border);
}

.device-avatar svg { width: 16px; height: 16px; color: var(--text-secondary); }

.device-info-name { font-weight: 500; font-size: 0.88rem; }
.device-info-fp { font-family: 'SF Mono', Menlo, monospace; font-size: 0.74rem; color: var(--text-tertiary); margin-top: 0.1rem; }

/* ── Code display ── */
code {
  font-family: 'SF Mono', SFMono-Regular, Menlo, monospace;
  font-size: 0.8rem;
  color: var(--text-secondary);
}

.token-chip {
  font-family: 'SF Mono', Menlo, monospace;
  font-size: 0.8rem;
  background: var(--bg-secondary);
  padding: 0.25rem 0.55rem;
  border-radius: 6px;
  border: 0.5px solid var(--border);
  color: var(--text);
  user-select: all;
  display: inline-block;
}

.cmd-block {
  font-family: 'SF Mono', Menlo, monospace;
  font-size: 0.76rem;
  background: var(--bg-secondary);
  padding: 0.35rem 0.6rem;
  border-radius: 6px;
  border: 0.5px solid var(--border);
  color: var(--text);
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
}

.copy-btn {
  background: none;
  border: none;
  cursor: pointer;
  padding: 0.2rem;
  color: var(--text-tertiary);
  border-radius: 4px;
  transition: all 0.2s cubic-bezier(0.25, 0.1, 0.25, 1);
  display: inline-flex;
}

.copy-btn:hover { color: var(--accent); }

/* ── Stale ── */
.stale { opacity: 0.45; }
.stale-dot {
  display: inline-block;
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--red);
  margin-left: 0.35rem;
  box-shadow: 0 0 6px var(--red);
  vertical-align: middle;
}

/* ── Actions ── */
.actions { display: flex; gap: 0.4rem; align-items: center; }
.actions form { display: inline; }

/* ── Section ── */
.section-gap { margin-top: 2.5rem; }

/* ── Timestamp ── */
.ts { color: var(--text-tertiary); font-size: 0.82rem; }

/* ── Awaiting verification ── */
.awaiting { color: var(--text-secondary); font-size: 0.82rem; }

/* ── Stats Bar ── */
.stats-bar {
  display: flex;
  gap: 1rem;
  margin-bottom: 1.5rem;
}

.stat-card {
  flex: 1;
  background: var(--bg-elevated);
  border: 0.5px solid var(--border);
  border-radius: 12px;
  padding: 1rem 1.15rem;
  box-shadow: var(--card-shadow);
}

.stat-label {
  font-size: 0.72rem;
  font-weight: 500;
  color: var(--text-tertiary);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.stat-value {
  font-size: 1.6rem;
  font-weight: 700;
  letter-spacing: -0.03em;
  margin-top: 0.2rem;
}

.stat-value.green { color: var(--green); }
.stat-value.yellow { color: var(--yellow); }
.stat-value.red { color: var(--red); }

/* ── Login ── */
.login-wrap {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 5rem 0;
}

.login-card {
  width: 100%;
  max-width: 360px;
  background: var(--bg-elevated);
  border: 0.5px solid var(--border);
  border-radius: 16px;
  padding: 2.5rem 2rem;
  text-align: center;
  box-shadow: var(--login-shadow);
}

.login-card h2 {
  font-size: 1.4rem;
  font-weight: 700;
  letter-spacing: -0.025em;
  margin-top: 1rem;
  margin-bottom: 0.2rem;
  background: linear-gradient(135deg, var(--heading-gradient-from), var(--heading-gradient-to));
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.login-card .sub { font-size: 0.85rem; color: var(--text-secondary); margin-bottom: 1.75rem; }

.login-card label {
  display: block;
  text-align: left;
  font-size: 0.78rem;
  font-weight: 500;
  color: var(--text-secondary);
  margin-bottom: 0.35rem;
}

.login-card input {
  width: 100%;
  padding: 0.65rem 0.9rem;
  border: 0.5px solid var(--border-strong);
  border-radius: 8px;
  font-size: 0.88rem;
  font-family: inherit;
  background: var(--bg-secondary);
  color: var(--text);
  outline: none;
  transition: border-color 0.2s cubic-bezier(0.25, 0.1, 0.25, 1), box-shadow 0.2s cubic-bezier(0.25, 0.1, 0.25, 1);
}

.login-card input:focus {
  border-color: var(--accent);
  box-shadow: 0 0 0 3px var(--accent-glow);
}

.login-card .btn-primary {
  width: 100%;
  justify-content: center;
  padding: 0.65rem;
  font-size: 0.88rem;
  margin-top: 1.25rem;
  border-radius: 8px;
}

.login-card .error { color: var(--red); font-size: 0.85rem; margin-bottom: 1rem; }

/* ── Confirm Dialog ── */
.confirm-dialog { background-color: rgba(0,0,0,0.25); backdrop-filter: blur(2px); -webkit-backdrop-filter: blur(2px); border: none; border-radius: 16px; padding: 0; }
.confirm-dialog::backdrop { background: rgba(0,0,0,0.25); backdrop-filter: blur(2px); -webkit-backdrop-filter: blur(2px); }
.confirm-dialog .dialog-body { background: var(--bg-elevated); border: 0.5px solid var(--border); border-radius: 16px; padding: 1.5em; max-width: 360px; }
.confirm-dialog h4 { margin: 0 0 0.5em; font-size: 1.05em; font-weight: 600; color: var(--text); }
.confirm-dialog p { margin: 0 0 1.5em; font-size: 0.9em; color: var(--text-secondary); line-height: 1.5; }
.confirm-dialog .dialog-actions { display: flex; gap: 0.5em; justify-content: flex-end; }
```

- [ ] **Step 2: Verify the file is valid CSS**

Run: `cat internal/hub/templates/theme.css | head -5`
Expected: The reset and `:root` opening are present.

- [ ] **Step 3: Commit**

```bash
git add internal/hub/templates/theme.css
git commit -m "feat: add theme.css with dual light/dark theme"
```

---

### Task 2: Update server.go — route, template funcs, dashboard data

**Files:**
- Modify: `internal/hub/server.go`

- [ ] **Step 1: Add `eventPillClass` to funcMap**

In `NewServer`, add to the `funcMap` (after the `"upper"` entry):

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

- [ ] **Step 2: Add `countByStatus` helper**

Add this function anywhere in `server.go` (e.g. after `deviceFromContext`):

```go
// countByStatus returns the number of devices with the given status.
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

- [ ] **Step 3: Update `handleDashboard` to pass counts**

Change the `renderTemplate` call in `handleDashboard` from:

```go
s.renderTemplate(w, "devices.html", map[string]any{
    "Devices": devices,
})
```

to:

```go
s.renderTemplate(w, "devices.html", map[string]any{
    "Devices":       devices,
    "TotalCount":    len(devices),
    "ApprovedCount": countByStatus(devices, model.StatusApproved),
    "PendingCount":  countByStatus(devices, model.StatusPending),
    "RevokedCount":  countByStatus(devices, model.StatusRevoked),
})
```

- [ ] **Step 4: Rename CSS route and ReadFile path**

Change line 127:
```go
s.mux.HandleFunc("/static/pico.min.css", s.handleStaticCSS)
```
to:
```go
s.mux.HandleFunc("/static/theme.css", s.handleStaticCSS)
```

Change line 244 inside `handleStaticCSS`:
```go
data, err := templateFS.ReadFile("templates/pico.min.css")
```
to:
```go
data, err := templateFS.ReadFile("templates/theme.css")
```

- [ ] **Step 5: Verify it compiles**

Run: `go vet ./internal/hub/...`
Expected: No errors.

- [ ] **Step 6: Commit**

```bash
git add internal/hub/server.go
git commit -m "feat: update server.go for theme.css route and new template funcs"
```

---

### Task 3: Delete pico.min.css

**Files:**
- Delete: `internal/hub/templates/pico.min.css`

- [ ] **Step 1: Remove the file via git**

```bash
git rm internal/hub/templates/pico.min.css
```

- [ ] **Step 2: Verify embed still works**

Run: `go build ./cmd/ssh-vault`
Expected: Builds successfully (no embed errors since `//go:embed templates/*` is a glob).

- [ ] **Step 3: Commit**

```bash
git commit -m "chore: remove pico.min.css — replaced by theme.css"
```

---

### Task 4: Rewrite layout.html

**Files:**
- Rewrite: `internal/hub/templates/layout.html`

- [ ] **Step 1: Replace layout.html contents**

```html
{{define "layout"}}
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <meta name="color-scheme" content="light dark">
    <title>SSH Vault — {{template "title" .}}</title>
    <link rel="icon" type="image/svg+xml" href="/static/logo.svg">
    <link rel="stylesheet" href="/static/theme.css">
</head>
<body>
    <nav class="top-bar">
        <div class="top-inner">
            <a class="top-brand" href="/">
                <img src="/static/logo.svg" alt="SSH Vault" style="width:26px;height:26px;border-radius:6px">
                SSH Vault
            </a>
            <ul class="top-nav">
                <li><a href="/"{{if eq (template "title" .) "Devices"}} class="active"{{end}}>Devices</a></li>
                <li><a href="/tokens"{{if eq (template "title" .) "Tokens"}} class="active"{{end}}>Tokens</a></li>
                <li><a href="/audit"{{if eq (template "title" .) "Audit Log"}} class="active"{{end}}>Audit Log</a></li>
            </ul>
            <div class="top-right">
                <form action="/logout" method="post" style="margin:0">
                    <button type="submit" class="btn-logout">Log out</button>
                </form>
            </div>
        </div>
    </nav>
    <div class="container">
        {{template "content" .}}
    </div>
    <script>
    document.querySelectorAll("td, small").forEach(function(el) {
        var text = el.textContent.trim();
        if (/^\d{4}-\d{2}-\d{2}T/.test(text)) {
            var d = new Date(text);
            if (!isNaN(d)) {
                el.textContent = d.toLocaleString(undefined, {
                    year: "numeric", month: "2-digit", day: "2-digit",
                    hour: "2-digit", minute: "2-digit"
                });
            }
        }
    });
    </script>
</body>
</html>
{{end}}
```

**Note:** The `{{if eq (template "title" .) "..."}}` pattern for active nav will NOT work in Go templates — `template` is an action, not a function. Instead, use a simpler approach: remove the active class from layout entirely and add a `<style>` override per page (or just leave all links unstyled — the URL bar shows which page you're on). Alternatively, pass an `ActivePage` string from each handler.

**Better approach — pass `ActivePage` from handlers:** Each handler already passes a map. Add `"ActivePage": "devices"` (or `"tokens"`, `"audit"`) to each handler's template data. Then in layout:

```html
<li><a href="/"{{if eq .ActivePage "devices"}} class="active"{{end}}>Devices</a></li>
<li><a href="/tokens"{{if eq .ActivePage "tokens"}} class="active"{{end}}>Tokens</a></li>
<li><a href="/audit"{{if eq .ActivePage "audit"}} class="active"{{end}}>Audit Log</a></li>
```

This requires adding `"ActivePage": "devices"` to `handleDashboard`, `"ActivePage": "tokens"` to `handleTokens`, `"ActivePage": "audit"` to `handleAudit`. The login page does not use the nav layout so it doesn't need this.

Update `layout.html` to use `.ActivePage`:

```html
{{define "layout"}}
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <meta name="color-scheme" content="light dark">
    <title>SSH Vault — {{template "title" .}}</title>
    <link rel="icon" type="image/svg+xml" href="/static/logo.svg">
    <link rel="stylesheet" href="/static/theme.css">
</head>
<body>
    {{if .ActivePage}}
    <nav class="top-bar">
        <div class="top-inner">
            <a class="top-brand" href="/">
                <img src="/static/logo.svg" alt="SSH Vault" style="width:26px;height:26px;border-radius:6px">
                SSH Vault
            </a>
            <ul class="top-nav">
                <li><a href="/"{{if eq .ActivePage "devices"}} class="active"{{end}}>Devices</a></li>
                <li><a href="/tokens"{{if eq .ActivePage "tokens"}} class="active"{{end}}>Tokens</a></li>
                <li><a href="/audit"{{if eq .ActivePage "audit"}} class="active"{{end}}>Audit Log</a></li>
            </ul>
            <div class="top-right">
                <form action="/logout" method="post" style="margin:0">
                    <button type="submit" class="btn-logout">Log out</button>
                </form>
            </div>
        </div>
    </nav>
    {{end}}
    <div class="container">
        {{template "content" .}}
    </div>
    <script>
    document.querySelectorAll("td, small").forEach(function(el) {
        var text = el.textContent.trim();
        if (/^\d{4}-\d{2}-\d{2}T/.test(text)) {
            var d = new Date(text);
            if (!isNaN(d)) {
                el.textContent = d.toLocaleString(undefined, {
                    year: "numeric", month: "2-digit", day: "2-digit",
                    hour: "2-digit", minute: "2-digit"
                });
            }
        }
    });
    </script>
</body>
</html>
{{end}}
```

- [ ] **Step 2: Add `ActivePage` to all dashboard handlers in server.go**

In `handleDashboard`:
```go
"ActivePage": "devices",
```

In `handleTokens` (GET branch):
```go
"ActivePage": "tokens",
```

In `handleAudit`:
```go
"ActivePage": "audit",
```

The login handler passes `map[string]any{"Error": true}` or `nil` — no `ActivePage` needed. The `{{if .ActivePage}}` guard in layout hides the nav for login.

**Important:** `handleLogin` renders with `map[string]any{"Error": true}` on failure and `nil` on GET. Since `.ActivePage` on a nil map will be falsy, the nav is hidden. But on error, `map[string]any{"Error": true}` has no `ActivePage` key, so `.ActivePage` returns empty string which is falsy — nav is hidden. This works correctly.

- [ ] **Step 3: Verify it compiles**

Run: `go vet ./internal/hub/...`
Expected: No errors.

- [ ] **Step 4: Commit**

```bash
git add internal/hub/templates/layout.html internal/hub/server.go
git commit -m "feat: rewrite layout.html with new nav and theme.css link"
```

---

### Task 5: Rewrite devices.html

**Files:**
- Rewrite: `internal/hub/templates/devices.html`

- [ ] **Step 1: Replace devices.html contents**

```html
{{define "title"}}Devices{{end}}
{{define "content"}}
<div class="page-header">
    <h1>Devices</h1>
</div>

<div class="stats-bar">
    <div class="stat-card">
        <div class="stat-label">Total</div>
        <div class="stat-value">{{.TotalCount}}</div>
    </div>
    <div class="stat-card">
        <div class="stat-label">Approved</div>
        <div class="stat-value green">{{.ApprovedCount}}</div>
    </div>
    <div class="stat-card">
        <div class="stat-label">Pending</div>
        <div class="stat-value yellow">{{.PendingCount}}</div>
    </div>
    <div class="stat-card">
        <div class="stat-label">Revoked</div>
        <div class="stat-value red">{{.RevokedCount}}</div>
    </div>
</div>

{{if .Devices}}
<div class="card">
    <table>
        <thead>
            <tr>
                <th>Device</th>
                <th>Status</th>
                <th>Last Sync</th>
                <th>Actions</th>
            </tr>
        </thead>
        <tbody>
            {{range .Devices}}
            <tr{{if isStale .}} class="stale"{{end}}>
                <td>
                    <div class="device-cell">
                        <div class="device-avatar">
                            <svg fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5"><rect x="2" y="3" width="20" height="14" rx="2"/><path d="M8 21h8m-4-4v4"/></svg>
                        </div>
                        <div>
                            <div class="device-info-name">{{.Name}}{{if isStale .}}<span class="stale-dot"></span>{{end}}</div>
                            <div class="device-info-fp">{{formatFingerprint .Fingerprint}}</div>
                        </div>
                    </div>
                </td>
                <td><span class="pill pill-{{.Status}}">{{upper .Status}}</span></td>
                <td class="ts">{{formatTimePtr .LastSyncAt}}</td>
                <td class="actions">
                    {{if eq .Status "pending"}}
                        {{if .Verified}}
                        <form method="post" action="/devices/{{.ID}}/approve">
                            <button type="submit" class="btn btn-primary btn-sm">Approve</button>
                        </form>
                        {{else}}
                        <span class="awaiting">Awaiting verification</span>
                        {{end}}
                    {{else if eq .Status "approved"}}
                        <form method="post" action="/devices/{{.ID}}/revoke"
                              onsubmit="return confirm('Revoke device {{.Name}}?')">
                            <button type="submit" class="btn btn-danger btn-sm">Revoke</button>
                        </form>
                    {{else if eq .Status "revoked"}}
                        <form method="post" action="/devices/{{.ID}}/remove"
                              onsubmit="return confirm('Remove device {{.Name}}?')">
                            <button type="submit" class="btn btn-danger btn-sm">Remove</button>
                        </form>
                    {{end}}
                </td>
            </tr>
            {{end}}
        </tbody>
    </table>
</div>
{{else}}
<p style="color:var(--text-secondary)">No devices enrolled yet. Generate a token and enroll your first device.</p>
{{end}}
{{end}}
{{template "layout" .}}
```

- [ ] **Step 2: Verify it compiles**

Run: `go build ./cmd/ssh-vault`
Expected: Builds successfully.

- [ ] **Step 3: Commit**

```bash
git add internal/hub/templates/devices.html
git commit -m "feat: rewrite devices.html with stats bar and new design"
```

---

### Task 6: Rewrite tokens.html

**Files:**
- Rewrite: `internal/hub/templates/tokens.html`

- [ ] **Step 1: Replace tokens.html contents**

```html
{{define "title"}}Tokens{{end}}
{{define "content"}}
{{if .ExternalURL}}
<div class="page-header">
    <h1>Quick Enrollment</h1>
    <form method="post" action="/tokens/generate-link" style="margin:0">
        <button type="submit" class="btn btn-primary">
            <svg width="13" height="13" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5"><path d="M12 5v14m-7-7h14"/></svg>
            Generate Link
        </button>
    </form>
</div>

{{if .ShortCodes}}
<div class="card">
    <table>
        <thead>
            <tr>
                <th>Code</th>
                <th>Command</th>
                <th>Status</th>
                <th>Expires</th>
            </tr>
        </thead>
        <tbody>
            {{range .ShortCodes}}
            <tr>
                <td><strong style="letter-spacing:0.05em">{{.Code}}</strong></td>
                <td>
                    <span class="cmd-block">
                        <code class="enroll-cmd" id="cmd-{{.Code}}">curl -sSL {{$.ExternalURL}}/e/{{.Code}} | sh</code>
                        <button class="copy-btn" onclick="copyCmd('cmd-{{.Code}}')" title="Copy command">
                            <svg width="13" height="13" viewBox="0 0 16 16" fill="currentColor"><path d="M0 6.75C0 5.784.784 5 1.75 5h1.5a.75.75 0 010 1.5h-1.5a.25.25 0 00-.25.25v7.5c0 .138.112.25.25.25h7.5a.25.25 0 00.25-.25v-1.5a.75.75 0 011.5 0v1.5A1.75 1.75 0 019.25 16h-7.5A1.75 1.75 0 010 14.25v-7.5z"/><path d="M5 1.75C5 .784 5.784 0 6.75 0h7.5C15.216 0 16 .784 16 1.75v7.5A1.75 1.75 0 0114.25 11h-7.5A1.75 1.75 0 015 9.25v-7.5zm1.75-.25a.25.25 0 00-.25.25v7.5c0 .138.112.25.25.25h7.5a.25.25 0 00.25-.25v-7.5a.25.25 0 00-.25-.25h-7.5z"/></svg>
                        </button>
                    </span>
                </td>
                <td>
                    {{if .Used}}<span class="pill pill-used">used</span>
                    {{else if .IsExpired}}<span class="pill pill-expired">expired</span>
                    {{else}}<span class="pill pill-active">active</span>
                    {{end}}
                </td>
                <td class="ts">{{formatTime .ExpiresAt}}</td>
            </tr>
            {{end}}
        </tbody>
    </table>
</div>
{{else}}
<p style="color:var(--text-secondary)">No enrollment links yet. Generate one to enroll a new device with a single command.</p>
{{end}}
{{end}}

<div class="section-gap">
    <div class="page-header">
        <h1>Onboarding Tokens</h1>
        <form method="post" action="/tokens" style="margin:0">
            <button type="submit" class="btn btn-secondary">
                <svg width="13" height="13" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5"><path d="M12 5v14m-7-7h14"/></svg>
                Generate Token
            </button>
        </form>
    </div>

    {{if .Tokens}}
    <div class="card">
        <table>
            <thead>
                <tr>
                    <th>Token</th>
                    <th>Expires</th>
                    <th>Actions</th>
                </tr>
            </thead>
            <tbody>
                {{range .Tokens}}
                <tr>
                    <td>
                        <span class="token-chip">{{.Value}}</span>
                        <button class="copy-btn" onclick="copyToken(this, '{{.Value}}')" title="Copy token">
                            <svg width="13" height="13" viewBox="0 0 16 16" fill="currentColor"><path d="M0 6.75C0 5.784.784 5 1.75 5h1.5a.75.75 0 010 1.5h-1.5a.25.25 0 00-.25.25v7.5c0 .138.112.25.25.25h7.5a.25.25 0 00.25-.25v-1.5a.75.75 0 011.5 0v1.5A1.75 1.75 0 019.25 16h-7.5A1.75 1.75 0 010 14.25v-7.5z"/><path d="M5 1.75C5 .784 5.784 0 6.75 0h7.5C15.216 0 16 .784 16 1.75v7.5A1.75 1.75 0 0114.25 11h-7.5A1.75 1.75 0 015 9.25v-7.5zm1.75-.25a.25.25 0 00-.25.25v7.5c0 .138.112.25.25.25h7.5a.25.25 0 00.25-.25v-7.5a.25.25 0 00-.25-.25h-7.5z"/></svg>
                        </button>
                    </td>
                    <td class="ts">{{formatTime .ExpiresAt}}</td>
                    <td class="actions">
                        <button class="btn btn-danger btn-sm" onclick="confirmRemove('{{.Value}}')">Remove</button>
                    </td>
                </tr>
                {{end}}
            </tbody>
        </table>
    </div>

    <dialog id="removeDialog" class="confirm-dialog">
        <div class="dialog-body">
            <h4>Remove token?</h4>
            <p>This token will be permanently removed. Devices that haven't enrolled yet won't be able to use it.</p>
            <div class="dialog-actions">
                <button class="btn btn-secondary btn-sm" onclick="document.getElementById('removeDialog').close()">Cancel</button>
                <form id="removeForm" method="post" style="margin:0">
                    <button type="submit" class="btn btn-danger btn-sm">Remove</button>
                </form>
            </div>
        </div>
    </dialog>

    <script>
    function confirmRemove(tokenValue) {
        document.getElementById('removeForm').action = '/tokens/' + tokenValue + '/remove';
        document.getElementById('removeDialog').showModal();
    }

    function copyToken(btn, value) {
        navigator.clipboard.writeText(value).then(function() {
            btn.innerHTML = '&#10003;';
            btn.title = 'Copied!';
            setTimeout(function() {
                btn.innerHTML = '<svg width="13" height="13" viewBox="0 0 16 16" fill="currentColor"><path d="M0 6.75C0 5.784.784 5 1.75 5h1.5a.75.75 0 010 1.5h-1.5a.25.25 0 00-.25.25v7.5c0 .138.112.25.25.25h7.5a.25.25 0 00.25-.25v-1.5a.75.75 0 011.5 0v1.5A1.75 1.75 0 019.25 16h-7.5A1.75 1.75 0 010 14.25v-7.5z"/><path d="M5 1.75C5 .784 5.784 0 6.75 0h7.5C15.216 0 16 .784 16 1.75v7.5A1.75 1.75 0 0114.25 11h-7.5A1.75 1.75 0 015 9.25v-7.5zm1.75-.25a.25.25 0 00-.25.25v7.5c0 .138.112.25.25.25h7.5a.25.25 0 00.25-.25v-7.5a.25.25 0 00-.25-.25h-7.5z"/></svg>';
                btn.title = 'Copy token';
            }, 1500);
        }).catch(function() {
            alert('Clipboard not available. Please select and copy the token manually.');
        });
    }
    </script>
    {{else}}
    <p style="color:var(--text-secondary)">No active tokens. Generate one to enroll a new device.</p>
    {{end}}
</div>

<script>
function copyCmd(elementId) {
    var el = document.getElementById(elementId);
    navigator.clipboard.writeText(el.textContent).then(function() {
        var btn = el.nextElementSibling;
        btn.innerHTML = '&#10003;';
        btn.title = 'Copied!';
        setTimeout(function() {
            btn.innerHTML = '<svg width="13" height="13" viewBox="0 0 16 16" fill="currentColor"><path d="M0 6.75C0 5.784.784 5 1.75 5h1.5a.75.75 0 010 1.5h-1.5a.25.25 0 00-.25.25v7.5c0 .138.112.25.25.25h7.5a.25.25 0 00.25-.25v-1.5a.75.75 0 011.5 0v1.5A1.75 1.75 0 019.25 16h-7.5A1.75 1.75 0 010 14.25v-7.5z"/><path d="M5 1.75C5 .784 5.784 0 6.75 0h7.5C15.216 0 16 .784 16 1.75v7.5A1.75 1.75 0 0114.25 11h-7.5A1.75 1.75 0 015 9.25v-7.5zm1.75-.25a.25.25 0 00-.25.25v7.5c0 .138.112.25.25.25h7.5a.25.25 0 00.25-.25v-7.5a.25.25 0 00-.25-.25h-7.5z"/></svg>';
            btn.title = 'Copy command';
        }, 1500);
    });
}
</script>
{{end}}
{{template "layout" .}}
```

- [ ] **Step 2: Verify it compiles**

Run: `go build ./cmd/ssh-vault`
Expected: Builds successfully.

- [ ] **Step 3: Commit**

```bash
git add internal/hub/templates/tokens.html
git commit -m "feat: rewrite tokens.html with new design"
```

---

### Task 7: Rewrite audit.html

**Files:**
- Rewrite: `internal/hub/templates/audit.html`

- [ ] **Step 1: Replace audit.html contents**

```html
{{define "title"}}Audit Log{{end}}
{{define "content"}}
<div class="page-header">
    <h1>Audit Log</h1>
</div>

{{if .Entries}}
<div class="card">
    <table>
        <thead>
            <tr>
                <th>Time</th>
                <th>Event</th>
                <th>Device</th>
                <th>Details</th>
            </tr>
        </thead>
        <tbody>
            {{range .Entries}}
            <tr>
                <td class="ts">{{formatTime .Timestamp}}</td>
                <td><span class="pill pill-{{eventPillClass .Event}}">{{upper .Event}}</span></td>
                <td><code>{{.DeviceID}}</code></td>
                <td style="color:var(--text-secondary)">{{.Details}}</td>
            </tr>
            {{end}}
        </tbody>
    </table>
</div>
{{else}}
<p style="color:var(--text-secondary)">No audit entries yet.</p>
{{end}}
{{end}}
{{template "layout" .}}
```

- [ ] **Step 2: Verify it compiles**

Run: `go build ./cmd/ssh-vault`
Expected: Builds successfully.

- [ ] **Step 3: Commit**

```bash
git add internal/hub/templates/audit.html
git commit -m "feat: rewrite audit.html with pill badges and eventPillClass"
```

---

### Task 8: Rewrite login.html

**Files:**
- Rewrite: `internal/hub/templates/login.html`

- [ ] **Step 1: Replace login.html contents**

```html
{{define "title"}}Login{{end}}
{{define "content"}}
<div class="login-wrap">
    <div class="login-card">
        <img src="/static/logo.svg" alt="SSH Vault" style="width:52px;height:52px;border-radius:13px">
        <h2>SSH Vault</h2>
        <p class="sub">Enter your password to continue</p>
        {{if .Error}}
        <p class="error">Invalid password. Please try again.</p>
        {{end}}
        <form method="post" action="/login">
            <label for="password">Password</label>
            <input type="password" id="password" name="password" required autofocus placeholder="Dashboard password">
            <button type="submit" class="btn btn-primary">Log In</button>
        </form>
    </div>
</div>
{{end}}
{{template "layout" .}}
```

- [ ] **Step 2: Verify it compiles**

Run: `go build ./cmd/ssh-vault`
Expected: Builds successfully.

- [ ] **Step 3: Commit**

```bash
git add internal/hub/templates/login.html
git commit -m "feat: rewrite login.html with centered card and gradient heading"
```

---

### Task 9: Final verification

- [ ] **Step 1: Run full test suite**

Run: `go test ./...`
Expected: All tests pass.

- [ ] **Step 2: Run vet and build**

Run: `go vet ./... && go build -o ssh-vault ./cmd/ssh-vault`
Expected: No errors, binary produced.

- [ ] **Step 3: Visual verification note**

Start the server and verify in browser:
- Devices page: stats bar, device avatars with fingerprint, pill badges, stale dot
- Tokens page: cmd-block with copy button, token chips, confirm dialog
- Audit page: pill badges with correct event-to-pill mapping
- Login page: centered card with gradient heading, focus ring on input
- Toggle OS dark/light mode: both themes render correctly
- Nav: active page highlighted in segmented control

This step is manual — run `./ssh-vault hub --password test` and open in browser.

- [ ] **Step 4: Clean up design files (optional)**

The `designs/` directory contains the prototypes. These can be kept for reference or removed:

```bash
git rm -r designs/
git commit -m "chore: remove design prototypes — implemented in templates"
```
