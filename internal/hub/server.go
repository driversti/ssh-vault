package hub

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/driversti/ssh-vault/internal/model"
)

//go:embed templates/*
var templateFS embed.FS

type contextKey string

const deviceContextKey contextKey = "device"

// Server is the hub HTTP server.
type Server struct {
	store       *FileStore
	password    string
	sessions    *sessionStore
	mux         *http.ServeMux
	addr        string
	tlsCert     string
	tlsKey      string
	tmpl        map[string]*template.Template
	externalURL string
	distDir     string
	enrollLimiter *RateLimiter
}

// ServerConfig holds configuration for creating a new Server.
type ServerConfig struct {
	Store       *FileStore
	Password    string
	Addr        string
	TLSCert     string
	TLSKey      string
	ExternalURL string
	DistDir     string
}

// NewServer creates a new hub server with all routes registered.
func NewServer(cfg ServerConfig) *Server {
	funcMap := template.FuncMap{
		"formatTime": func(t time.Time) string {
			if t.IsZero() {
				return "—"
			}
			return t.UTC().Format(time.RFC3339)
		},
		"formatTimePtr": func(t *time.Time) string {
			if t == nil {
				return "never"
			}
			return t.UTC().Format(time.RFC3339)
		},
		"formatFingerprint": func(fp string) string {
			if len(fp) > 20 {
				return fp[:20] + "..."
			}
			return fp
		},
		"isStale": func(d model.Device) bool {
			if d.LastSyncAt == nil {
				return d.Status == model.StatusApproved
			}
			return time.Since(*d.LastSyncAt) > 15*time.Minute
		},
		"upper": strings.ToUpper,
		"eventPillClass": func(event string) string {
			switch event {
			case "approved":
				return "approved"
			case "enrolled":
				return "enrolled"
			case "revoked", "auth_failed":
				return "revoked"
			case "token_used", "shortcode_used", "shortcode_created":
				return "used"
			case "token_removed", "shortcode_expired":
				return "expired"
			default:
				return "expired"
			}
		},
	}

	// Parse each page template individually with the layout to avoid
	// colliding "content"/"title" block definitions.
	layoutText, _ := fs.ReadFile(templateFS, "templates/layout.html")
	pages := []string{"devices.html", "tokens.html", "audit.html", "login.html"}
	tmpls := make(map[string]*template.Template, len(pages))
	for _, page := range pages {
		pageText, _ := fs.ReadFile(templateFS, "templates/"+page)
		t := template.Must(
			template.Must(
				template.New(page).Funcs(funcMap).Parse(string(layoutText)),
			).Parse(string(pageText)),
		)
		tmpls[page] = t
	}

	s := &Server{
		store:         cfg.Store,
		password:      cfg.Password,
		sessions:      newSessionStore(),
		mux:           http.NewServeMux(),
		addr:          cfg.Addr,
		tlsCert:       cfg.TLSCert,
		tlsKey:        cfg.TLSKey,
		tmpl:          tmpls,
		externalURL:   cfg.ExternalURL,
		distDir:       cfg.DistDir,
		enrollLimiter: NewRateLimiter(1*time.Minute, 10),
	}
	s.registerRoutes()
	return s
}

func (s *Server) registerRoutes() {
	// API routes (bearer token auth — no session required)
	s.mux.HandleFunc("/api/enroll", s.handleEnroll)
	s.mux.HandleFunc("/api/enroll/verify", s.handleEnrollVerify)
	s.mux.HandleFunc("/api/keys", s.requireBearerAuth(s.handleKeys))

	// Auth routes (no session required)
	s.mux.HandleFunc("/login", s.handleLogin)
	s.mux.HandleFunc("/logout", s.handleLogout)

	// Static assets
	s.mux.HandleFunc("/static/theme.css", s.handleStaticCSS)
	s.mux.HandleFunc("/static/logo.svg", s.handleStaticLogo)

	// Health check (no auth)
	s.mux.HandleFunc("/healthz", s.handleHealthz)

	// Binary downloads (public)
	s.mux.HandleFunc("/download/", s.handleDownload)

	// Short enrollment URL (public, rate-limited)
	s.mux.HandleFunc("/e/", s.handleShortCodeEnroll)

	// Dashboard routes (session auth required)
	s.mux.HandleFunc("/", s.requireSession(s.handleDashboard))
	s.mux.HandleFunc("/tokens", s.requireSession(s.handleTokens))
	s.mux.HandleFunc("/tokens/generate-link", s.requireSession(s.handleGenerateLink))
	s.mux.HandleFunc("/tokens/", s.requireSession(s.handleTokenAction))
	s.mux.HandleFunc("/audit", s.requireSession(s.handleAudit))
	s.mux.HandleFunc("/devices/", s.requireSession(s.handleDeviceAction))
}

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

