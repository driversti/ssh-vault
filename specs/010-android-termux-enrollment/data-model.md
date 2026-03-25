# Data Model: Android Termux Enrollment

No data model changes required. Android/Termux devices are stored using the existing `Device` model with `OS: "linux"` and `Arch: "arm64"`. The hub does not distinguish between Android and standard Linux devices at the data layer.

## Existing Entities (unchanged)

### Device
- `ID` (string): Unique device identifier
- `Name` (string): User-provided device name
- `PublicKey` (string): SSH public key
- `OS` (string): Operating system — Android/Termux devices register as `"linux"`
- `Arch` (string): Architecture — Android/Termux devices register as `"arm64"`
- `Status` (string): `pending` → `approved` → `revoked`
- `APIToken` (string): Bearer token for sync API access
- `CreatedAt` (time): Enrollment timestamp

### Config (agent-side, ~/.ssh-vault/agent.json)
- `HubURL` (string): Hub server URL
- `Interval` (duration): Sync interval (5m default for daemon, 15m cron on Termux)
- `KeyPath` (string): Path to SSH private key
- `AuthKeysPath` (string): Path to authorized_keys file
- `APIToken` (string): Bearer token
- `DeviceID` (string): Device identity

No new fields, relationships, or state transitions are introduced by this feature.
