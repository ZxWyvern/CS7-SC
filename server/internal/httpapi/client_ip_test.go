package httpapi

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"cs7sc/server/internal/config"
)

func requestWithXFF(xff string, remoteAddr string) *http.Request {
	request := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	request.RemoteAddr = remoteAddr
	if xff != "" {
		request.Header.Set("X-Forwarded-For", xff)
	}
	return request
}

func TestClientIP(t *testing.T) {
	cases := []struct {
		name       string
		hops       int
		forwarded  string
		remoteAddr string
		want       string
	}{
		{name: "tanpa hop pakai koneksi langsung", hops: 0, forwarded: "203.0.113.9", remoteAddr: "198.51.100.7:443", want: "198.51.100.7"},
		{name: "satu hop pakai entri kanan", hops: 1, forwarded: "203.0.113.9", remoteAddr: "10.0.0.1:8080", want: "203.0.113.9"},
		{name: "satu hop abaikan bumbuan penyerang", hops: 1, forwarded: "203.0.113.9,198.51.100.7", remoteAddr: "10.0.0.1:8080", want: "198.51.100.7"},
		{name: "satu hop buang spasi", hops: 1, forwarded: "  198.51.100.7  ", remoteAddr: "10.0.0.1:8080", want: "198.51.100.7"},
		{name: "header kosong jatuh ke koneksi langsung", hops: 1, forwarded: "", remoteAddr: "198.51.100.7:443", want: "198.51.100.7"},
		{name: "entri bukan IP jatuh ke koneksi langsung", hops: 1, forwarded: "bukan-alamat", remoteAddr: "198.51.100.7:443", want: "198.51.100.7"},
		{name: "hop lebih banyak dari entri jatuh ke koneksi langsung", hops: 2, forwarded: "203.0.113.9", remoteAddr: "198.51.100.7:443", want: "198.51.100.7"},
		{name: "IPv6 dipotong ke /64", hops: 1, forwarded: "2001:db8:1:2:3:4:5:6", remoteAddr: "10.0.0.1:8080", want: "2001:db8:1:2::"},
		{name: "IPv4 tidak dipotong", hops: 1, forwarded: "203.0.113.9", remoteAddr: "10.0.0.1:8080", want: "203.0.113.9"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			server := &Server{cfg: config.Config{TrustedProxyHops: tc.hops}}

			got := server.clientIP(requestWithXFF(tc.forwarded, tc.remoteAddr))
			if got != tc.want {
				t.Fatalf("clientIP = %q, mau %q", got, tc.want)
			}
		})
	}
}

func TestMaskIPv6To64(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{in: "2001:db8:1:2:3:4:5:6", want: "2001:db8:1:2::"},
		{in: "2001:db8:1:2::", want: "2001:db8:1:2::"},
		{in: "203.0.113.9", want: "203.0.113.9"},
		{in: "::ffff:203.0.113.9", want: "203.0.113.9"},
		{in: "::1", want: "::"},
	}

	for _, tc := range cases {
		if got := maskIPv6To64(net.ParseIP(tc.in)); got != tc.want {
			t.Fatalf("maskIPv6To64(%q) = %q, mau %q", tc.in, got, tc.want)
		}
	}
}