// handleDashboard shows the device list.
func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	devices := s.store.ListDevices()
	// Sort: pending first, then approved, then revoked
	sort.Slice(devices, func(i, j int) bool {
		order := map[string]int{model.StatusPending: 0, model.StatusApproved: 1, model.StatusRevoked: 2}
		return order[devices[i].Status] < order[devices[j].Status]
	})
	s.renderTemplate(w, "devices.html", map[string]any{
		"Devices":       devices,
		"ActivePage":    "devices",
		"TotalCount":    len(devices),
		"ApprovedCount": countByStatus(devices, model.StatusApproved),
		"PendingCount":  countByStatus(devices, model.StatusPending),
		"RevokedCount":  countByStatus(devices, model.StatusRevoked),
	})
}

// handleTokens shows the token list and handles token generation.
func (s *Server) handleTokens(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// Purge expired tokens on page load (FR-009).
		if purged, err := s.store.PurgeExpiredTokens(); err != nil {
			slog.Error("purging expired tokens", "error", err)
		} else if purged > 0 {
			slog.Info("purged expired tokens", "count", purged)
		}

		// Purge expired short codes and orphaned tokens.
		if purged, err := s.store.PurgeExpiredShortCodes(); err != nil {
			slog.Error("purging expired short codes", "error", err)
		} else if purged > 0 {
			slog.Info("purged expired short codes", "count", purged)
		}

		tokens := s.store.ListTokens()
		// Filter to active (unused, not expired)
		var active []model.Token
		for _, t := range tokens {
			if t.IsValid() {
				active = append(active, t)
			}
		}

		// Collect active short codes for display
		shortCodes := s.store.ListShortCodes()
		var activeShortCodes []model.ShortCode
		for _, sc := range shortCodes {
			if sc.IsValid() {
				activeShortCodes = append(activeShortCodes, sc)
			}
		}

		s.renderTemplate(w, "tokens.html", map[string]any{
			"Tokens":      active,
			"ShortCodes":  activeShortCodes,
			"ExternalURL": s.externalURL,
			"ActivePage":  "tokens",
		})
	case http.MethodPost:
		tok, err := model.NewToken(24 * time.Hour)
		if err != nil {
			slog.Error("generating token", "error", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if err := s.store.AddToken(*tok); err != nil {
			slog.Error("saving token", "error", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		slog.Info("token generated", "expires", tok.ExpiresAt)
		http.Redirect(w, r, "/tokens", http.StatusSeeOther)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleAudit shows the audit log.
func (s *Server) handleAudit(w http.ResponseWriter, r *http.Request) {
	entries := s.store.ListAuditLog()
	// Reverse chronological order
	for i, j := 0, len(entries)-1; i < j; i, j = i+1, j-1 {
		entries[i], entries[j] = entries[j], entries[i]
	}
	s.renderTemplate(w, "audit.html", map[string]any{
		"Entries":    entries,
		"ActivePage": "audit",
	})
}

// handleStaticCSS serves the embedded Pico CSS file.
func (s *Server) handleStaticCSS(w http.ResponseWriter, r *http.Request) {
	data, err := templateFS.ReadFile("templates/theme.css")
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/css")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.Write(data)
}

// handleStaticLogo serves the embedded SVG logo file.
func (s *Server) handleStaticLogo(w http.ResponseWriter, r *http.Request) {
	data, err := templateFS.ReadFile("templates/logo.svg")
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.Write(data)
}

// handleTokenAction routes /tokens/{value}/remove.
func (s *Server) handleTokenAction(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	switch {
	case strings.HasSuffix(path, "/remove"):
		s.handleRemoveToken(w, r)
	default:
		http.NotFound(w, r)
	}
}

// handleDeviceAction routes /devices/{id}/approve and /devices/{id}/revoke.
func (s *Server) handleDeviceAction(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	switch {
	case strings.HasSuffix(path, "/approve"):
		s.handleApprove(w, r)
	case strings.HasSuffix(path, "/revoke"):
		s.handleRevoke(w, r)
	case strings.HasSuffix(path, "/remove"):
		s.handleRemoveDevice(w, r)
	default:
		http.NotFound(w, r)
	}
}

// requireBearerAuth is middleware that validates the Bearer token
// and attaches the device to the request context.
func (s *Server) requireBearerAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing bearer token"})
			return
		}
		token := strings.TrimPrefix(auth, "Bearer ")

		device, err := s.store.GetDeviceByAPIToken(token)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid token"})
			return
		}

		if device.Status == model.StatusRevoked {
			s.store.AddAuditEntry(model.NewAuditEntry(
				model.EventAuthFailed, device.ID,
				fmt.Sprintf("Revoked device '%s' attempted sync", device.Name),
			))
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "device revoked"})
			return
		}

		ctx := context.WithValue(r.Context(), deviceContextKey, device)
		next(w, r.WithContext(ctx))
	}
}

