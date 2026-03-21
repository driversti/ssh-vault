package model

import (
	"errors"
	"fmt"
	"time"
)

const (
	StatusPending  = "pending"
	StatusApproved = "approved"
	StatusRevoked  = "revoked"
)

// Device represents a registered machine that syncs SSH keys.
type Device struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	PublicKey   string     `json:"public_key"`
	Fingerprint string     `json:"fingerprint"`
	Status      string     `json:"status"`
	APIToken    string     `json:"api_token"`
	EnrolledAt  time.Time  `json:"enrolled_at"`
	ApprovedAt  *time.Time `json:"approved_at,omitempty"`
	RevokedAt   *time.Time `json:"revoked_at,omitempty"`
	LastSyncAt  *time.Time `json:"last_sync_at,omitempty"`
	Challenge   string     `json:"challenge"`
	Verified    bool       `json:"verified"`
}

// Approve transitions the device from pending+verified to approved.
func (d *Device) Approve() error {
	if d.Status != StatusPending {
		return fmt.Errorf("cannot approve device with status %q: must be %q", d.Status, StatusPending)
	}
	if !d.Verified {
		return errors.New("cannot approve device: not verified")
	}
	now := time.Now().UTC()
	d.Status = StatusApproved
	d.ApprovedAt = &now
	return nil
}

// Revoke transitions the device from approved to revoked.
func (d *Device) Revoke() error {
	if d.Status != StatusApproved {
		return fmt.Errorf("cannot revoke device with status %q: must be %q", d.Status, StatusApproved)
	}
	now := time.Now().UTC()
	d.Status = StatusRevoked
	d.RevokedAt = &now
	return nil
}

// Validate checks that the device has the required fields.
func (d *Device) Validate() error {
	if d.Name == "" {
		return errors.New("device name must not be empty")
	}
	if len(d.Name) > 255 {
		return fmt.Errorf("device name must not exceed 255 characters, got %d", len(d.Name))
	}
	if d.PublicKey == "" {
		return errors.New("device public key must not be empty")
	}
	return nil
}
