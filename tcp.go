package main

import (
	"fmt"
	"math/rand"
	"net"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

// createTCPSynHeader builds ONLY the TCP header bytes.
// Supports both IPv4 and IPv6 by choosing the correct pseudo-header
// for the checksum calculation. The OS adds the actual IP header.
func createTCPSynHeader(srcIP, dstIP net.IP, srcPort, dstPort uint16) ([]byte, error) {
	tcp := &layers.TCP{
		SrcPort: layers.TCPPort(srcPort),
		DstPort: layers.TCPPort(dstPort),
		SYN:     true,
		Window:  1024,
		Seq:     rand.Uint32(),
	}

	// Choose the right IP layer for the checksum pseudo-header
	if IsIPv6(dstIP) {
		ipv6 := &layers.IPv6{
			SrcIP:      srcIP,
			DstIP:      dstIP,
			NextHeader: layers.IPProtocolTCP,
		}
		if err := tcp.SetNetworkLayerForChecksum(ipv6); err != nil {
			return nil, fmt.Errorf("IPv6 checksum setup failed: %w", err)
		}
	} else {
		ipv4 := &layers.IPv4{
			SrcIP:    srcIP.To4(),
			DstIP:    dstIP.To4(),
			Protocol: layers.IPProtocolTCP,
		}
		if err := tcp.SetNetworkLayerForChecksum(ipv4); err != nil {
			return nil, fmt.Errorf("IPv4 checksum setup failed: %w", err)
		}
	}

	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{
		ComputeChecksums: true,
		FixLengths:       true,
	}

	// Serialize ONLY the TCP layer (OS adds IP header when sending)
	if err := gopacket.SerializeLayers(buf, opts, tcp); err != nil {
		return nil, fmt.Errorf("failed to serialize TCP header: %w", err)
	}

	return buf.Bytes(), nil
}

// ScanTCPPort sends a TCP SYN and waits for a reply via pcap.
// The raw socket conn is shared across goroutines to avoid opening too many.
// Returns "open" (SYN-ACK), "closed" (RST), or "filtered" (no reply).
func ScanTCPPort(iface string, srcIP, dstIP net.IP, dstPort int, timeoutMs int, conn net.PacketConn) (string, error) {
	srcPort := uint16(rand.Intn(64511) + 1024)

	// Open pcap handle for listening only
	handle, err := pcap.OpenLive(iface, 65535, false, 10 * time.Millisecond)
	if err != nil {
		return "", fmt.Errorf("failed to open pcap: %w", err)
	}
	defer handle.Close()

	// BPF filter works the same for both IPv4 and IPv6
	filter := fmt.Sprintf("tcp and src host %s and src port %d", dstIP.String(), dstPort)
	if err := handle.SetBPFFilter(filter); err != nil {
		return "", fmt.Errorf("filter failed: %w", err)
	}

	// Build the SYN header
	tcpBytes, err := createTCPSynHeader(srcIP, dstIP, srcPort, uint16(dstPort))
	if err != nil {
		return "", err
	}

	// Send via the shared raw IP socket
	if _, err := conn.WriteTo(tcpBytes, &net.IPAddr{IP: dstIP}); err != nil {
		return "", fmt.Errorf("failed to send SYN: %w", err)
	}

	// Read replies until deadline
	deadline := time.Now().Add(time.Duration(timeoutMs) * time.Millisecond)
	for time.Now().Before(deadline) {
		data, _, err := handle.ReadPacketData()
		if err != nil {
			continue
		}

		packet := gopacket.NewPacket(data, handle.LinkType(), gopacket.DecodeOptions{Lazy: true})
		tcpLayer := packet.Layer(layers.LayerTypeTCP)
		if tcpLayer == nil {
			continue
		}
		tcp, _ := tcpLayer.(*layers.TCP)

		// Make sure the reply is for our SYN (matches our source port)
		if uint16(tcp.DstPort) != srcPort {
			continue
		}

		// SYN-ACK = port is open
		if tcp.SYN && tcp.ACK {
			return "open", nil
		}

		// RST = port is closed
		if tcp.RST {
			return "closed", nil
		}
	}

	// No reply within timeout = filtered
	return "filtered", nil
}