// requireSession checks for a valid session cookie, redirecting to /login if absent.
func (s *Server) requireSession(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(sessionCookieName)
		if err != nil || !s.sessions.valid(cookie.Value) {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next(w, r)
	}
}

// deviceFromContext retrieves the authenticated device from request context.
func deviceFromContext(ctx context.Context) *model.Device {
	d, _ := ctx.Value(deviceContextKey).(*model.Device)
	return d
}

// countByStatus returns the number of devices with the given status.
func countByStatus(devices []model.Device, status string) int {
	n := 0
	for _, d := range devices {
		if d.Status == status {
			n++
		}
	}
	return n
}

// renderTemplate renders a named template with the given data.
func (s *Server) renderTemplate(w http.ResponseWriter, name string, data any) {
	t, ok := s.tmpl[name]
	if !ok {
		slog.Error("template not found", "template", name)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := t.ExecuteTemplate(w, name, data); err != nil {
		slog.Error("rendering template", "template", name, "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

// handleRevoke handles POST /devices/{id}/revoke.
func (s *Server) handleRevoke(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	deviceID := extractPathParam(r.URL.Path, "/devices/", "/revoke")
	if deviceID == "" {
		http.Error(w, "invalid device ID", http.StatusBadRequest)
		return
	}

	device, err := s.store.GetDevice(deviceID)
	if err != nil {
		http.Error(w, "device not found", http.StatusNotFound)
		return
	}

	if err := device.Revoke(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := s.store.UpdateDevice(*device); err != nil {
		slog.Error("updating device", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	s.store.AddAuditEntry(model.NewAuditEntry(
		model.EventRevoked, device.ID,
		fmt.Sprintf("Device '%s' revoked", device.Name),
	))

	slog.Info("device revoked", "device_id", device.ID, "name", device.Name)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// handleRemoveDevice handles POST /devices/{id}/remove.
func (s *Server) handleRemoveDevice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	deviceID := extractPathParam(r.URL.Path, "/devices/", "/remove")
	if deviceID == "" {
		http.Error(w, "invalid device ID", http.StatusBadRequest)
		return
	}

	device, err := s.store.GetDevice(deviceID)
	if err != nil {
		http.Error(w, "device not found", http.StatusNotFound)
		return
	}

	if err := s.store.RemoveDevice(deviceID); err != nil {
		slog.Error("removing device", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.store.AddAuditEntry(model.NewAuditEntry(
		model.EventRevoked, device.ID,
		fmt.Sprintf("Device '%s' removed", device.Name),
	))

	slog.Info("device removed", "device_id", device.ID, "name", device.Name)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// ListenAndServe starts the HTTP server with graceful shutdown support.
func (s *Server) ListenAndServe() error {
	srv := &http.Server{
		Addr:    s.addr,
		Handler: s.mux,
	}

	// Graceful shutdown on SIGINT/SIGTERM
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	errCh := make(chan error, 1)
	go func() {
		if s.tlsCert != "" && s.tlsKey != "" {
			slog.Info("starting hub with TLS", "addr", s.addr)
			errCh <- srv.ListenAndServeTLS(s.tlsCert, s.tlsKey)
		} else {
			slog.Warn("running without TLS — use a reverse proxy or SSH tunnel for encrypted connections")
			slog.Info("starting hub", "addr", s.addr)
			errCh <- srv.ListenAndServe()
		}
	}()

	select {
	case err := <-errCh:
		return err
	case sig := <-sigCh:
		slog.Info("received signal, shutting down", "signal", sig)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			slog.Error("shutdown error", "error", err)
			return err
		}
		slog.Info("hub stopped gracefully")
		return nil
	}
}
