package main

import (
	"bytes"
	"net"
	"testing"
)

// Tests for createTCPSynHeader with both IPv4 and IPv6 addresses.
func TestCreateTCPSynHeader(t *testing.T) {
	srcIPv4 := net.ParseIP("192.168.1.1").To4()
	dstIPv4 := net.ParseIP("8.8.8.8").To4()
	srcIPv6 := net.ParseIP("::1")
	dstIPv6 := net.ParseIP("2001:db8::1")

	// IPv4 tests
	t.Run("IPv4: returns non-empty bytes", func(t *testing.T) {
		pkt, err := createTCPSynHeader(srcIPv4, dstIPv4, 12345, 80)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(pkt) == 0 {
			t.Error("returned empty bytes")
		}
	})

	t.Run("IPv4: different source ports produce different output", func(t *testing.T) {
		pkt1, err1 := createTCPSynHeader(srcIPv4, dstIPv4, 1111, 80)
		pkt2, err2 := createTCPSynHeader(srcIPv4, dstIPv4, 2222, 80)
		if err1 != nil || err2 != nil {
			t.Fatalf("unexpected errors: %v, %v", err1, err2)
		}
		if bytes.Equal(pkt1, pkt2) {
			t.Error("headers with different source ports should differ")
		}
	})

	t.Run("IPv4: header length is at least 20 bytes", func(t *testing.T) {
		pkt, err := createTCPSynHeader(srcIPv4, dstIPv4, 12345, 80)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(pkt) < 20 {
			t.Errorf("header too short: got %d bytes, want at least 20", len(pkt))
		}
	})

	t.Run("IPv4: minimum port 1 is accepted", func(t *testing.T) {
		_, err := createTCPSynHeader(srcIPv4, dstIPv4, 1024, 1)
		if err != nil {
			t.Fatalf("port 1 should be valid, got error: %v", err)
		}
	})

	t.Run("IPv4: maximum port 65535 is accepted", func(t *testing.T) {
		_, err := createTCPSynHeader(srcIPv4, dstIPv4, 1024, 65535)
		if err != nil {
			t.Fatalf("port 65535 should be valid, got error: %v", err)
		}
	})

	// IPv6 tests
	t.Run("IPv6: returns non-empty bytes", func(t *testing.T) {
		pkt, err := createTCPSynHeader(srcIPv6, dstIPv6, 12345, 80)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(pkt) == 0 {
			t.Error("returned empty bytes")
		}
	})

	t.Run("IPv6: header length is at least 20 bytes", func(t *testing.T) {
		pkt, err := createTCPSynHeader(srcIPv6, dstIPv6, 12345, 443)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(pkt) < 20 {
			t.Errorf("header too short: got %d bytes, want at least 20", len(pkt))
		}
	})

	t.Run("IPv6: different destination ports produce different output", func(t *testing.T) {
		pkt1, err1 := createTCPSynHeader(srcIPv6, dstIPv6, 12345, 80)
		pkt2, err2 := createTCPSynHeader(srcIPv6, dstIPv6, 12345, 443)
		if err1 != nil || err2 != nil {
			t.Fatalf("unexpected errors: %v, %v", err1, err2)
		}
		if bytes.Equal(pkt1, pkt2) {
			t.Error("headers with different dst ports should differ")
		}
	})

	t.Run("IPv4 and IPv6 headers for the same ports differ (different checksum)", func(t *testing.T) {
		pkt4, err4 := createTCPSynHeader(srcIPv4, dstIPv4, 12345, 80)
		pkt6, err6 := createTCPSynHeader(srcIPv6, dstIPv6, 12345, 80)
		if err4 != nil || err6 != nil {
			t.Fatalf("unexpected errors: %v, %v", err4, err6)
		}
		// Checksums are computed over different pseudo-headers so they must differ
		if bytes.Equal(pkt4, pkt6) {
			t.Error("IPv4 and IPv6 TCP headers should have different checksums")
		}
	})
}
