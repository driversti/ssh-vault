package model

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestDevice_Approve(t *testing.T) {
	tests := []struct {
		name    string
		device  Device
		wantErr bool
	}{
		{
			name:    "pending and verified",
			device:  Device{Status: StatusPending, Verified: true},
			wantErr: false,
		},
		{
			name:    "pending but not verified",
			device:  Device{Status: StatusPending, Verified: false},
			wantErr: true,
		},
		{
			name:    "already approved",
			device:  Device{Status: StatusApproved, Verified: true},
			wantErr: true,
		},
		{
			name:    "revoked",
			device:  Device{Status: StatusRevoked, Verified: true},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.device.Approve()
			if (err != nil) != tt.wantErr {
				t.Errorf("Approve() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if tt.device.Status != StatusApproved {
					t.Errorf("Status = %q, want %q", tt.device.Status, StatusApproved)
				}
				if tt.device.ApprovedAt == nil {
					t.Error("ApprovedAt should be set")
				}
			}
		})
	}
}

func TestDevice_Revoke(t *testing.T) {
	tests := []struct {
		name    string
		device  Device
		wantErr bool
	}{
		{
			name:    "approved device",
			device:  Device{Status: StatusApproved},
			wantErr: false,
		},
		{
			name:    "pending device",
			device:  Device{Status: StatusPending},
			wantErr: true,
		},
		{
			name:    "already revoked",
			device:  Device{Status: StatusRevoked},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.device.Revoke()
			if (err != nil) != tt.wantErr {
				t.Errorf("Revoke() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if tt.device.Status != StatusRevoked {
					t.Errorf("Status = %q, want %q", tt.device.Status, StatusRevoked)
				}
				if tt.device.RevokedAt == nil {
					t.Error("RevokedAt should be set")
				}
			}
		})
	}
}

func TestDevice_Validate(t *testing.T) {
	tests := []struct {
		name    string
		device  Device
		wantErr bool
	}{
		{
			name:    "valid device",
			device:  Device{Name: "my-laptop", PublicKey: "ssh-ed25519 AAAA..."},
			wantErr: false,
		},
		{
			name:    "empty name",
			device:  Device{Name: "", PublicKey: "ssh-ed25519 AAAA..."},
			wantErr: true,
		},
		{
			name:    "name too long",
			device:  Device{Name: strings.Repeat("a", 256), PublicKey: "ssh-ed25519 AAAA..."},
			wantErr: true,
		},
		{
			name:    "empty public key",
			device:  Device{Name: "my-laptop", PublicKey: ""},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.device.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDevice_JSONRoundTrip(t *testing.T) {
	d := Device{
		ID:        "test-id",
		Name:      "my-laptop",
		PublicKey: "ssh-ed25519 AAAA...",
		Status:    StatusPending,
		Verified:  true,
	}

	data, err := json.Marshal(d)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var got Device
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if got.ID != d.ID || got.Name != d.Name || got.Status != d.Status || got.Verified != d.Verified {
		t.Errorf("round-trip mismatch: got %+v, want %+v", got, d)
	}
}
