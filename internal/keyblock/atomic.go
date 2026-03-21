package keyblock

import (
	"os"
	"path/filepath"
)

// WriteFileAtomic writes data to path atomically by creating a temp file in the
// same directory and renaming it into place. Symlinks are resolved so the final
// rename targets the real file.
func WriteFileAtomic(path string, data []byte, perm os.FileMode) error {
	resolvedPath, err := filepath.EvalSymlinks(path)
	if err != nil {
		// File may not exist yet — use the original path.
		resolvedPath = path
	}

	tmp, err := os.CreateTemp(filepath.Dir(resolvedPath), ".tmp-*")
	if err != nil {
		return err
	}
	tempPath := tmp.Name()

	// Clean up on any failure path.
	success := false
	defer func() {
		if !success {
			os.Remove(tempPath)
		}
	}()

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}

	if err := os.Chmod(tempPath, perm); err != nil {
		return err
	}

	if err := os.Rename(tempPath, resolvedPath); err != nil {
		return err
	}

	success = true
	return nil
}
