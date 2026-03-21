package hub

import (
	"net"
	"net/http"
	"strings"
)

var privateRanges []net.IPNet

func init() {
	for _, cidr := range []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"127.0.0.0/8",
	} {
		_, network, _ := net.ParseCIDR(cidr)
		privateRanges = append(privateRanges, *network)
	}
}

// clientIP extracts the client IP from the request. When the request comes
// from a private IP (e.g., a reverse proxy in Docker), it trusts the
// X-Forwarded-For header. For direct public connections, it ignores XFF
// to prevent IP spoofing.
func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	ip := net.ParseIP(host)
	if ip != nil && isPrivateIP(ip) {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			parts := strings.SplitN(xff, ",", 2)
			return strings.TrimSpace(parts[0])
		}
	}

	return host
}

func isPrivateIP(ip net.IP) bool {
	for _, r := range privateRanges {
		if r.Contains(ip) {
			return true
		}
	}
	return false
}
