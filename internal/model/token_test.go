package model

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNewToken(t *testing.T) {
	tok, err := NewToken(24 * time.Hour)
	if err != nil {
		t.Fatalf("NewToken: %v", err)
	}

	if len(tok.Value) != 64 { // 32 bytes = 64 hex chars
		t.Errorf("Value length = %d, want 64", len(tok.Value))
	}
	if tok.Used {
		t.Error("new token should not be used")
	}
	if tok.ExpiresAt.Before(tok.CreatedAt) {
		t.Error("ExpiresAt should be after CreatedAt")
	}
}

func TestToken_IsValid(t *testing.T) {
	tests := []struct {
		name string
		tok  Token
		want bool
	}{
		{
			name: "fresh token",
			tok: Token{
				ExpiresAt: time.Now().UTC().Add(1 * time.Hour),
				Used:      false,
			},
			want: true,
		},
		{
			name: "used token",
			tok: Token{
				ExpiresAt: time.Now().UTC().Add(1 * time.Hour),
				Used:      true,
			},
			want: false,
		},
		{
			name: "expired token",
			tok: Token{
				ExpiresAt: time.Now().UTC().Add(-1 * time.Second),
				Used:      false,
			},
			want: false,
		},
		{
			name: "used and expired",
			tok: Token{
				ExpiresAt: time.Now().UTC().Add(-1 * time.Second),
				Used:      true,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tok.IsValid(); got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToken_IsExpired(t *testing.T) {
	tests := []struct {
		name string
		tok  Token
		want bool
	}{
		{
			name: "future expiry",
			tok:  Token{ExpiresAt: time.Now().UTC().Add(1 * time.Hour)},
			want: false,
		},
		{
			name: "past expiry",
			tok:  Token{ExpiresAt: time.Now().UTC().Add(-1 * time.Second)},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tok.IsExpired(); got != tt.want {
				t.Errorf("IsExpired() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToken_JSONRoundTrip(t *testing.T) {
	tok, err := NewToken(1 * time.Hour)
	if err != nil {
		t.Fatalf("NewToken: %v", err)
	}

	data, err := json.Marshal(tok)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var got Token
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if got.Value != tok.Value || got.Used != tok.Used {
		t.Errorf("round-trip mismatch: got %+v, want %+v", got, tok)
	}
}

func TestNewToken_Uniqueness(t *testing.T) {
	tok1, err := NewToken(1 * time.Hour)
	if err != nil {
		t.Fatalf("NewToken: %v", err)
	}
	tok2, err := NewToken(1 * time.Hour)
	if err != nil {
		t.Fatalf("NewToken: %v", err)
	}
	if tok1.Value == tok2.Value {
		t.Error("two tokens should have different values")
	}
}
