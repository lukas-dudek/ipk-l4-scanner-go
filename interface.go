package main

import (
	"fmt"
	"net"
)

// ListInterfaces finds all network interfaces on this computer
func ListInterfaces() error {
	ifaces, err := net.Interfaces()
	if err != nil {
		return err
	}

	// Loop through all the interfaces found
	for _, iface := range ifaces {
		// Only print interfaces that are actively running
		if iface.Flags&net.FlagUp != 0 {
			fmt.Println(iface.Name)
		}
	}

	return nil
}

// GetLocalIPv4 returns the first IPv4 address assigned to the given interface.
// Used as the source IP in outgoing IPv4 packets.
func GetLocalIPv4(ifaceName string) (net.IP, error) {
	addrs, err := getIfaceAddrs(ifaceName)
	if err != nil {
		return nil, err
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok {
			if ipnet.IP.To4() != nil {
				return ipnet.IP, nil
			}
		}
	}

	return nil, fmt.Errorf("no IPv4 address found on interface %s", ifaceName)
}

// GetLocalIPv6 returns the first global (non-link-local) IPv6 address
// assigned to the given interface. Falls back to link-local if no global found.
func GetLocalIPv6(ifaceName string) (net.IP, error) {
	addrs, err := getIfaceAddrs(ifaceName)
	if err != nil {
		return nil, err
	}

	var linkLocal net.IP

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok {
			ip := ipnet.IP
			// Skip IPv4 addresses
			if ip.To4() != nil {
				continue
			}
			// Prefer global unicast over link-local
			if ip.IsGlobalUnicast() {
				return ip, nil
			}
			// Remember link-local as fallback
			if ip.IsLinkLocalUnicast() && linkLocal == nil {
				linkLocal = ip
			}
		}
	}

	if linkLocal != nil {
		return linkLocal, nil
	}

	return nil, fmt.Errorf("no IPv6 address found on interface %s", ifaceName)
}

// IsIPv6 returns true if the given IP address is an IPv6 address.
func IsIPv6(ip net.IP) bool {
	return ip.To4() == nil
}

// getIfaceAddrs is a helper that returns all addresses of a named interface.
func getIfaceAddrs(ifaceName string) ([]net.Addr, error) {
	iface, err := net.InterfaceByName(ifaceName)
	if err != nil {
		return nil, fmt.Errorf("could not find interface %s: %w", ifaceName, err)
	}

	addrs, err := iface.Addrs()
	if err != nil {
		return nil, fmt.Errorf("could not get addresses for %s: %w", ifaceName, err)
	}

	return addrs, nil
}
