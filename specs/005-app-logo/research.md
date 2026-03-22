# Research: App Logo Integration

## Phase 0 Research

No NEEDS CLARIFICATION items existed in the Technical Context. Research focused on confirming implementation patterns.

### 1. SVG Favicon Browser Support

**Decision**: Use SVG favicon with `<link rel="icon" type="image/svg+xml">`
**Rationale**: Supported by Chrome 80+, Firefox 41+, Safari 16+, Edge 80+ — covers ~97% of web users. No `.ico` conversion needed.
**Alternatives considered**:
- PNG favicon: Requires image conversion tooling, multiple sizes
- .ico file: Legacy format, requires conversion, no benefit for modern browsers
- Base64 inline: Bloats HTML, harder to cache

### 2. Existing Static Asset Pattern

**Decision**: Follow `handleStaticCSS` pattern exactly for the logo handler
**Rationale**: The project already has a proven pattern for serving embedded static files:
1. Read from `templateFS` (embedded filesystem)
2. Set appropriate `Content-Type` header
3. Set `Cache-Control: public, max-age=86400`
4. Write bytes to response

This pattern satisfies the constitution's Simplicity principle.
**Alternatives considered**:
- Generic static file server: Over-engineered for 2 files
- Template function to inline SVG: Works for header but not favicon

### 3. Header Logo Sizing

**Decision**: Use CSS `height` constraint (e.g., `height: 1.5em`) to match nav bar text
**Rationale**: Pico CSS nav elements use relative sizing. Using `em` units ensures the logo scales with text. The SVG's `viewBox` handles aspect ratio automatically.
**Alternatives considered**:
- Fixed pixel size: Doesn't scale with user font preferences
- Width-based sizing: Could distort if aspect ratio changes
