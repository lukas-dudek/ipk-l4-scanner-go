package main

import (
	"net"
	"testing"
)

// Tests for the ResolveHost function.
func TestResolveHost(t *testing.T) {
	tests := []struct {
		name    string
		host    string
		wantErr bool
		check   func(ips []net.IP) bool
	}{
		{
			name:    "Valid IPv4 address",
			host:    "8.8.8.8",
			wantErr: false,
			check: func(ips []net.IP) bool {
				return len(ips) == 1 && ips[0].String() == "8.8.8.8"
			},
		},
		{
			name:    "Valid IPv6 address",
			host:    "::1",
			wantErr: false,
			check: func(ips []net.IP) bool {
				return len(ips) == 1
			},
		},
		{
			name:    "Valid full IPv6 address",
			host:    "2001:4860:4860::8888",
			wantErr: false,
			check: func(ips []net.IP) bool {
				return len(ips) == 1 && ips[0].To4() == nil
			},
		},
		{
			name:    "Localhost hostname resolves to at least one IP",
			host:    "localhost",
			wantErr: false,
			check: func(ips []net.IP) bool {
				return len(ips) >= 1
			},
		},
		{
			name:    "Non-existent domain",
			host:    "tato.domena.vubec.neexistuje.invalid",
			wantErr: true,
		},
		{
			name:    "Empty string is not a valid host",
			host:    "",
			wantErr: true,
		},
		{
			name:    "IPv4 loopback address",
			host:    "127.0.0.1",
			wantErr: false,
			check: func(ips []net.IP) bool {
				return len(ips) == 1 && ips[0].String() == "127.0.0.1"
			},
		},
		{
			name:    "IPv6 loopback address",
			host:    "::1",
			wantErr: false,
			check: func(ips []net.IP) bool {
				return len(ips) == 1 && ips[0].String() == "::1"
			},
		},
		{
			name:    "IPv4 broadcast address is parsed without DNS",
			host:    "255.255.255.255",
			wantErr: false,
			check: func(ips []net.IP) bool {
				return len(ips) == 1
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ips, err := ResolveHost(tt.host)
			if (err != nil) != tt.wantErr {
				t.Errorf("ResolveHost(%q) error = %v, wantErr %v", tt.host, err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.check != nil {
				if !tt.check(ips) {
					t.Errorf("ResolveHost(%q) returned unexpected IPs: %v", tt.host, ips)
				}
			}
		})
	}
}
