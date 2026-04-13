package main

import (
	"net"
	"testing"
)

// Tests for the IsIPv6 helper function.
func TestIsIPv6(t *testing.T) {
	tests := []struct {
		name string
		ip   net.IP
		want bool
	}{
		{"IPv4 address", net.ParseIP("192.168.1.1").To4(), false},
		{"IPv4 loopback", net.ParseIP("127.0.0.1").To4(), false},
		{"IPv6 loopback", net.ParseIP("::1"), true},
		{"IPv6 global", net.ParseIP("2001:db8::1"), true},
		{"IPv6 link-local", net.ParseIP("fe80::1"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsIPv6(tt.ip)
			if got != tt.want {
				t.Errorf("IsIPv6(%s) = %v, want %v", tt.ip, got, tt.want)
			}
		})
	}
}

// Tests for GetLocalIPv4 to ensure it returns a valid IPv4 from the loopback.
func TestGetLocalIPv4(t *testing.T) {
	t.Run("Loopback interface has an IPv4 address", func(t *testing.T) {
		ip, err := GetLocalIPv4("lo0")
		if err != nil {
			t.Skipf("lo0 not available on this system: %v", err)
		}
		if ip.To4() == nil {
			t.Errorf("GetLocalIPv4(lo0) returned non-IPv4: %s", ip)
		}
	})

	t.Run("Non-existent interface returns error", func(t *testing.T) {
		_, err := GetLocalIPv4("nonexistent99")
		if err == nil {
			t.Error("expected error for non-existent interface")
		}
	})
}

// Tests for GetLocalIPv6 to ensure it returns a valid IPv6 from the loopback.
func TestGetLocalIPv6(t *testing.T) {
	t.Run("Loopback interface has an IPv6 address", func(t *testing.T) {
		ip, err := GetLocalIPv6("lo0")
		if err != nil {
			t.Skipf("lo0 has no IPv6 on this system: %v", err)
		}
		if ip.To4() != nil {
			t.Errorf("GetLocalIPv6(lo0) returned IPv4: %s", ip)
		}
	})

	t.Run("Non-existent interface returns error", func(t *testing.T) {
		_, err := GetLocalIPv6("nonexistent99")
		if err == nil {
			t.Error("expected error for non-existent interface")
		}
	})
}
