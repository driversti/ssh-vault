package model

import "time"

// ShortCode represents a short-lived enrollment code that maps to an enrollment token.
type ShortCode struct {
	Code       string     `json:"code"`
	TokenValue string     `json:"token_value"`
	CreatedAt  time.Time  `json:"created_at"`
	ExpiresAt  time.Time  `json:"expires_at"`
	Used       bool       `json:"used"`
	UsedAt     *time.Time `json:"used_at,omitempty"`
	UsedByIP   string     `json:"used_by_ip,omitempty"`
}

// NewShortCode creates a new ShortCode with the given code, linked token value, and TTL.
func NewShortCode(code, tokenValue string, ttl time.Duration) *ShortCode {
	now := time.Now().UTC()
	return &ShortCode{
		Code:       code,
		TokenValue: tokenValue,
		CreatedAt:  now,
		ExpiresAt:  now.Add(ttl),
	}
}

// IsValid returns true if the short code has not been used and has not expired.
func (sc *ShortCode) IsValid() bool {
	return !sc.Used && !sc.IsExpired()
}

// IsExpired returns true if the current time is at or past ExpiresAt.
func (sc *ShortCode) IsExpired() bool {
	return !time.Now().UTC().Before(sc.ExpiresAt)
}

// MarkUsed marks the short code as used by the given IP address.
func (sc *ShortCode) MarkUsed(ip string) {
	now := time.Now().UTC()
	sc.Used = true
	sc.UsedAt = &now
	sc.UsedByIP = ip
}
