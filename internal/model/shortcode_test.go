package model

import (
	"testing"
	"time"
)

func TestNewShortCode(t *testing.T) {
	sc := NewShortCode("123456", "abc123", 15*time.Minute)

	if sc.Code != "123456" {
		t.Errorf("Code = %q, want %q", sc.Code, "123456")
	}
	if sc.TokenValue != "abc123" {
		t.Errorf("TokenValue = %q, want %q", sc.TokenValue, "abc123")
	}
	if sc.Used {
		t.Error("new short code should not be used")
	}
	if sc.UsedAt != nil {
		t.Error("new short code should have nil UsedAt")
	}
	if sc.UsedByIP != "" {
		t.Error("new short code should have empty UsedByIP")
	}
	if sc.ExpiresAt.Before(sc.CreatedAt) {
		t.Error("ExpiresAt should be after CreatedAt")
	}
	expectedExpiry := sc.CreatedAt.Add(15 * time.Minute)
	if sc.ExpiresAt.Sub(expectedExpiry) > time.Second {
		t.Errorf("ExpiresAt = %v, want ~%v", sc.ExpiresAt, expectedExpiry)
	}
}

func TestShortCode_IsValid(t *testing.T) {
	tests := []struct {
		name string
		sc   ShortCode
		want bool
	}{
		{
			name: "valid code",
			sc: ShortCode{
				ExpiresAt: time.Now().UTC().Add(10 * time.Minute),
				Used:      false,
			},
			want: true,
		},
		{
			name: "used code",
			sc: ShortCode{
				ExpiresAt: time.Now().UTC().Add(10 * time.Minute),
				Used:      true,
			},
			want: false,
		},
		{
			name: "expired code",
			sc: ShortCode{
				ExpiresAt: time.Now().UTC().Add(-1 * time.Minute),
				Used:      false,
			},
			want: false,
		},
		{
			name: "used and expired",
			sc: ShortCode{
				ExpiresAt: time.Now().UTC().Add(-1 * time.Minute),
				Used:      true,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.sc.IsValid(); got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestShortCode_IsExpired(t *testing.T) {
	tests := []struct {
		name string
		sc   ShortCode
		want bool
	}{
		{
			name: "not expired",
			sc:   ShortCode{ExpiresAt: time.Now().UTC().Add(10 * time.Minute)},
			want: false,
		},
		{
			name: "expired",
			sc:   ShortCode{ExpiresAt: time.Now().UTC().Add(-1 * time.Minute)},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.sc.IsExpired(); got != tt.want {
				t.Errorf("IsExpired() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestShortCode_MarkUsed(t *testing.T) {
	sc := NewShortCode("654321", "def456", 15*time.Minute)
	sc.MarkUsed("192.168.1.100")

	if !sc.Used {
		t.Error("MarkUsed should set Used to true")
	}
	if sc.UsedAt == nil {
		t.Fatal("MarkUsed should set UsedAt")
	}
	if sc.UsedByIP != "192.168.1.100" {
		t.Errorf("UsedByIP = %q, want %q", sc.UsedByIP, "192.168.1.100")
	}
	if !sc.IsExpired() && sc.IsValid() {
		t.Error("used short code should not be valid")
	}
}
