package main

import (
	"reflect"
	"testing"
)

// Tests for ParseArgs to ensure correct parsing of command line parameters.
func TestParseArgs(t *testing.T) {
	tests := []struct {
		name    string
		argv    []string
		want    Args
		wantErr bool
	}{
		{
			name: "Basic TCP scan",
			argv: []string{"-i", "eth0", "-t", "80", "localhost"},
			want: Args{
				Interface: "eth0",
				TCPPorts:  []int{80},
				Host:      "localhost",
				Timeout:   1000,
			},
			wantErr: false,
		},
		{
			name: "UDP list of ports",
			argv: []string{"-i", "wlan0", "-u", "53,67,68", "8.8.8.8"},
			want: Args{
				Interface: "wlan0",
				UDPPorts:  []int{53, 67, 68},
				Host:      "8.8.8.8",
				Timeout:   1000,
			},
			wantErr: false,
		},
		{
			name: "Port range and custom timeout",
			argv: []string{"-i", "lo", "-t", "1-3", "127.0.0.1", "-w", "500"},
			want: Args{
				Interface: "lo",
				TCPPorts:  []int{1, 2, 3},
				Host:      "127.0.0.1",
				Timeout:   500,
			},
			wantErr: false,
		},
		{
			name:    "Help flag",
			argv:    []string{"-h"},
			want:    Args{Help: true, Timeout: 1000},
			wantErr: false,
		},
		{
			name:    "Listing interfaces only",
			argv:    []string{"-i"},
			want:    Args{IfaceOnly: true, Timeout: 1000},
			wantErr: false,
		},
		{
			name:    "Missing interface",
			argv:    []string{"-t", "80", "localhost"},
			wantErr: true,
		},
		{
			name:    "Invalid port number",
			argv:    []string{"-i", "eth0", "-t", "70000", "localhost"},
			wantErr: true,
		},
		{
			name:    "Port zero is invalid",
			argv:    []string{"-i", "eth0", "-t", "0", "localhost"},
			wantErr: true,
		},
		{
			name:    "Missing host",
			argv:    []string{"-i", "eth0", "-t", "80"},
			wantErr: true,
		},
		{
			name:    "No ports specified",
			argv:    []string{"-i", "eth0", "localhost"},
			wantErr: true,
		},
		{
			name:    "Invalid port range (start > end)",
			argv:    []string{"-i", "eth0", "-t", "100-10", "localhost"},
			wantErr: true,
		},
		{
			name: "TCP and UDP together",
			argv: []string{"-i", "eth0", "-t", "80", "-u", "53", "localhost"},
			want: Args{
				Interface: "eth0",
				TCPPorts:  []int{80},
				UDPPorts:  []int{53},
				Host:      "localhost",
				Timeout:   1000,
			},
			wantErr: false,
		},
		// IPv6 address as HOST
		{
			name: "IPv6 address as host",
			argv: []string{"-i", "eth0", "-t", "80", "2001:db8::1"},
			want: Args{
				Interface: "eth0",
				TCPPorts:  []int{80},
				Host:      "2001:db8::1",
				Timeout:   1000,
			},
			wantErr: false,
		},
		{
			name: "IPv6 loopback as host",
			argv: []string{"-i", "lo0", "-u", "53", "::1"},
			want: Args{
				Interface: "lo0",
				UDPPorts:  []int{53},
				Host:      "::1",
				Timeout:   1000,
			},
			wantErr: false,
		},
		// Edge cases for port boundaries
		{
			name: "Maximum valid port 65535",
			argv: []string{"-i", "eth0", "-t", "65535", "localhost"},
			want: Args{
				Interface: "eth0",
				TCPPorts:  []int{65535},
				Host:      "localhost",
				Timeout:   1000,
			},
			wantErr: false,
		},
		{
			name: "Minimum valid port 1",
			argv: []string{"-i", "eth0", "-u", "1", "localhost"},
			want: Args{
				Interface: "eth0",
				UDPPorts:  []int{1},
				Host:      "localhost",
				Timeout:   1000,
			},
			wantErr: false,
		},
		{
			name:    "Port above 65535 is invalid",
			argv:    []string{"-i", "eth0", "-t", "65536", "localhost"},
			wantErr: true,
		},
		{
			name:    "Negative port is invalid",
			argv:    []string{"-i", "eth0", "-t", "-1", "localhost"},
			wantErr: true,
		},
		// Arguments in different order (zadání: "All arguments can be in any order")
		{
			name: "Host before flags",
			argv: []string{"localhost", "-i", "eth0", "-t", "22"},
			want: Args{
				Interface: "eth0",
				TCPPorts:  []int{22},
				Host:      "localhost",
				Timeout:   1000,
			},
			wantErr: false,
		},
		{
			name: "Timeout before other flags",
			argv: []string{"-w", "2000", "-i", "eth0", "-t", "443", "example.com"},
			want: Args{
				Interface: "eth0",
				TCPPorts:  []int{443},
				Host:      "example.com",
				Timeout:   2000,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseArgs(tt.argv)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseArgs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("ParseArgs() got = %+v, want %+v", got, tt.want)
				}
			}
		})
	}
}
