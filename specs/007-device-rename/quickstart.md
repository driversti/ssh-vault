# Quickstart: Device Rename

## What to Build

1. **Backend**: Add `POST /devices/{id}/rename` handler that accepts `{"name": "..."}`, validates, saves, audits, and returns JSON.
2. **Frontend**: Make device names clickable in `devices.html` — click swaps text for an `<input>`, blur/Enter saves via `fetch()`, Escape cancels.

## Files to Modify

| File | Change |
|------|--------|
| `internal/model/audit.go` | Add `EventDeviceRenamed = "device_renamed"` constant |
| `internal/hub/server.go` | Add `"/rename"` case to `handleDeviceAction` switch; add `"device_renamed"` → `"used"` mapping in `eventPillClass` |
| `internal/hub/handlers.go` | Add `handleRename` method |
| `internal/hub/handlers_test.go` | Add table-driven tests for rename handler |
| `internal/hub/templates/devices.html` | Add inline edit markup and JS |

## Build & Test

```bash
go build -o ssh-vault ./cmd/ssh-vault
go test ./...
go vet ./...
```

## Implementation Order

1. Add `EventDeviceRenamed` constant (model layer)
2. Add `handleRename` handler (hub layer)
3. Wire route in `handleDeviceAction` + add pill class mapping
4. Add handler tests
5. Update `devices.html` with inline edit UI + JS
6. Manual smoke test: rename a device, verify audit log
