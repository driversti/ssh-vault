# Research: Android Termux Enrollment

## R1: Existing linux/arm64 Binary for Termux

**Decision**: The existing `linux/arm64` cross-compilation with `CGO_ENABLED=0` is sufficient for Termux. No Android-specific build flags needed.

**Rationale**: Termux provides a standard Linux ARM64 userland. Go's static linking (`CGO_ENABLED=0`) produces a fully self-contained binary with no libc dependency, which avoids Termux's non-standard Bionic libc paths. The binary format is standard ELF for linux/arm64.

**Alternatives considered**:
- Android NDK cross-compilation: Unnecessary — Termux is a Linux environment, not Android native
- Dynamic linking with Termux's libc: Rejected — adds complexity and fragility

## R2: Dockerfile Missing linux/arm64 in Dist Stage

**Decision**: Add a `build-linux-arm64` stage to the Dockerfile and copy its output to the dist stage.

**Rationale**: The Dockerfile has `build-linux-amd64` and `build-darwin-arm64` stages but does NOT build or include `linux/arm64` in the dist assembly stage. The download handler already validates and serves `linux/arm64` (validation maps include it), but the binary doesn't exist. This is the primary blocker for Android enrollment.

**Alternatives considered**:
- Serving darwin/arm64 for Android: Incorrect — different OS ABI
- Building only at deploy time: Rejected — inconsistent with current Docker-based build pipeline

## R3: CLI Subcommand Architecture for One-Shot Sync

**Decision**: Add a `sync` subcommand to the existing flag-based CLI in `main.go` that calls the existing `syncOnce()` function (exported as `SyncOnce()`).

**Rationale**: The CLI uses simple `os.Args[1]` switch routing with `flag.NewFlagSet` per command. Adding `sync` follows this established pattern. The `syncOnce()` function in `agent.go` already does exactly what's needed (fetch keys + write authorized_keys), it just needs to be exported.

**Alternatives considered**:
- Adding a `--once` flag to `agent` command: Less discoverable for cron usage
- Separate binary for sync: Violates constitution (one binary principle)

## R4: Enrollment Script Termux Detection

**Decision**: Detect Termux via `$TERMUX_VERSION` environment variable (set by all Termux installations). Fall back to checking `$PREFIX` starts with `/data/data/com.termux`.

**Rationale**: `TERMUX_VERSION` is the canonical way to detect Termux. The enrollment script already has OS/arch detection — Termux detection adds a conditional branch for installation paths and post-enrollment setup (cron).

**Alternatives considered**:
- Check for `/data/data/com.termux/` path only: Less reliable — could exist on rooted devices without Termux
- Check `uname -o` for Android: Not all Termux versions report this consistently

## R5: Cron Job Setup on Termux

**Decision**: After enrollment, if Termux is detected, install a crontab entry via `crontab -l | { cat; echo "*/15 * * * * ..."; } | crontab -` pattern. Require `crond` from `termux-services` to be installed and running.

**Rationale**: Termux's `crond` (from `termux-services` package) supports standard crontab syntax. The enrollment script should check that `sv-enable crond` has been run (or run it), then install the crontab entry. The 15-minute interval was chosen to balance battery life with key freshness.

**Alternatives considered**:
- `termux-job-scheduler`: Deprecated and less reliable
- `termux-boot` for daemon auto-restart: Only runs on device boot, not app restart; doesn't solve Android process killing
- Manual user setup: Poor UX — the enrollment script should automate this

## R6: Enrollment Script Path Adaptation

**Decision**: When Termux detected, install binary to `$PREFIX/bin` (no sudo needed). For non-Termux, keep existing logic (`/usr/local/bin` with sudo, fallback to `~/.local/bin`).

**Rationale**: `$PREFIX/bin` is in Termux's default PATH. Termux has no `sudo` by default and doesn't need it — the user owns the entire `$PREFIX` tree. The existing fallback to `~/.local/bin` would also work but `$PREFIX/bin` is more idiomatic for Termux.

**Alternatives considered**:
- Always use `~/.local/bin` on Termux: Works but non-idiomatic; may not be in PATH by default
- Require user to add custom PATH: Poor UX
