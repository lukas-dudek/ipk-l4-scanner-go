package main

import (
	"bytes"
	"net"
	"testing"
)

// Tests for createUDPHeader with both IPv4 and IPv6 addresses.
func TestCreateUDPHeader(t *testing.T) {
	srcIPv4 := net.ParseIP("192.168.1.1").To4()
	dstIPv4 := net.ParseIP("8.8.8.8").To4()
	srcIPv6 := net.ParseIP("::1")
	dstIPv6 := net.ParseIP("2001:db8::1")

	// IPv4 tests
	t.Run("IPv4: returns non-empty bytes", func(t *testing.T) {
		pkt, err := createUDPHeader(srcIPv4, dstIPv4, 12345, 53)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(pkt) == 0 {
			t.Error("returned empty bytes")
		}
	})

	t.Run("IPv4: different destination ports produce different output", func(t *testing.T) {
		pkt1, err1 := createUDPHeader(srcIPv4, dstIPv4, 12345, 53)
		pkt2, err2 := createUDPHeader(srcIPv4, dstIPv4, 12345, 123)
		if err1 != nil || err2 != nil {
			t.Fatalf("unexpected errors: %v, %v", err1, err2)
		}
		if bytes.Equal(pkt1, pkt2) {
			t.Error("headers with different destination ports should differ")
		}
	})

	t.Run("IPv4: UDP header length is exactly 8 bytes", func(t *testing.T) {
		pkt, err := createUDPHeader(srcIPv4, dstIPv4, 12345, 53)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(pkt) != 8 {
			t.Errorf("header length wrong: got %d bytes, want 8", len(pkt))
		}
	})

	t.Run("IPv4: minimum port 1 is accepted", func(t *testing.T) {
		_, err := createUDPHeader(srcIPv4, dstIPv4, 1024, 1)
		if err != nil {
			t.Fatalf("port 1 should be valid, got error: %v", err)
		}
	})

	t.Run("IPv4: maximum port 65535 is accepted", func(t *testing.T) {
		_, err := createUDPHeader(srcIPv4, dstIPv4, 1024, 65535)
		if err != nil {
			t.Fatalf("port 65535 should be valid, got error: %v", err)
		}
	})

	// IPv6 tests
	t.Run("IPv6: returns non-empty bytes", func(t *testing.T) {
		pkt, err := createUDPHeader(srcIPv6, dstIPv6, 12345, 53)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(pkt) == 0 {
			t.Error("returned empty bytes")
		}
	})

	t.Run("IPv6: UDP header length is exactly 8 bytes", func(t *testing.T) {
		pkt, err := createUDPHeader(srcIPv6, dstIPv6, 12345, 53)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(pkt) != 8 {
			t.Errorf("header length wrong: got %d bytes, want 8", len(pkt))
		}
	})

	t.Run("IPv6: different source IPs produce different checksum", func(t *testing.T) {
		pkt1, err1 := createUDPHeader(srcIPv6, dstIPv6, 12345, 53)
		pkt2, err2 := createUDPHeader(net.ParseIP("fe80::1"), dstIPv6, 12345, 53)
		if err1 != nil || err2 != nil {
			t.Fatalf("unexpected errors: %v, %v", err1, err2)
		}
		if bytes.Equal(pkt1, pkt2) {
			t.Error("different source IPs should produce different checksums")
		}
	})

	t.Run("IPv4 and IPv6 headers differ (different checksum)", func(t *testing.T) {
		pkt4, err4 := createUDPHeader(srcIPv4, dstIPv4, 12345, 53)
		pkt6, err6 := createUDPHeader(srcIPv6, dstIPv6, 12345, 53)
		if err4 != nil || err6 != nil {
			t.Fatalf("unexpected errors: %v, %v", err4, err6)
		}
		if bytes.Equal(pkt4, pkt6) {
			t.Error("IPv4 and IPv6 UDP headers should have different checksums")
		}
	})
}
