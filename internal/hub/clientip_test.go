package hub

import (
	"net"
	"net/http"
	"testing"
)

func TestClientIP(t *testing.T) {
	tests := []struct {
		name       string
		remoteAddr string
		xff        string
		want       string
	}{
		{
			name:       "public IP without XFF",
			remoteAddr: "203.0.113.5:12345",
			xff:        "",
			want:       "203.0.113.5",
		},
		{
			name:       "public IP with XFF ignored",
			remoteAddr: "203.0.113.5:12345",
			xff:        "10.0.0.1",
			want:       "203.0.113.5",
		},
		{
			name:       "Docker bridge IP trusts XFF",
			remoteAddr: "172.17.0.2:54321",
			xff:        "198.51.100.42",
			want:       "198.51.100.42",
		},
		{
			name:       "private 10.x trusts XFF",
			remoteAddr: "10.0.0.1:1234",
			xff:        "203.0.113.10",
			want:       "203.0.113.10",
		},
		{
			name:       "private 192.168.x trusts XFF",
			remoteAddr: "192.168.1.1:8080",
			xff:        "198.51.100.1",
			want:       "198.51.100.1",
		},
		{
			name:       "loopback trusts XFF",
			remoteAddr: "127.0.0.1:9999",
			xff:        "203.0.113.50",
			want:       "203.0.113.50",
		},
		{
			name:       "private IP without XFF returns host",
			remoteAddr: "172.17.0.2:54321",
			xff:        "",
			want:       "172.17.0.2",
		},
		{
			name:       "multiple XFF values takes first",
			remoteAddr: "172.17.0.2:54321",
			xff:        "203.0.113.1, 10.0.0.1, 172.17.0.1",
			want:       "203.0.113.1",
		},
		{
			name:       "XFF with spaces trimmed",
			remoteAddr: "10.0.0.1:1234",
			xff:        "  203.0.113.5 , 10.0.0.2",
			want:       "203.0.113.5",
		},
		{
			name:       "malformed remote addr returns as-is",
			remoteAddr: "not-an-addr",
			xff:        "203.0.113.1",
			want:       "not-an-addr",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &http.Request{
				RemoteAddr: tt.remoteAddr,
				Header:     http.Header{},
			}
			if tt.xff != "" {
				r.Header.Set("X-Forwarded-For", tt.xff)
			}
			got := clientIP(r)
			if got != tt.want {
				t.Errorf("clientIP() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIsPrivateIP(t *testing.T) {
	tests := []struct {
		ip   string
		want bool
	}{
		{"10.0.0.1", true},
		{"10.255.255.255", true},
		{"172.16.0.1", true},
		{"172.31.255.255", true},
		{"172.15.255.255", false},
		{"172.32.0.0", false},
		{"192.168.0.1", true},
		{"192.168.255.255", true},
		{"127.0.0.1", true},
		{"127.255.255.255", true},
		{"8.8.8.8", false},
		{"203.0.113.1", false},
		{"1.1.1.1", false},
	}

	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			ip := net.ParseIP(tt.ip)
			if got := isPrivateIP(ip); got != tt.want {
				t.Errorf("isPrivateIP(%s) = %v, want %v", tt.ip, got, tt.want)
			}
		})
	}
}
