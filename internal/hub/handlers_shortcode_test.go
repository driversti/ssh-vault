package hub

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/driversti/ssh-vault/internal/model"
)

// testServerWithEnrollment creates a Server configured for short enrollment testing.
func testServerWithEnrollment(t *testing.T) *Server {
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
		GithubRepo:  "testowner/testrepo",
		ReleaseTag:  "v1.0.0",
	})
}

func TestHandleGenerateLink(t *testing.T) {
	s := testServerWithEnrollment(t)

	req := httptest.NewRequest(http.MethodPost, "/tokens/generate-link", nil)
	addSessionCookie(t, s, req)
	rr := httptest.NewRecorder()

	s.mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusSeeOther)
	}
	if loc := rr.Header().Get("Location"); loc != "/tokens" {
		t.Errorf("Location = %q, want /tokens", loc)
	}

	// Verify a short code was created
	codes := s.store.ListShortCodes()
	if len(codes) != 1 {
		t.Fatalf("expected 1 short code, got %d", len(codes))
	}
	sc := codes[0]
	if len(sc.Code) != 6 {
		t.Errorf("code length = %d, want 6", len(sc.Code))
	}
	if sc.TokenValue == "" {
		t.Error("short code should have a linked token")
	}

	// Verify the linked token was created
	tok, err := s.store.GetToken(sc.TokenValue)
	if err != nil {
		t.Fatalf("linked token not found: %v", err)
	}
	if tok.Used {
		t.Error("linked token should not be used yet")
	}

	// Verify audit entry
	entries := s.store.ListAuditLog()
	if len(entries) == 0 {
		t.Fatal("expected audit entry")
	}
	found := false
	for _, e := range entries {
		if e.Event == model.EventShortCodeCreated {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected shortcode_created audit entry")
	}
}

func TestHandleGenerateLink_RequiresSession(t *testing.T) {
	s := testServerWithEnrollment(t)

	req := httptest.NewRequest(http.MethodPost, "/tokens/generate-link", nil)
	rr := httptest.NewRecorder()

	s.mux.ServeHTTP(rr, req)

	// Should redirect to login (requireSession middleware)
	if rr.Code != http.StatusSeeOther {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusSeeOther)
	}
	if loc := rr.Header().Get("Location"); loc != "/login" {
		t.Errorf("Location = %q, want /login", loc)
	}
}

func TestHandleGenerateLink_NotConfigured(t *testing.T) {
	// Server without ExternalURL
	s := testServer(t)

	req := httptest.NewRequest(http.MethodPost, "/tokens/generate-link", nil)
	addSessionCookie(t, s, req)
	rr := httptest.NewRecorder()

	s.mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusInternalServerError)
	}
}

func TestHandleShortCodeEnroll_ValidCode(t *testing.T) {
	s := testServerWithEnrollment(t)

	// Create a token and short code
	tok, _ := model.NewToken(24 * time.Hour)
	s.store.AddToken(*tok)
	sc := model.NewShortCode("123456", tok.Value, 15*time.Minute)
	s.store.AddShortCode(*sc)

	req := httptest.NewRequest(http.MethodGet, "/e/123456", nil)
	rr := httptest.NewRecorder()

	s.mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}

	ct := rr.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "text/plain") {
		t.Errorf("Content-Type = %q, want text/plain", ct)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "#!/bin/sh") {
		t.Error("response should contain shell shebang")
	}
	if !strings.Contains(body, tok.Value) {
		t.Error("response should contain the enrollment token")
	}
	if !strings.Contains(body, "https://test.example.com") {
		t.Error("response should contain the hub URL")
	}
	if !strings.Contains(body, "testowner/testrepo") {
		t.Error("response should contain the GitHub repo in download URL")
	}
}

