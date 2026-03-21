package hub

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

const (
	sessionCookieName = "session"
	sessionTTL        = 24 * time.Hour
)

// sessionStore manages in-memory sessions for dashboard authentication.
type sessionStore struct {
	mu       sync.RWMutex
	sessions map[string]time.Time // token → expiry
}

func newSessionStore() *sessionStore {
	return &sessionStore{
		sessions: make(map[string]time.Time),
	}
}

// create generates a new session token and stores it.
func (ss *sessionStore) create() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	token := hex.EncodeToString(b)

	ss.mu.Lock()
	defer ss.mu.Unlock()
	ss.sessions[token] = time.Now().UTC().Add(sessionTTL)
	return token, nil
}

// valid checks if a session token exists and is not expired.
func (ss *sessionStore) valid(token string) bool {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	expiry, ok := ss.sessions[token]
	if !ok {
		return false
	}
	return time.Now().UTC().Before(expiry)
}

// remove deletes a session token.
func (ss *sessionStore) remove(token string) {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	delete(ss.sessions, token)
}

// handleLogin handles GET /login (show form) and POST /login (authenticate).
func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.renderTemplate(w, "login.html", map[string]any{
			"Error": r.URL.Query().Get("error") == "1",
		})
	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		password := r.FormValue("password")
		if subtle.ConstantTimeCompare([]byte(password), []byte(s.password)) != 1 {
			slog.Warn("failed login attempt")
			http.Redirect(w, r, "/login?error=1", http.StatusSeeOther)
			return
		}

		token, err := s.sessions.create()
		if err != nil {
			slog.Error("creating session", "error", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     sessionCookieName,
			Value:    token,
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
			MaxAge:   int(sessionTTL.Seconds()),
		})

		slog.Info("user logged in")
		http.Redirect(w, r, "/", http.StatusSeeOther)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleLogout handles POST /logout.
func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if cookie, err := r.Cookie(sessionCookieName); err == nil {
		s.sessions.remove(cookie.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
