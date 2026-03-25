package hub

import (
	"html/template"
	"strings"
	"testing"

	"golang.org/x/crypto/ssh"
)

// testKeys holds real SSH public keys for template function tests.
var testKeys = struct {
	ed25519UserHost string // comment: testuser@testhost
	rsaUserIP       string // comment: admin@192.168.1.100
	ed25519NoAt     string // comment: my-work-laptop (no @)
	ed25519Empty    string // no comment
}{
	ed25519UserHost: "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIGut3SUCwGfH3tC/Y369uKdKKqYdZpBpFF9mpcTUtr41 testuser@testhost",
	rsaUserIP:       "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC4xGIDPihbQmvMrLQ1jUl0Nc0RL62zp/mZaazsGs5ymZluSyRKGTpcXbRy8YkwzJk24HeqbWyNVK02ldK+6NPnJ8NUj+5nlL3qBDEze66iOAWOgDprKNY91988uqwsm0Sn9272wjPnZ4Mh8TDfFJxAeeGtY6cABfwUfsm8WjA3VJNKEdDptKnSSh3cSI+mNXdKVxYyKgsn35v4xJUdxcD7+05PpeDVe5x8BOc4FhxpfxaQZrI5rMY4mTQIJCC2+Vgr7N84eckbyvE9Q+o8oMbIPCROQtl/ZXpL3y1GPiFQGeoe8O+P76rWcud6uxpFrio4uT6FKdLK/8QjcPgAepg1 admin@192.168.1.100",
}

func init() {
	// Build keys with modified comments from the ed25519 base key.
	// Parse the base key, then reconstruct with different comments.
	key, _, _, _, err := ssh.ParseAuthorizedKey([]byte(testKeys.ed25519UserHost))
	if err != nil {
		panic("failed to parse test ed25519 key: " + err.Error())
	}
	base := string(ssh.MarshalAuthorizedKey(key))
	base = strings.TrimSpace(base)
	testKeys.ed25519NoAt = base + " my-work-laptop"
	testKeys.ed25519Empty = base
}

// buildFuncMap creates the same template.FuncMap used by NewServer,
// extracting just the key-related functions for testing.
func buildFuncMap() template.FuncMap {
	return template.FuncMap{
		"keyUser": func(publicKey string) string {
			_, comment, _, _, err := ssh.ParseAuthorizedKey([]byte(publicKey))
			if err != nil || comment == "" {
				return ""
			}
			user, _, found := strings.Cut(comment, "@")
			if !found {
				return comment
			}
			return user
		},
		"keyHost": func(publicKey string) string {
			_, comment, _, _, err := ssh.ParseAuthorizedKey([]byte(publicKey))
			if err != nil || comment == "" {
				return ""
			}
			_, host, found := strings.Cut(comment, "@")
			if !found {
				return ""
			}
			return host
		},
		"keyType": func(publicKey string) string {
			key, _, _, _, err := ssh.ParseAuthorizedKey([]byte(publicKey))
			if err != nil {
				return ""
			}
			return strings.TrimPrefix(key.Type(), "ssh-")
		},
	}
}

func TestKeyUser(t *testing.T) {
	fm := buildFuncMap()
	keyUser := fm["keyUser"].(func(string) string)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"user@host comment", testKeys.ed25519UserHost, "testuser"},
		{"user@IP comment", testKeys.rsaUserIP, "admin"},
		{"freeform comment (no @)", testKeys.ed25519NoAt, "my-work-laptop"},
		{"empty comment", testKeys.ed25519Empty, ""},
		{"malformed key", "not-a-valid-key", ""},
		{"empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := keyUser(tt.input)
			if got != tt.expected {
				t.Errorf("keyUser() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestKeyHost(t *testing.T) {
	fm := buildFuncMap()
	keyHost := fm["keyHost"].(func(string) string)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"user@host comment", testKeys.ed25519UserHost, "testhost"},
		{"user@IP comment", testKeys.rsaUserIP, "192.168.1.100"},
		{"freeform comment (no @)", testKeys.ed25519NoAt, ""},
		{"empty comment", testKeys.ed25519Empty, ""},
		{"malformed key", "not-a-valid-key", ""},
		{"empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := keyHost(tt.input)
			if got != tt.expected {
				t.Errorf("keyHost() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestKeyType(t *testing.T) {
	fm := buildFuncMap()
	keyType := fm["keyType"].(func(string) string)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"ed25519 key", testKeys.ed25519UserHost, "ed25519"},
		{"RSA key", testKeys.rsaUserIP, "rsa"},
		{"ed25519 no comment", testKeys.ed25519Empty, "ed25519"},
		{"malformed key", "not-a-valid-key", ""},
		{"empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := keyType(tt.input)
			if got != tt.expected {
				t.Errorf("keyType() = %q, want %q", got, tt.expected)
			}
		})
	}
}
