package hub

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/driversti/ssh-vault/internal/model"
	"golang.org/x/crypto/ssh"
)

// testServer creates a Server with a temporary file store for testing.
func testServer(t *testing.T) *Server {
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

// generateTestKey creates an ed25519 SSH key pair for testing.
func generateTestKey(t *testing.T) (ssh.Signer, string) {
	t.Helper()
	_, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	signer, err := ssh.NewSignerFromKey(privKey)
	if err != nil {
		t.Fatalf("NewSignerFromKey: %v", err)
	}
	pubKeyStr := strings.TrimSpace(string(ssh.MarshalAuthorizedKey(signer.PublicKey())))
	return signer, pubKeyStr
}

// addSessionCookie creates a session and adds the cookie to the request.
func addSessionCookie(t *testing.T, s *Server, req *http.Request) {
	t.Helper()
	token, err := s.sessions.create()
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: token})
}

// addValidToken adds a valid token to the store and returns its value.
func addValidToken(t *testing.T, s *Server) string {
	t.Helper()
	tok, err := model.NewToken(24 * time.Hour)
	if err != nil {
		t.Fatalf("NewToken: %v", err)
	}
	if err := s.store.AddToken(*tok); err != nil {
		t.Fatalf("AddToken: %v", err)
	}
	return tok.Value
}

func TestHandleEnroll_ValidToken(t *testing.T) {
	s := testServer(t)
	_, pubKey := generateTestKey(t)
	token := addValidToken(t, s)

	body, _ := json.Marshal(map[string]string{
		"token":      token,
		"public_key": pubKey,
		"name":       "test-device",
	})

	req := httptest.NewRequest("POST", "/api/enroll", bytes.NewReader(body))
	w := httptest.NewRecorder()
	s.handleEnroll(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", w.Code, w.Body.String())
	}

	var resp enrollResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.DeviceID == "" {
		t.Error("DeviceID should not be empty")
	}
	if resp.Challenge == "" {
		t.Error("Challenge should not be empty")
	}
}

