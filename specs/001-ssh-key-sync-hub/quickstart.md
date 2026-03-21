# Quickstart: SSH Key Sync Hub

## Prerequisites

- Go (latest stable) installed on the hub machine
- SSH key pair on each device to enroll (e.g., `~/.ssh/id_ed25519`)
- Network connectivity from agent devices to the hub

## 1. Build

```bash
git clone <repo-url> && cd ssh-vault
go build -o ssh-vault ./cmd/ssh-vault
```

This produces a single `ssh-vault` binary.

## 2. Start the Hub

On your always-on server:

```bash
export VAULT_PASSWORD="your-secret-passphrase"
./ssh-vault hub --addr :8080 --data ./data
```

The hub starts listening on port 8080 and creates `./data/data.json`
for storage.

## 3. Generate an Onboarding Token

Open the dashboard in your browser: `http://hub-address:8080`

1. Log in with the passphrase you set in `VAULT_PASSWORD`
2. Navigate to the **Tokens** page
3. Click **Generate Token**
4. Copy the displayed token (valid for 24 hours, single use)

## 4. Enroll a Device

On the device you want to enroll:

```bash
./ssh-vault enroll \
  --hub-url http://hub-address:8080 \
  --token <paste-token-here> \
  --key ~/.ssh/id_ed25519
```

The agent will register with the hub and display:
```
Device registered. Awaiting approval.
```

## 5. Approve the Device

Back on the dashboard:

1. Refresh the device list — the new device appears as **Pending**
2. Click **Approve**

## 6. Start the Agent

On the enrolled device:

```bash
./ssh-vault agent \
  --hub-url http://hub-address:8080 \
  --key ~/.ssh/id_ed25519 \
  --interval 5m
```

The agent will:
- Sync approved keys every 5 minutes
- Write them into a managed block in `~/.ssh/authorized_keys`
- Preserve any manually added keys outside the managed block

## 7. Verify

From another enrolled device, SSH into the newly enrolled device:

```bash
ssh user@new-device
```

It should connect without prompting for a password — the public key
was distributed automatically.

## Revoking a Device

On the dashboard, click **Revoke** next to any device. Within one sync
interval (default 5 minutes), the revoked device's key is removed from
all other devices' `authorized_keys` files.

## Troubleshooting

- **Agent can't reach hub**: Check network connectivity and that the
  `--hub-url` is correct. The agent logs connection failures to stderr.
- **Keys not appearing**: Ensure the device is in "approved" status on
  the dashboard and the agent is running.
- **Permission denied on authorized_keys**: The agent needs write access
  to `~/.ssh/authorized_keys`. Check file ownership and permissions
  (should be `0600`).
