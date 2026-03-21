# Research: Self-Serve Binary Distribution

**Feature**: 004-self-serve-binaries
**Date**: 2026-03-21

## Decision 1: File Serving Approach

**Decision**: Use `http.ServeFile` from the standard library to serve binaries.

**Rationale**: `http.ServeFile` handles Content-Type detection, Content-Length, Range requests (resume support), Last-Modified headers, and conditional requests (If-Modified-Since) automatically. It's the idiomatic Go approach and avoids reimplementing HTTP semantics.

**Alternatives considered**:
- Manual `os.Open` + `io.Copy`: More control but loses range request support and requires manual header management. Unnecessary complexity.
- `http.FileServer`: Designed for serving entire directory trees, which would expose the filesystem structure. Too broad — we only want to serve specific named files.

## Decision 2: Path Traversal Prevention

**Decision**: Validate `{os}` and `{arch}` against a strict allowlist map before constructing the file path. Never use user input directly in path construction.

**Rationale**: An allowlist is the simplest and most secure approach. Since we only support 2 OS values and 2 arch values (4 combinations total), a map lookup is trivial and eliminates any path traversal concern by construction.

**Alternatives considered**:
- `filepath.Clean` + prefix check: Works but is error-prone (subtle edge cases with symlinks, double encoding). Allowlist is simpler and provably safe.
- Serving the entire dist directory via `http.FileServer` with `http.StripPrefix`: Would expose directory listings and any other files placed in the dist dir. Too permissive.

## Decision 3: Checksum File Serving

**Decision**: Serve `checksums.txt` from the dist directory via a dedicated path `GET /download/checksums.txt`, using the same handler with a special case.

**Rationale**: The enrollment script already downloads `checksums.txt` alongside the binary. Changing the checksum URL to point to the hub keeps the entire flow self-contained. The checksums file is placed in the dist directory alongside the binaries by the same external build process.

**Alternatives considered**:
- Separate `/checksums` endpoint: Unnecessary new endpoint; the file lives in the same directory.
- Embedding checksums in the enrollment script: Would make the script huge and require regeneration on every build. Not practical.

## Decision 4: Graceful Degradation Without --dist-dir

**Decision**: The hub starts normally without `--dist-dir`. The download endpoint returns 501 Not Configured. The enrollment script generation checks for `distDir` instead of `githubRepo` in the configuration guard.

**Rationale**: The hub serves many functions beyond enrollment (dashboard, API, key sync). Requiring `--dist-dir` would break existing deployments that don't use quick enrollment. The enrollment script already handles the "binary not available" case by checking if `ssh-vault` is already in PATH.

**Alternatives considered**:
- Require `--dist-dir` at startup: Too restrictive; breaks existing hub-only deployments.
- Silent 404 on download: Less helpful than an explicit "not configured" message for operators debugging setup.

## Decision 5: Download Handler File Organization

**Decision**: Create a new `handlers_download.go` file for the download handler, following the existing pattern where `handlers_shortcode.go` is separate from `handlers.go`.

**Rationale**: The codebase already separates handler groups by file (handlers.go for core routes, handlers_shortcode.go for enrollment). A new file keeps the download logic isolated and easy to find.

**Alternatives considered**:
- Add to `handlers_shortcode.go`: The download endpoint is related to enrollment but is a distinct concern (serving files vs. generating scripts). Separate file is cleaner.
- Add to `handlers.go`: Already handles dashboard/tokens/audit. Adding binary serving would make it a grab-bag.
