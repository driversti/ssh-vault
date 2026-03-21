package keyblock

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadBlock_FileNotExist(t *testing.T) {
	keys, err := ReadBlock("/nonexistent/path")
	if err != nil {
		t.Fatalf("ReadBlock: %v", err)
	}
	if len(keys) != 0 {
		t.Errorf("expected empty keys, got %v", keys)
	}
}

func TestReadBlock_NoMarkers(t *testing.T) {
	f := filepath.Join(t.TempDir(), "authorized_keys")
	os.WriteFile(f, []byte("ssh-rsa AAAA... user@host\n"), 0600)

	keys, err := ReadBlock(f)
	if err != nil {
		t.Fatalf("ReadBlock: %v", err)
	}
	if len(keys) != 0 {
		t.Errorf("expected empty keys, got %v", keys)
	}
}

func TestReadBlock_WithBlock(t *testing.T) {
	content := "ssh-rsa AAAA... manual@host\n" +
		BlockBegin + "\n" +
		"ssh-ed25519 AAAA... device-a\n" +
		"ssh-ed25519 BBBB... device-b\n" +
		BlockEnd + "\n"

	f := filepath.Join(t.TempDir(), "authorized_keys")
	os.WriteFile(f, []byte(content), 0600)

	keys, err := ReadBlock(f)
	if err != nil {
		t.Fatalf("ReadBlock: %v", err)
	}
	if len(keys) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(keys))
	}
	if keys[0] != "ssh-ed25519 AAAA... device-a" {
		t.Errorf("keys[0] = %q", keys[0])
	}
}

func TestWriteBlock_NewFile(t *testing.T) {
	f := filepath.Join(t.TempDir(), "authorized_keys")

	keys := []string{
		"ssh-ed25519 BBBB... device-b",
		"ssh-ed25519 AAAA... device-a",
	}
	if err := WriteBlock(f, keys); err != nil {
		t.Fatalf("WriteBlock: %v", err)
	}

	data, err := os.ReadFile(f)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	content := string(data)

	// Keys should be sorted
	if !strings.Contains(content, "ssh-ed25519 AAAA... device-a\nssh-ed25519 BBBB... device-b") {
		t.Errorf("keys not sorted in output:\n%s", content)
	}
	if !strings.HasPrefix(content, BlockBegin) {
		t.Error("should start with BEGIN marker")
	}
	if !strings.Contains(content, BlockEnd) {
		t.Error("should contain END marker")
	}
}

func TestWriteBlock_PreservesExistingContent(t *testing.T) {
	f := filepath.Join(t.TempDir(), "authorized_keys")
	existing := "ssh-rsa AAAA... manual@host\n"
	os.WriteFile(f, []byte(existing), 0600)

	if err := WriteBlock(f, []string{"ssh-ed25519 AAAA... device-a"}); err != nil {
		t.Fatalf("WriteBlock: %v", err)
	}

	data, _ := os.ReadFile(f)
	content := string(data)

	if !strings.Contains(content, "ssh-rsa AAAA... manual@host") {
		t.Error("existing content should be preserved")
	}
	if !strings.Contains(content, BlockBegin) {
		t.Error("should contain managed block")
	}
}

func TestWriteBlock_ReplacesExistingBlock(t *testing.T) {
	f := filepath.Join(t.TempDir(), "authorized_keys")
	existing := "manual-key\n" +
		BlockBegin + "\n" +
		"old-key\n" +
		BlockEnd + "\n" +
		"another-manual-key\n"
	os.WriteFile(f, []byte(existing), 0600)

	if err := WriteBlock(f, []string{"new-key"}); err != nil {
		t.Fatalf("WriteBlock: %v", err)
	}

	data, _ := os.ReadFile(f)
	content := string(data)

	if strings.Contains(content, "old-key") {
		t.Error("old block content should be replaced")
	}
	if !strings.Contains(content, "new-key") {
		t.Error("new key should be present")
	}
	if !strings.Contains(content, "manual-key") {
		t.Error("content before block should be preserved")
	}
	if !strings.Contains(content, "another-manual-key") {
		t.Error("content after block should be preserved")
	}
}

func TestWriteBlock_EmptyKeys(t *testing.T) {
	f := filepath.Join(t.TempDir(), "authorized_keys")

	if err := WriteBlock(f, nil); err != nil {
		t.Fatalf("WriteBlock: %v", err)
	}

	data, _ := os.ReadFile(f)
	content := string(data)

	if !strings.Contains(content, BlockBegin) || !strings.Contains(content, BlockEnd) {
		t.Error("empty block should still have markers")
	}
}

func TestWriteBlock_FilePermissions(t *testing.T) {
	f := filepath.Join(t.TempDir(), "authorized_keys")

	if err := WriteBlock(f, []string{"key1"}); err != nil {
		t.Fatalf("WriteBlock: %v", err)
	}

	info, err := os.Stat(f)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Errorf("permissions = %o, want 0600", perm)
	}
}

func TestWriteFileAtomic_Symlink(t *testing.T) {
	dir := t.TempDir()
	realFile := filepath.Join(dir, "real_file")
	symlink := filepath.Join(dir, "link_file")

	// Create the real file first
	os.WriteFile(realFile, []byte("original"), 0600)
	os.Symlink(realFile, symlink)

	if err := WriteFileAtomic(symlink, []byte("updated"), 0600); err != nil {
		t.Fatalf("WriteFileAtomic: %v", err)
	}

	// Read via the real path
	data, _ := os.ReadFile(realFile)
	if string(data) != "updated" {
		t.Errorf("content = %q, want %q", string(data), "updated")
	}
}
