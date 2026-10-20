package util

import (
	"fmt"
	"net"
	"net/http"
	"syscall"
	"time"
)

var (
	// cgnatNet is 100.64.0.0/10 (RFC 6598 - Carrier-Grade NAT).
	cgnatNet = &net.IPNet{
		IP:   net.IPv4(100, 64, 0, 0),
		Mask: net.CIDRMask(10, 32),
	}
	// linkLocalNet is 169.254.0.0/16 (RFC 3927 - Link-Local / Cloud Metadata).
	linkLocalNet = &net.IPNet{
		IP:   net.IPv4(169, 254, 0, 0),
		Mask: net.CIDRMask(16, 32),
	}
)

// IsBlockedIP returns true if the IP belongs to loopback, private (RFC 1918),
// link-local, cloud metadata, multicast, unspecified, or CGNAT address spaces.
func IsBlockedIP(ip net.IP) bool {
	if ip == nil {
		return true
	}

	// Normalize IPv4-mapped IPv6 addresses (e.g. ::ffff:127.0.0.1)
	if ip4 := ip.To4(); ip4 != nil {
		ip = ip4
	}

	if ip.IsLoopback() ||
		ip.IsPrivate() ||
		ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() ||
		ip.IsInterfaceLocalMulticast() ||
		ip.IsMulticast() ||
		ip.IsUnspecified() {
		return true
	}

	if cgnatNet.Contains(ip) || linkLocalNet.Contains(ip) {
		return true
	}

	// 0.0.0.0/8 (Current network / "this" network)
	if ip4 := ip.To4(); ip4 != nil && ip4[0] == 0 {
		return true
	}

	return false
}

// NewSafeTransport returns an http.Transport configured with a socket Control
// hook that blocks connections to loopback, RFC 1918, and cloud metadata addresses.
func NewSafeTransport() *http.Transport {
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		Control: func(network, address string, c syscall.RawConn) error {
			host, _, err := net.SplitHostPort(address)
			if err != nil {
				return err
			}

			ip := net.ParseIP(host)
			if IsBlockedIP(ip) {
				return fmt.Errorf("connection to %s is blocked: private or local address", host)
			}

			return nil
		},
	}

	return &http.Transport{
		Proxy:                 nil,
		DialContext:           dialer.DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}
