package keyblock

import (
	"errors"
	"os"
	"sort"
	"strings"
)

const (
	BlockBegin = "# BEGIN SSH-VAULT MANAGED BLOCK — DO NOT EDIT"
	BlockEnd   = "# END SSH-VAULT MANAGED BLOCK"
)

// ReadBlock reads the managed block from filePath and returns the key lines
// between the BEGIN and END markers. If the file does not exist or markers are
// not found, it returns an empty slice and nil error.
func ReadBlock(filePath string) ([]string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	lines := strings.Split(string(data), "\n")

	var keys []string
	inside := false
	for _, line := range lines {
		if line == BlockBegin {
			inside = true
			continue
		}
		if line == BlockEnd {
			break
		}
		if inside && line != "" {
			keys = append(keys, line)
		}
	}

	return keys, nil
}

// WriteBlock replaces (or appends) the managed block in filePath with the given
// keys. Keys are sorted alphabetically for deterministic output. Content outside
// the block is preserved verbatim.
func WriteBlock(filePath string, keys []string) error {
	sort.Strings(keys)

	existing, err := os.ReadFile(filePath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	var result string

	if len(existing) > 0 {
		lines := strings.Split(string(existing), "\n")
		beginIdx, endIdx := -1, -1
		for i, line := range lines {
			if line == BlockBegin {
				beginIdx = i
			}
			if line == BlockEnd {
				endIdx = i
				break
			}
		}

		if beginIdx >= 0 && endIdx >= 0 {
			// Replace existing block.
			var out []string
			out = append(out, lines[:beginIdx]...)
			out = append(out, buildBlock(keys)...)
			out = append(out, lines[endIdx+1:]...)
			result = strings.Join(out, "\n")
		} else {
			// Append block at end.
			content := string(existing)
			// Ensure trailing newline + blank separator.
			if !strings.HasSuffix(content, "\n") {
				content += "\n"
			}
			content += "\n"
			content += strings.Join(buildBlock(keys), "\n") + "\n"
			result = content
		}
	} else {
		// New file.
		result = strings.Join(buildBlock(keys), "\n") + "\n"
	}

	return WriteFileAtomic(filePath, []byte(result), 0600)
}

// buildBlock returns the lines for a managed block, including markers.
func buildBlock(keys []string) []string {
	lines := []string{BlockBegin}
	lines = append(lines, keys...)
	lines = append(lines, BlockEnd)
	return lines
}
