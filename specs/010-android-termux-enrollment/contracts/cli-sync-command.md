# Contract: `ssh-vault sync` CLI Command

## Command Signature

```
ssh-vault sync [flags]
```

## Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--config` | string | `~/.ssh-vault/agent.json` | Path to agent config file |

## Behavior

1. Loads agent configuration from config file
2. Performs a single sync cycle (fetch keys from hub, write authorized_keys)
3. Exits with code 0 on success, non-zero on failure
4. Outputs structured log line on completion

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Sync completed successfully |
| 1 | Configuration error (missing config, invalid token) |
| 2 | Hub unreachable (network error) |
| 3 | Device revoked |

## Cron Integration

```crontab
*/15 * * * * /path/to/ssh-vault sync >> ~/.ssh-vault/sync.log 2>&1
```

## Example Output

```
2026-03-25T10:15:00Z INFO sync complete keys=5
```