func TestHandleShortCodeEnroll_ExpiredCode(t *testing.T) {
	s := testServerWithEnrollment(t)

	tok, _ := model.NewToken(24 * time.Hour)
	s.store.AddToken(*tok)
	sc := model.NewShortCode("654321", tok.Value, 15*time.Minute)
	// Manually expire it
	expired := time.Now().UTC().Add(-1 * time.Minute)
	sc.ExpiresAt = expired
	s.store.AddShortCode(*sc)

	req := httptest.NewRequest(http.MethodGet, "/e/654321", nil)
	rr := httptest.NewRecorder()

	s.mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusGone {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusGone)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "expired") {
		t.Error("response should mention expiration")
	}
}

func TestHandleShortCodeEnroll_UsedCode(t *testing.T) {
	s := testServerWithEnrollment(t)

	tok, _ := model.NewToken(24 * time.Hour)
	s.store.AddToken(*tok)
	sc := model.NewShortCode("111111", tok.Value, 15*time.Minute)
	sc.MarkUsed("1.2.3.4")
	s.store.AddShortCode(*sc)

	req := httptest.NewRequest(http.MethodGet, "/e/111111", nil)
	rr := httptest.NewRecorder()

	s.mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusNotFound)
	}
}

func TestHandleShortCodeEnroll_InvalidCode(t *testing.T) {
	s := testServerWithEnrollment(t)

	req := httptest.NewRequest(http.MethodGet, "/e/999999", nil)
	rr := httptest.NewRecorder()

	s.mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusNotFound)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "echo") {
		t.Error("error response should be a shell script")
	}
}

func TestHandleShortCodeEnroll_RateLimit(t *testing.T) {
	s := testServerWithEnrollment(t)
	// Override with a very low limit for testing
	s.enrollLimiter = NewRateLimiter(1*time.Minute, 2)

	tok, _ := model.NewToken(24 * time.Hour)
	s.store.AddToken(*tok)
	sc := model.NewShortCode("222222", tok.Value, 15*time.Minute)
	s.store.AddShortCode(*sc)

	// First 2 requests should succeed
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/e/222222", nil)
		req.RemoteAddr = "10.0.0.1:1234"
		rr := httptest.NewRecorder()
		s.mux.ServeHTTP(rr, req)
		if rr.Code == http.StatusTooManyRequests {
			t.Errorf("request %d should not be rate-limited", i+1)
		}
	}

	// 3rd request should be rate-limited
	req := httptest.NewRequest(http.MethodGet, "/e/222222", nil)
	req.RemoteAddr = "10.0.0.1:1234"
	rr := httptest.NewRecorder()
	s.mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusTooManyRequests {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusTooManyRequests)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "Too many requests") {
		t.Error("response should mention rate limiting")
	}
}

func TestHandleShortCodeEnroll_ScriptContainsTemplateVars(t *testing.T) {
	s := testServerWithEnrollment(t)

	tok, _ := model.NewToken(24 * time.Hour)
	s.store.AddToken(*tok)
	sc := model.NewShortCode("333333", tok.Value, 15*time.Minute)
	s.store.AddShortCode(*sc)

	req := httptest.NewRequest(http.MethodGet, "/e/333333", nil)
	rr := httptest.NewRecorder()

	s.mux.ServeHTTP(rr, req)

	body := rr.Body.String()

	checks := []struct {
		name    string
		content string
	}{
		{"hub URL variable", `VAULT_HUB_URL="https://test.example.com"`},
		{"token variable", `VAULT_TOKEN="` + tok.Value + `"`},
		{"download base URL", `releases/download/v1.0.0`},
		{"platform detection", `uname -s`},
		{"arch detection", `uname -m`},
		{"SSH key discovery", `~/.ssh/id_ed25519.pub`},
		{"hostname detection", `hostname`},
		{"enroll command", `enroll`},
		{"cleanup trap", `trap cleanup EXIT`},
	}

	for _, check := range checks {
		if !strings.Contains(body, check.content) {
			t.Errorf("script missing %s: expected %q in output", check.name, check.content)
		}
	}
}
