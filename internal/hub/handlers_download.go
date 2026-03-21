package hub

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var (
	validOS   = map[string]bool{"linux": true, "darwin": true}
	validArch = map[string]bool{"amd64": true, "arm64": true}
)

// handleDownload serves pre-built binaries from the dist directory.
// Routes: GET /download/{os}/{arch} and GET /download/checksums.txt
func (s *Server) handleDownload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.distDir == "" {
		http.Error(w, "binary distribution not configured", http.StatusNotImplemented)
		return
	}

	// Strip prefix to get the remaining path segments.
	remainder := strings.TrimPrefix(r.URL.Path, "/download/")

	// Special case: checksums.txt
	if remainder == "checksums.txt" {
		s.serveChecksums(w, r)
		return
	}

	// Expect exactly {os}/{arch}
	parts := strings.SplitN(remainder, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		http.Error(w, supportedPlatformsMsg(), http.StatusBadRequest)
		return
	}
	osName, arch := parts[0], parts[1]

	if !validOS[osName] || !validArch[arch] {
		http.Error(w, fmt.Sprintf("unsupported platform: %s/%s. %s", osName, arch, supportedPlatformsMsg()), http.StatusBadRequest)
		return
	}

	filename := fmt.Sprintf("ssh-vault_%s_%s", osName, arch)
	filePath := filepath.Join(s.distDir, filename)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, fmt.Sprintf("binary not available for %s/%s", osName, arch), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	http.ServeFile(w, r, filePath)
}

// serveChecksums serves the checksums.txt file from the dist directory.
func (s *Server) serveChecksums(w http.ResponseWriter, r *http.Request) {
	filePath := filepath.Join(s.distDir, "checksums.txt")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "checksums not available", http.StatusNotFound)
		return
	}
	http.ServeFile(w, r, filePath)
}

func supportedPlatformsMsg() string {
	return "Supported: linux/amd64, linux/arm64, darwin/amd64, darwin/arm64"
}
