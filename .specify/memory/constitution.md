<!--
Sync Impact Report
- Version change: N/A → 1.0.0 (initial ratification)
- Added principles:
  - I. Idiomatic Go
  - II. Testing
  - III. Simplicity
- Added sections:
  - Technology Stack
  - Development Workflow
- Templates requiring updates:
  - plan-template.md ✅ no changes needed (generic Constitution Check gate)
  - spec-template.md ✅ no changes needed (technology-agnostic)
  - tasks-template.md ✅ no changes needed (language-agnostic structure)
- Follow-up TODOs: none
-->

# SSH Vault Constitution

## Core Principles

### I. Idiomatic Go

All code MUST follow standard Go conventions and idioms:
- Use `gofmt`/`goimports` for formatting — no exceptions
- Errors MUST be handled explicitly; never discard errors with `_`
- Prefer the standard library over third-party dependencies
  when the standard library is sufficient
- Follow [Effective Go](https://go.dev/doc/effective_go) and
  the Go Code Review Comments guidelines
- Export only what consumers need; keep the public API surface
  minimal

### II. Testing

All non-trivial code MUST have tests:
- Use the standard `testing` package
- Table-driven tests are the default pattern for multiple cases
- Tests MUST be runnable with `go test ./...` from the project root
- Test files live alongside the code they test (`*_test.go`)

### III. Simplicity

Start with the simplest solution that works:
- YAGNI: do not build for hypothetical future requirements
- Prefer flat package structures; introduce nesting only when
  a clear organizational benefit exists
- Avoid abstractions until duplication or complexity demands them
- One binary, one `main` package unless the project scope
  explicitly requires more

## Technology Stack

- **Language**: Go (latest stable)
- **Build**: `go build` / `go install`
- **Testing**: `go test`
- **Linting**: `go vet` + `golangci-lint`
- **Module management**: Go modules (`go.mod` / `go.sum`)
- **Dependencies**: Minimize; prefer standard library

## Development Workflow

- Feature branches merged via Pull Requests
- All PRs MUST pass `go vet`, `golangci-lint`, and `go test ./...`
- Commits SHOULD be atomic and descriptive
- Always ask before creating commits (per project policy)
- Always rebase before merging to maintain a clean history. Squash before
  merging if necessary.
- Never assume, always ask if unclear.

## Governance

- This constitution supersedes conflicting ad-hoc practices
- Amendments require updating this file, incrementing the version,
  and documenting the change in the Sync Impact Report comment
- Versioning follows semver: MAJOR (principle removal/redefinition),
  MINOR (new principle/section), PATCH (clarifications/typos)
- All plan-level Constitution Checks reference these principles

**Version**: 1.0.0 | **Ratified**: 2026-03-21 | **Last Amended**: 2026-03-21
