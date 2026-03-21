# Quickstart: Token Management Enhancements

**Feature**: 002-token-management | **Date**: 2026-03-21

## Prerequisites

- Go 1.26+ installed
- Repository cloned and on `002-token-management` branch

## Build & Run

```bash
# Build
go build -o ssh-vault ./cmd/ssh-vault

# Run hub (default port 8443)
./ssh-vault hub --password mysecret --data ./data/data.json
```

## Verify New Features

### 1. Token Removal

1. Log in to the dashboard at `https://localhost:8443`
2. Navigate to **Tokens** page
3. Click **Generate Token** to create a new token
4. Click the **Remove** button next to the token
5. Confirm the removal in the browser dialog
6. Verify the token is gone from the list

### 2. Token Usage Audit

1. Generate a new token on the Tokens page
2. Copy the token value using the copy icon
3. Enroll a device using the token:
   ```bash
   ./ssh-vault agent enroll --hub https://localhost:8443 --token <paste-token>
   ```
4. Navigate to **Audit Log** page
5. Verify a `TOKEN_USED` event appears with the truncated token prefix

### 3. Copy to Clipboard

1. Navigate to **Tokens** page with at least one active token
2. Click the clipboard icon next to a token
3. Verify visual feedback (checkmark appears briefly)
4. Paste into a text editor — the full 64-character token value should match

### 4. Expired Token Purge

1. Tokens with past expiry times are automatically cleaned up when viewing the Tokens page
2. No manual action required — verify by checking that the data file does not accumulate expired tokens

## Run Tests

```bash
# All tests
go test ./...

# Store tests only (includes new RemoveToken/PurgeExpiredTokens tests)
go test ./internal/hub/ -run TestFileStore

# Handler tests only (includes new token removal endpoint tests)
go test ./internal/hub/ -run TestHandle

# With verbose output
go test -v ./internal/hub/...
```

## Key Files Changed

| File | Change |
|------|--------|
| `internal/model/audit.go` | New event constants: `EventTokenUsed`, `EventTokenRemoved` |
| `internal/hub/store.go` | New methods: `RemoveToken()`, `PurgeExpiredTokens()` |
| `internal/hub/handlers.go` | New handler: `handleRemoveToken()`; audit entry in `handleEnroll` |
| `internal/hub/server.go` | New route registration; purge call in token list handler |
| `internal/hub/templates/tokens.html` | Remove button + copy icon per token row |
| `internal/hub/templates/layout.html` | CSS for new badge types and copy button |
| `internal/hub/store_test.go` | Tests for `RemoveToken`, `PurgeExpiredTokens` |
| `internal/hub/handlers_test.go` | Tests for token removal endpoint and audit entries |
