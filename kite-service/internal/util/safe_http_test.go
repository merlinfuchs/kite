package util

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsBlockedIP(t *testing.T) {
	tests := []struct {
		ip      string
		blocked bool
	}{
		// IPv4 Loopback
		{"127.0.0.1", true},
		{"127.0.0.2", true},
		{"127.255.255.254", true},

		// RFC 1918 Private
		{"10.0.0.1", true},
		{"10.254.254.254", true},
		{"172.16.0.1", true},
		{"172.31.255.255", true},
		{"192.168.0.1", true},
		{"192.168.1.100", true},

		// RFC 3927 Link-Local / Cloud Metadata
		{"169.254.169.254", true},
		{"169.254.1.1", true},

		// CGNAT
		{"100.64.0.1", true},
		{"100.127.255.254", true},

		// Current network / Unspecified
		{"0.0.0.0", true},
		{"0.1.2.3", true},

		// Multicast / Broadcast
		{"224.0.0.1", true},
		{"239.255.255.250", true},

		// IPv6 Loopback / Local / Private
		{"::1", true},
		{"::", true},
		{"fe80::1", true},
		{"fc00::1", true},
		{"fd12:3456:789a:1::1", true},

		// IPv4-mapped IPv6
		{"::ffff:127.0.0.1", true},
		{"::ffff:169.254.169.254", true},
		{"::ffff:10.0.0.1", true},
		{"::ffff:192.168.1.1", true},

		// Public Addresses
		{"8.8.8.8", false},
		{"1.1.1.1", false},
		{"93.184.216.34", false},
		{"2606:4700:4700::1111", false},
	}

	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			ip := net.ParseIP(tt.ip)
			require.NotNil(t, ip, "failed to parse test IP")
			assert.Equal(t, tt.blocked, IsBlockedIP(ip), "IP %s blocked status mismatch", tt.ip)
		})
	}
}

func TestSafeTransportBlocksLoopbackServer(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("internal secret content"))
	}))
	defer ts.Close()

	client := &http.Client{
		Timeout:   2 * time.Second,
		Transport: NewSafeTransport(),
	}

	_, err := client.Get(ts.URL)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "blocked: private or local address")
}
