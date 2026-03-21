package hub

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
)

func testServerWithTemplates(t *testing.T) *Server {
	t.Helper()
	path := filepath.Join(t.TempDir(), "data.json")
	store, err := NewFileStore(path)
	if err != nil {
		t.Fatalf("NewFileStore: %v", err)
	}
	return NewServer(ServerConfig{
		Store:    store,
		Password: "test-password",
		Addr:     ":0",
	})
}

func TestLogin_CorrectPassword(t *testing.T) {
	s := testServerWithTemplates(t)

	form := url.Values{"password": {"test-password"}}
	req := httptest.NewRequest("POST", "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	s.mux.ServeHTTP(w, req)

	if w.Code != http.StatusSeeOther {
		t.Errorf("status = %d, want 303", w.Code)
	}
	if loc := w.Header().Get("Location"); loc != "/" {
		t.Errorf("Location = %q, want /", loc)
	}

	// Should have a session cookie
	cookies := w.Result().Cookies()
	found := false
	for _, c := range cookies {
		if c.Name == sessionCookieName && c.Value != "" {
			found = true
			if !c.HttpOnly {
				t.Error("session cookie should be HttpOnly")
			}
		}
	}
	if !found {
		t.Error("expected session cookie to be set")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	s := testServerWithTemplates(t)

	form := url.Values{"password": {"wrong-password"}}
	req := httptest.NewRequest("POST", "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	s.mux.ServeHTTP(w, req)

	if w.Code != http.StatusSeeOther {
		t.Errorf("status = %d, want 303", w.Code)
	}
	if loc := w.Header().Get("Location"); loc != "/login?error=1" {
		t.Errorf("Location = %q, want /login?error=1", loc)
	}
}

func TestDashboard_RequiresSession(t *testing.T) {
	s := testServerWithTemplates(t)

	// Request without session cookie
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	s.mux.ServeHTTP(w, req)

	if w.Code != http.StatusSeeOther {
		t.Errorf("status = %d, want 303 redirect to login", w.Code)
	}
	if loc := w.Header().Get("Location"); loc != "/login" {
		t.Errorf("Location = %q, want /login", loc)
	}
}

func TestDashboard_WithValidSession(t *testing.T) {
	s := testServerWithTemplates(t)

	// Create a session
	token, err := s.sessions.create()
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: token})
	w := httptest.NewRecorder()
	s.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestLogout_ClearsSession(t *testing.T) {
	s := testServerWithTemplates(t)

	token, _ := s.sessions.create()

	req := httptest.NewRequest("POST", "/logout", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: token})
	w := httptest.NewRecorder()
	s.mux.ServeHTTP(w, req)

	if w.Code != http.StatusSeeOther {
		t.Errorf("status = %d, want 303", w.Code)
	}

	// Session should be invalidated
	if s.sessions.valid(token) {
		t.Error("session should be invalid after logout")
	}
}

func TestAPIRoutes_BypassSessionAuth(t *testing.T) {
	s := testServerWithTemplates(t)

	// API routes should not redirect to login
	req := httptest.NewRequest("POST", "/api/enroll", strings.NewReader("{}"))
	w := httptest.NewRecorder()
	s.mux.ServeHTTP(w, req)

	// Should get a 400 or 401 (bad request body), not a 303 redirect
	if w.Code == http.StatusSeeOther {
		t.Error("API routes should not redirect to login")
	}
}