func TestHandleEnroll_ExpiredToken(t *testing.T) {
	s := testServer(t)
	_, pubKey := generateTestKey(t)

	// Create expired token
	tok := model.Token{
		Value:     "expired-token",
		CreatedAt: time.Now().UTC().Add(-48 * time.Hour),
		ExpiresAt: time.Now().UTC().Add(-24 * time.Hour),
	}
	s.store.AddToken(tok)

	body, _ := json.Marshal(map[string]string{
		"token":      "expired-token",
		"public_key": pubKey,
		"name":       "test-device",
	})

	req := httptest.NewRequest("POST", "/api/enroll", bytes.NewReader(body))
	w := httptest.NewRecorder()
	s.handleEnroll(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

func TestHandleEnroll_UsedToken(t *testing.T) {
	s := testServer(t)
	_, pubKey := generateTestKey(t)

	tok := model.Token{
		Value:     "used-token",
		CreatedAt: time.Now().UTC(),
		ExpiresAt: time.Now().UTC().Add(24 * time.Hour),
		Used:      true,
		UsedBy:    "some-device",
	}
	s.store.AddToken(tok)

	body, _ := json.Marshal(map[string]string{
		"token":      "used-token",
		"public_key": pubKey,
		"name":       "test-device",
	})

	req := httptest.NewRequest("POST", "/api/enroll", bytes.NewReader(body))
	w := httptest.NewRecorder()
	s.handleEnroll(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

func TestHandleEnroll_InvalidSSHKey(t *testing.T) {
	s := testServer(t)
	token := addValidToken(t, s)

	body, _ := json.Marshal(map[string]string{
		"token":      token,
		"public_key": "not-a-valid-key",
		"name":       "test-device",
	})

	req := httptest.NewRequest("POST", "/api/enroll", bytes.NewReader(body))
	w := httptest.NewRecorder()
	s.handleEnroll(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestHandleEnrollVerify_ValidSignature(t *testing.T) {
	s := testServer(t)
	signer, pubKey := generateTestKey(t)
	token := addValidToken(t, s)

	// Enroll first
	enrollBody, _ := json.Marshal(map[string]string{
		"token":      token,
		"public_key": pubKey,
		"name":       "test-device",
	})
	req := httptest.NewRequest("POST", "/api/enroll", bytes.NewReader(enrollBody))
	w := httptest.NewRecorder()
	s.handleEnroll(w, req)

	var enrollResp enrollResponse
	json.Unmarshal(w.Body.Bytes(), &enrollResp)

	// Sign the challenge
	challengeBytes, _ := hex.DecodeString(enrollResp.Challenge)
	sig, err := signer.Sign(rand.Reader, challengeBytes)
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}
	sigHex := hex.EncodeToString(ssh.Marshal(sig))

	// Verify
	verifyBody, _ := json.Marshal(map[string]string{
		"device_id": enrollResp.DeviceID,
		"signature": sigHex,
	})
	req2 := httptest.NewRequest("POST", "/api/enroll/verify", bytes.NewReader(verifyBody))
	w2 := httptest.NewRecorder()
	s.handleEnrollVerify(w2, req2)

	if w2.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", w2.Code, w2.Body.String())
	}

	// Verify device is now verified
	device, _ := s.store.GetDevice(enrollResp.DeviceID)
	if !device.Verified {
		t.Error("device should be verified")
	}
	if device.Challenge != "" {
		t.Error("challenge should be cleared")
	}
}

func TestHandleEnrollVerify_InvalidSignature(t *testing.T) {
	s := testServer(t)
	_, pubKey := generateTestKey(t)
	token := addValidToken(t, s)

	// Enroll first
	enrollBody, _ := json.Marshal(map[string]string{
		"token":      token,
		"public_key": pubKey,
		"name":       "test-device",
	})
	req := httptest.NewRequest("POST", "/api/enroll", bytes.NewReader(enrollBody))
	w := httptest.NewRecorder()
	s.handleEnroll(w, req)

	var enrollResp enrollResponse
	json.Unmarshal(w.Body.Bytes(), &enrollResp)

	// Send wrong signature
	verifyBody, _ := json.Marshal(map[string]string{
		"device_id": enrollResp.DeviceID,
		"signature": hex.EncodeToString([]byte("wrong-signature")),
	})
	req2 := httptest.NewRequest("POST", "/api/enroll/verify", bytes.NewReader(verifyBody))
	w2 := httptest.NewRecorder()
	s.handleEnrollVerify(w2, req2)

	if w2.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w2.Code)
	}
}

func TestHandleKeys_ValidBearer(t *testing.T) {
	s := testServer(t)

	// Add two approved devices
	d1 := model.Device{
		ID:        "dev-1",
		Name:      "device-1",
		PublicKey: "ssh-ed25519 AAAA1 device-1",
		Status:    model.StatusApproved,
		APIToken:  "token-1",
	}
	d2 := model.Device{
		ID:        "dev-2",
		Name:      "device-2",
		PublicKey: "ssh-ed25519 AAAA2 device-2",
		Status:    model.StatusApproved,
		APIToken:  "token-2",
	}
	s.store.AddDevice(d1)
	s.store.AddDevice(d2)

	// Request as device-1 — should get device-2's key only
	req := httptest.NewRequest("GET", "/api/keys", nil)
	req.Header.Set("Authorization", "Bearer token-1")
	w := httptest.NewRecorder()

	// Use the mux to test through middleware
	s.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", w.Code, w.Body.String())
	}

	var resp keysResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if len(resp.Keys) != 1 {
		t.Fatalf("expected 1 key, got %d", len(resp.Keys))
	}
	if resp.Keys[0] != "ssh-ed25519 AAAA2 device-2" {
		t.Errorf("key = %q, want device-2's key", resp.Keys[0])
	}
}

func TestHandleKeys_RevokedToken(t *testing.T) {
	s := testServer(t)

	d := model.Device{
		ID:       "dev-1",
		Name:     "device-1",
		Status:   model.StatusRevoked,
		APIToken: "revoked-token",
	}
	s.store.AddDevice(d)

	req := httptest.NewRequest("GET", "/api/keys", nil)
	req.Header.Set("Authorization", "Bearer revoked-token")
	w := httptest.NewRecorder()
	s.mux.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

func TestHandleKeys_UnknownToken(t *testing.T) {
	s := testServer(t)

	req := httptest.NewRequest("GET", "/api/keys", nil)
	req.Header.Set("Authorization", "Bearer unknown-token")
	w := httptest.NewRecorder()
	s.mux.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

func TestHandleKeys_NoBearer(t *testing.T) {
	s := testServer(t)

	req := httptest.NewRequest("GET", "/api/keys", nil)
	w := httptest.NewRecorder()
	s.mux.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

// === Revocation Tests (T025a) ===

func TestHandleRevoke_ApprovedDevice(t *testing.T) {
	s := testServer(t)

	d := model.Device{
		ID:       "dev-1",
		Name:     "device-1",
		Status:   model.StatusApproved,
		APIToken: "token-1",
	}
	s.store.AddDevice(d)

	req := httptest.NewRequest("POST", "/devices/dev-1/revoke", nil)
	addSessionCookie(t, s, req)
	w := httptest.NewRecorder()
	s.mux.ServeHTTP(w, req)

	if w.Code != http.StatusSeeOther {
		t.Errorf("status = %d, want 303; body: %s", w.Code, w.Body.String())
	}

	device, _ := s.store.GetDevice("dev-1")
	if device.Status != model.StatusRevoked {
		t.Errorf("status = %q, want revoked", device.Status)
	}
}

func TestHandleRevoke_PendingDevice(t *testing.T) {
	s := testServer(t)

	d := model.Device{
		ID:     "dev-1",
		Name:   "device-1",
		Status: model.StatusPending,
	}
	s.store.AddDevice(d)

	req := httptest.NewRequest("POST", "/devices/dev-1/revoke", nil)
	addSessionCookie(t, s, req)
	w := httptest.NewRecorder()
	s.mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestHandleRevoke_RevokedDeviceExcludedFromKeys(t *testing.T) {
	s := testServer(t)

	// Three approved devices
	for i := 1; i <= 3; i++ {
		d := model.Device{
			ID:        fmt.Sprintf("dev-%d", i),
			Name:      fmt.Sprintf("device-%d", i),
			PublicKey: fmt.Sprintf("ssh-ed25519 AAAA%d device-%d", i, i),
			Status:    model.StatusApproved,
			APIToken:  fmt.Sprintf("token-%d", i),
		}
		s.store.AddDevice(d)
	}

	// Revoke device 2
	req := httptest.NewRequest("POST", "/devices/dev-2/revoke", nil)
	addSessionCookie(t, s, req)
	w := httptest.NewRecorder()
	s.mux.ServeHTTP(w, req)

	// Device 1 syncs — should only see device 3's key
	req2 := httptest.NewRequest("GET", "/api/keys", nil)
	req2.Header.Set("Authorization", "Bearer token-1")
	w2 := httptest.NewRecorder()
	s.mux.ServeHTTP(w2, req2)

	var resp keysResponse
	json.Unmarshal(w2.Body.Bytes(), &resp)
	if len(resp.Keys) != 1 {
		t.Fatalf("expected 1 key, got %d: %v", len(resp.Keys), resp.Keys)
	}
	if !strings.Contains(resp.Keys[0], "AAAA3") {
		t.Errorf("expected device-3 key, got %q", resp.Keys[0])
	}
}

func TestRevokedDevice_Gets401OnSync(t *testing.T) {
	s := testServer(t)

	d := model.Device{
		ID:       "dev-1",
		Name:     "device-1",
		Status:   model.StatusRevoked,
		APIToken: "token-1",
	}
	s.store.AddDevice(d)

	req := httptest.NewRequest("GET", "/api/keys", nil)
	req.Header.Set("Authorization", "Bearer token-1")
	w := httptest.NewRecorder()
	s.mux.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}

	var errResp map[string]string
	json.Unmarshal(w.Body.Bytes(), &errResp)
	if errResp["error"] != "device revoked" {
		t.Errorf("error = %q, want 'device revoked'", errResp["error"])
	}
}
