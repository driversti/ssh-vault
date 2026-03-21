package model

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// Token represents a one-time enrollment token.
type Token struct {
	Value     string    `json:"value"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	UsedBy    string    `json:"used_by"`
	Used      bool      `json:"used"`
}

// NewToken generates a cryptographically random 32-byte hex token
// that expires after the given duration.
func NewToken(expiry time.Duration) (*Token, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return nil, fmt.Errorf("generating random token: %w", err)
	}
	now := time.Now().UTC()
	return &Token{
		Value:     hex.EncodeToString(b),
		CreatedAt: now,
		ExpiresAt: now.Add(expiry),
	}, nil
}

// IsValid returns true if the token has not been used and has not expired.
func (t *Token) IsValid() bool {
	return !t.Used && !t.IsExpired()
}

// IsExpired returns true if the current time is at or past ExpiresAt.
func (t *Token) IsExpired() bool {
	return !time.Now().UTC().Before(t.ExpiresAt)
}
