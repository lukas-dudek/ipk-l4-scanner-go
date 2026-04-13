package main

import (
	"fmt"
	"net"
	"os"
	"sync"
	"time"
)

// Scan is the main scanning loop.
// Each port is scanned in its own goroutine for speed.
// A semaphore limits how many goroutines run at once to avoid flooding
// the network card and getting wrong results.
func Scan(args Args, ips []net.IP) {
	// Open shared raw sockets (one pair per IP version)
	// IPv4 sockets
	tcpConn4, err := net.ListenPacket("ip4:tcp", "0.0.0.0")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create IPv4 TCP raw socket: %v\n", err)
		os.Exit(1)
	}
	defer tcpConn4.Close()

	udpConn4, err := net.ListenPacket("ip4:udp", "0.0.0.0")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create IPv4 UDP raw socket: %v\n", err)
		os.Exit(1)
	}
	defer udpConn4.Close()

	// IPv6 sockets
	tcpConn6, err := net.ListenPacket("ip6:tcp", "::")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create IPv6 TCP raw socket: %v\n", err)
		os.Exit(1)
	}
	defer tcpConn6.Close()

	udpConn6, err := net.ListenPacket("ip6:udp", "::")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create IPv6 UDP raw socket: %v\n", err)
		os.Exit(1)
	}
	defer udpConn6.Close()

	// WaitGroup keeps track of how many goroutines are still running
	var wg sync.WaitGroup

	// Mutex protects fmt.Printf so output lines don't overlap
	var mu sync.Mutex

	// Semaphore: limits concurrent goroutines to avoid overwhelming the NIC
	const maxConcurrent = 50
	sem := make(chan struct{}, maxConcurrent)

	// Mandatory pause between launching scans to avoid packet drops
	const pauseMs = 20

	// Loop over all resolved IP addresses
	for _, ip := range ips {
		ipCopy := ip // capture for goroutine

		// Pick the right source IP and raw sockets for this target
		var srcIP net.IP
		var tcpConn, udpConn net.PacketConn

		if IsIPv6(ipCopy) {
			srcIP, err = GetLocalIPv6(args.Interface)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Skipping %s: %v\n", ipCopy, err)
				continue
			}
			tcpConn = tcpConn6
			udpConn = udpConn6
		} else {
			srcIP, err = GetLocalIPv4(args.Interface)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Skipping %s: %v\n", ipCopy, err)
				continue
			}
			tcpConn = tcpConn4
			udpConn = udpConn4
		}

		// TCP ports
		for _, port := range args.TCPPorts {
			portCopy := port

			wg.Add(1)
			sem <- struct{}{} // acquire a slot
			go func() {
				defer wg.Done()
				defer func() { <-sem }() // release slot

				state, err := ScanTCPPort(args.Interface, srcIP, ipCopy, portCopy, args.Timeout, tcpConn)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error scanning TCP %d on %s: %v\n", portCopy, ipCopy, err)
					return
				}
				mu.Lock()
				fmt.Printf("%s %d tcp %s\n", ipCopy.String(), portCopy, state)
				mu.Unlock()
			}()

			time.Sleep(time.Duration(pauseMs) * time.Millisecond)
		}

		// UDP ports
		for _, port := range args.UDPPorts {
			portCopy := port

			wg.Add(1)
			sem <- struct{}{}
			go func() {
				defer wg.Done()
				defer func() { <-sem }()

				state, err := ScanUDPPort(args.Interface, srcIP, ipCopy, portCopy, args.Timeout, udpConn)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error scanning UDP %d on %s: %v\n", portCopy, ipCopy, err)
					return
				}
				mu.Lock()
				fmt.Printf("%s %d udp %s\n", ipCopy.String(), portCopy, state)
				mu.Unlock()
			}()

			time.Sleep(time.Duration(pauseMs) * time.Millisecond)
		}
	}

	// Wait for all goroutines to finish
	wg.Wait()
}
