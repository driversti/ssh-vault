package hub

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/driversti/ssh-vault/internal/model"
	"golang.org/x/crypto/ssh"
)

// enrollRequest is the JSON body for POST /api/enroll.
type enrollRequest struct {
	Token     string `json:"token"`
	PublicKey string `json:"public_key"`
	Name      string `json:"name"`
}

// enrollResponse is the JSON response for POST /api/enroll.
type enrollResponse struct {
	DeviceID  string `json:"device_id"`
	Challenge string `json:"challenge"`
}

// verifyRequest is the JSON body for POST /api/enroll/verify.
type verifyRequest struct {
	DeviceID  string `json:"device_id"`
	Signature string `json:"signature"`
}

// verifyResponse is the JSON response for POST /api/enroll/verify.
type verifyResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// keysResponse is the JSON response for GET /api/keys.
type keysResponse struct {
	Keys      []string  `json:"keys"`
	UpdatedAt time.Time `json:"updated_at"`
}

// handleEnroll handles POST /api/enroll.
func (s *Server) handleEnroll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req enrollRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	// Validate token
	tok, err := s.store.GetToken(req.Token)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid token"})
		return
	}
	if !tok.IsValid() {
		reason := "token expired"
		if tok.Used {
			reason = "token already used"
		}
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": reason})
		return
	}

	// Parse and validate SSH public key
	pubKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(req.PublicKey))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid SSH public key"})
		return
	}
	fingerprint := ssh.FingerprintSHA256(pubKey)

	// Generate device ID
	idBytes := make([]byte, 16)
	if _, err := rand.Read(idBytes); err != nil {
		slog.Error("generating device ID", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	deviceID := fmt.Sprintf("%x-%x-%x-%x-%x",
		idBytes[0:4], idBytes[4:6], idBytes[6:8], idBytes[8:10], idBytes[10:16])

	// Generate challenge
	challengeBytes := make([]byte, 32)
	if _, err := rand.Read(challengeBytes); err != nil {
		slog.Error("generating challenge", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	challenge := hex.EncodeToString(challengeBytes)

	// Create device
	device := model.Device{
		ID:          deviceID,
		Name:        req.Name,
		PublicKey:   req.PublicKey,
		Fingerprint: fingerprint,
		Status:      model.StatusPending,
		EnrolledAt:  time.Now().UTC(),
		Challenge:   challenge,
	}
	if err := device.Validate(); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	if err := s.store.AddDevice(device); err != nil {
		slog.Error("adding device", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// Mark token as used
	if err := s.store.UseToken(req.Token, deviceID); err != nil {
		slog.Error("marking token used", "error", err)
	}

	// Audit
	s.store.AddAuditEntry(model.NewAuditEntry(
		model.EventEnrolled, deviceID,
		fmt.Sprintf("Device '%s' enrolled via token %s...", req.Name, req.Token[:8]),
	))
	s.store.AddAuditEntry(model.NewAuditEntry(
		model.EventTokenUsed, deviceID,
		fmt.Sprintf("Token %s... used by device '%s'", req.Token[:8], req.Name),
	))

	slog.Info("device enrolled", "device_id", deviceID, "name", req.Name)
	writeJSON(w, http.StatusOK, enrollResponse{
		DeviceID:  deviceID,
		Challenge: challenge,
	})
}

// handleEnrollVerify handles POST /api/enroll/verify.
func (s *Server) handleEnrollVerify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req verifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	device, err := s.store.GetDevice(req.DeviceID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "device not found"})
		return
	}
	if device.Status != model.StatusPending {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "device not in pending state"})
		return
	}
	if device.Challenge == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "no challenge pending"})
		return
	}

	// Parse the stored public key
	pubKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(device.PublicKey))
	if err != nil {
		slog.Error("parsing stored public key", "error", err, "device_id", device.ID)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// Decode the signature from the request
	sigBytes, err := hex.DecodeString(req.Signature)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid signature encoding"})
		return
	}

	// Unmarshal the SSH signature
	sig := new(ssh.Signature)
	if err := ssh.Unmarshal(sigBytes, sig); err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid signature format"})
		return
	}

	// Verify the signature against the challenge
	challengeBytes, _ := hex.DecodeString(device.Challenge)
	if err := pubKey.Verify(challengeBytes, sig); err != nil {
		slog.Warn("signature verification failed", "device_id", device.ID, "error", err)
		s.store.AddAuditEntry(model.NewAuditEntry(
			model.EventAuthFailed, device.ID,
			fmt.Sprintf("Signature verification failed for device '%s'", device.Name),
		))
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "signature verification failed"})
		return
	}

	// Mark as verified, clear challenge
	device.Verified = true
	device.Challenge = ""
	if err := s.store.UpdateDevice(*device); err != nil {
		slog.Error("updating device", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	slog.Info("device verified", "device_id", device.ID, "name", device.Name)
	writeJSON(w, http.StatusOK, verifyResponse{
		Status:  "pending",
		Message: "Device registered. Awaiting approval.",
	})
}

// handleApprove handles POST /devices/{id}/approve.
func (s *Server) handleApprove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	deviceID := extractPathParam(r.URL.Path, "/devices/", "/approve")
	if deviceID == "" {
		http.Error(w, "invalid device ID", http.StatusBadRequest)
		return
	}

	device, err := s.store.GetDevice(deviceID)
	if err != nil {
		http.Error(w, "device not found", http.StatusNotFound)
		return
	}

	if !device.Verified {
		http.Error(w, "device not verified", http.StatusBadRequest)
		return
	}

	if err := device.Approve(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Generate API token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		slog.Error("generating API token", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	device.APIToken = hex.EncodeToString(tokenBytes)

	if err := s.store.UpdateDevice(*device); err != nil {
		slog.Error("updating device", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	s.store.AddAuditEntry(model.NewAuditEntry(
		model.EventApproved, device.ID,
		fmt.Sprintf("Device '%s' approved", device.Name),
	))

	slog.Info("device approved", "device_id", device.ID, "name", device.Name)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// handleKeys handles GET /api/keys.
func (s *Server) handleKeys(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get requesting device from context (set by auth middleware)
	device := deviceFromContext(r.Context())
	if device == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	// Update last sync timestamp
	now := time.Now().UTC()
	device.LastSyncAt = &now
	s.store.UpdateDevice(*device)

	// Collect approved devices' keys, excluding the requesting device
	approved := s.store.ListDevicesByStatus(model.StatusApproved)
	var keys []string
	for _, d := range approved {
		if d.ID != device.ID {
			keys = append(keys, d.PublicKey)
		}
	}
	sort.Strings(keys)

	writeJSON(w, http.StatusOK, keysResponse{
		Keys:      keys,
		UpdatedAt: now,
	})
}

// handleRemoveToken handles POST /tokens/{value}/remove.
func (s *Server) handleRemoveToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tokenValue := extractPathParam(r.URL.Path, "/tokens/", "/remove")
	if tokenValue == "" {
		http.Error(w, "invalid token value", http.StatusBadRequest)
		return
	}

	// Verify token exists and is unused before removing.
	tok, err := s.store.GetToken(tokenValue)
	if err != nil {
		http.Error(w, "token not found", http.StatusNotFound)
		return
	}
	if tok.Used {
		http.Error(w, "cannot remove used token", http.StatusBadRequest)
		return
	}

	if err := s.store.RemoveToken(tokenValue); err != nil {
		slog.Error("removing token", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// Audit the removal with truncated token prefix.
	s.store.AddAuditEntry(model.NewAuditEntry(
		model.EventTokenRemoved, "",
		fmt.Sprintf("Token %s... removed", tokenValue[:8]),
	))

	slog.Info("token removed", "token_prefix", tokenValue[:8])
	http.Redirect(w, r, "/tokens", http.StatusSeeOther)
}

// extractPathParam extracts a value between prefix and suffix in a URL path.
func extractPathParam(path, prefix, suffix string) string {
	if !strings.HasPrefix(path, prefix) {
		return ""
	}
	rest := strings.TrimPrefix(path, prefix)
	if suffix != "" {
		rest = strings.TrimSuffix(rest, suffix)
	}
	if rest == "" || strings.Contains(rest, "/") {
		return ""
	}
	return rest
}

// writeJSON writes a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
