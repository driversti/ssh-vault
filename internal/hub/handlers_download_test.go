package hub

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// testServerWithDist creates a Server with a dist directory for testing.
func testServerWithDist(t *testing.T, distDir string) *Server {
	t.Helper()
	path := filepath.Join(t.TempDir(), "data.json")
	store, err := NewFileStore(path)
	if err != nil {
		t.Fatalf("NewFileStore: %v", err)
	}
	return NewServer(ServerConfig{
		Store:       store,
		Password:    "test-password",
		Addr:        ":0",
		ExternalURL: "https://test.example.com",
		DistDir:     distDir,
	})
}

func TestHandleDownload(t *testing.T) {
	// Create a dist directory with a fake binary
	distDir := t.TempDir()
	fakeContent := []byte("fake-binary-content")
	if err := os.WriteFile(filepath.Join(distDir, "ssh-vault_linux_amd64"), fakeContent, 0o755); err != nil {
		t.Fatalf("writing fake binary: %v", err)
	}
	if err := os.WriteFile(filepath.Join(distDir, "checksums.txt"), []byte("abc123 ssh-vault_linux_amd64\n"), 0o644); err != nil {
		t.Fatalf("writing checksums: %v", err)
	}

	s := testServerWithDist(t, distDir)

	tests := []struct {
		name           string
		path           string
		wantStatus     int
		wantBodySubstr string
	}{
		{
			name:       "valid binary download",
			path:       "/download/linux/amd64",
			wantStatus: http.StatusOK,
		},
		{
			name:           "invalid OS",
			path:           "/download/windows/amd64",
			wantStatus:     http.StatusBadRequest,
			wantBodySubstr: "unsupported platform",
		},
		{
			name:           "invalid arch",
			path:           "/download/linux/mips",
			wantStatus:     http.StatusBadRequest,
			wantBodySubstr: "unsupported platform",
		},
		{
			name:           "missing binary",
			path:           "/download/darwin/arm64",
			wantStatus:     http.StatusNotFound,
			wantBodySubstr: "binary not available",
		},
		{
			name:       "checksums.txt",
			path:       "/download/checksums.txt",
			wantStatus: http.StatusOK,
		},
		{
			name:           "directory traversal via valid-looking path",
			path:           "/download/..%2F..%2Fetc/passwd",
			wantStatus:     http.StatusBadRequest,
			wantBodySubstr: "unsupported platform",
		},
		{
			name:           "empty path segments",
			path:           "/download/",
			wantStatus:     http.StatusBadRequest,
			wantBodySubstr: "Supported:",
		},
		{
			name:           "only OS no arch",
			path:           "/download/linux",
			wantStatus:     http.StatusBadRequest,
			wantBodySubstr: "Supported:",
		},
		{
			name:           "method not allowed",
			path:           "/download/linux/amd64",
			wantStatus:     http.StatusMethodNotAllowed,
			wantBodySubstr: "method not allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			method := http.MethodGet
			if tt.name == "method not allowed" {
				method = http.MethodPost
			}
			req := httptest.NewRequest(method, tt.path, nil)
			rr := httptest.NewRecorder()

			s.mux.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d (body: %s)", rr.Code, tt.wantStatus, rr.Body.String())
			}
			if tt.wantBodySubstr != "" {
				body := rr.Body.String()
				if !strings.Contains(body, tt.wantBodySubstr) {
					t.Errorf("body %q does not contain %q", body, tt.wantBodySubstr)
				}
			}
		})
	}

	// Verify Content-Disposition on successful download
	t.Run("content-disposition header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/download/linux/amd64", nil)
		rr := httptest.NewRecorder()
		s.mux.ServeHTTP(rr, req)

		cd := rr.Header().Get("Content-Disposition")
		if cd != `attachment; filename="ssh-vault_linux_amd64"` {
			t.Errorf("Content-Disposition = %q, want attachment with filename", cd)
		}
	})
}

func TestHandleDownload_DistNotConfigured(t *testing.T) {
	s := testServerWithDist(t, "") // empty distDir

	req := httptest.NewRequest(http.MethodGet, "/download/linux/amd64", nil)
	rr := httptest.NewRecorder()
	s.mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotImplemented {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusNotImplemented)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "not configured") {
		t.Errorf("body %q should mention not configured", body)
	}
}

func TestHandleDownload_ChecksumsNotAvailable(t *testing.T) {
	distDir := t.TempDir() // empty directory, no checksums.txt
	s := testServerWithDist(t, distDir)

	req := httptest.NewRequest(http.MethodGet, "/download/checksums.txt", nil)
	rr := httptest.NewRecorder()
	s.mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusNotFound)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "checksums not available") {
		t.Errorf("body %q should mention checksums not available", body)
	}
}

