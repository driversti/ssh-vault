# Research: Auto SSH Key Generation During Enrollment

## Phase 0 Research

No NEEDS CLARIFICATION items in Technical Context. Research focused on best practices.

### 1. ssh-keygen Invocation for Non-Interactive Key Generation

**Decision**: Use `ssh-keygen -t ed25519 -f <path> -N "" -q`
**Rationale**:
- `-t ed25519`: Preferred modern key type, consistent with existing codebase
- `-f <path>`: Explicit output path (avoids interactive prompt for filename)
- `-N ""`: Empty passphrase (avoids interactive passphrase prompt)
- `-q`: Quiet mode (suppresses unnecessary output)
- This combination is fully non-interactive — works in both shell scripts and `os/exec` calls
**Alternatives considered**:
- `-t rsa`: Larger key, slower generation, no security benefit for this use case
- `-t ecdsa`: Less widely supported than ed25519
- Go `crypto/ed25519`: Requires manual OpenSSH format encoding and file permission management

### 2. Directory and Permission Handling

**Decision**: Create `~/.ssh/` with mode `700` if missing; rely on `ssh-keygen` for key file permissions
**Rationale**: `ssh-keygen` automatically sets private key to `600` and public key to `644`. We only need to ensure the parent directory exists with correct permissions. In the shell script, `mkdir -p -m 700 ~/.ssh` handles this. In Go, `os.MkdirAll` with `0700` permission.
**Alternatives considered**:
- Let `ssh-keygen` fail and report the error: Less user-friendly, defeats the purpose of the feature

### 3. Preventing Overwrites of Existing Keys

**Decision**: Check for file existence before calling `ssh-keygen`; never pass `-y` or force flags
**Rationale**: `ssh-keygen` itself refuses to overwrite existing keys (prompts for confirmation), but since we run it non-interactively, it would fail rather than overwrite. As defense-in-depth, we explicitly check for existing keys before invoking `ssh-keygen`. This matches FR-005 requirement.
**Alternatives considered**:
- Rely solely on `ssh-keygen`'s built-in protection: Works, but explicit check provides clearer error messages and is more testable

### 4. Go EnsureSSHKey Function Design

**Decision**: `EnsureSSHKey(keyPath string, promptFn func(string) bool) (string, error)`
**Rationale**: The `promptFn` parameter allows the caller to control whether/how to prompt the user. The CLI sets it to a stdin-based prompt; tests can pass a mock. Returns the key path (unchanged if key exists, or the newly generated path).
**Alternatives considered**:
- Boolean `prompt` flag: Less flexible, harder to test
- Separate functions for prompted/unprompted: Duplication for a single conditional
