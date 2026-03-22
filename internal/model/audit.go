package model

import "time"

const (
	EventEnrolled         = "enrolled"
	EventApproved         = "approved"
	EventRevoked          = "revoked"
	EventAuthFailed       = "auth_failed"
	EventTokenUsed        = "token_used"
	EventTokenRemoved     = "token_removed"
	EventShortCodeCreated = "shortcode_created"
	EventShortCodeUsed    = "shortcode_used"
	EventShortCodeExpired = "shortcode_expired"
	EventDeviceRenamed    = "device_renamed"
)

// AuditEntry records a notable event in the system.
type AuditEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Event     string    `json:"event"`
	DeviceID  string    `json:"device_id"`
	Details   string    `json:"details"`
}

// NewAuditEntry creates an AuditEntry with the current UTC timestamp.
func NewAuditEntry(event, deviceID, details string) AuditEntry {
	return AuditEntry{
		Timestamp: time.Now().UTC(),
		Event:     event,
		DeviceID:  deviceID,
		Details:   details,
	}
}
