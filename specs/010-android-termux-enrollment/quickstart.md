# Quickstart: Android Termux Enrollment

## Prerequisites

On the Android device, install Termux from F-Droid (not Google Play — that version is outdated).

In Termux, install required packages:

```bash
pkg update && pkg install openssh curl termux-services
```

Enable and start crond:

```bash
sv-enable crond
```

## Enrollment

1. On the hub dashboard, click **Generate Enrollment Link**
2. Copy the enrollment command
3. Paste and run it in Termux:

```bash
curl -sSL https://your-hub/e/ABCDEF | sh
```

The script will:
- Detect the Termux environment automatically
- Download the linux/arm64 agent binary to `$PREFIX/bin`
- Generate an SSH key if none exists
- Complete the enrollment handshake
- Set up a cron job to sync keys every 15 minutes

4. On the hub dashboard, approve the pending device

## Manual Sync

To trigger an immediate key sync:

```bash
ssh-vault sync
```

## Verify Cron

Check that the sync cron job is installed:

```bash
crontab -l
```

Expected output includes:

```
*/15 * * * * /data/data/com.termux/files/usr/bin/ssh-vault sync >> /data/data/com.termux/files/home/.ssh-vault/sync.log 2>&1
```

## Troubleshooting

- **"crond not found"**: Run `pkg install termux-services && sv-enable crond`
- **"unsupported platform"**: Ensure you're on a 64-bit ARM device (`uname -m` should show `aarch64`)
- **Binary not found after install**: Check `$PREFIX/bin` is in your PATH (it should be by default in Termux)
