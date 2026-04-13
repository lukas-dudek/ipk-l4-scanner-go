package main

import (
	"fmt"
	"net"
)

// If it's already an IP, it will just return a slice with that one IP.
// If it's a domain, it will ask the DNS server to give us all IPs.
func ResolveHost(host string) ([]net.IP, error) {
	// check if the user just straight up gave us an IP address
	// IfParseIP doesn't return nil, it's a valid IP.
	if parsedIP := net.ParseIP(host); parsedIP != nil {
		//no DNS needed. We just put it into a slice and return.
		return []net.IP{parsedIP}, nil
	}

	// LookupIP will give us all A and AAAA records (IPv4 and IPv6).
	ips, err := net.LookupIP(host)
	if err != nil {
		return nil, fmt.Errorf("DNS lookup failed for %s: %v", host, err)
	}

	// Double check we actually got something back
	if len(ips) == 0 {
		return nil, fmt.Errorf("DNS lookup didn't find any IPs for %s", host)
	}

	// Return the list of resolved IP addresses.
	return ips, nil
}
