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

// createUDPHeader builds ONLY the UDP header bytes (+ empty payload).
// Supports both IPv4 and IPv6 pseudo-headers for checksum.
func createUDPHeader(srcIP, dstIP net.IP, srcPort, dstPort uint16) ([]byte, error) {
	udp := &layers.UDP{
		SrcPort: layers.UDPPort(srcPort),
		DstPort: layers.UDPPort(dstPort),
	}

	// Choose the right IP layer for the checksum pseudo-header
	if IsIPv6(dstIP) {
		ipv6 := &layers.IPv6{
			SrcIP:      srcIP,
			DstIP:      dstIP,
			NextHeader: layers.IPProtocolUDP,
		}
		if err := udp.SetNetworkLayerForChecksum(ipv6); err != nil {
			return nil, fmt.Errorf("IPv6 checksum setup failed: %w", err)
		}
	} else {
		ipv4 := &layers.IPv4{
			SrcIP:    srcIP,
			DstIP:    dstIP,
			Protocol: layers.IPProtocolUDP,
		}
		if err := udp.SetNetworkLayerForChecksum(ipv4); err != nil {
			return nil, fmt.Errorf("IPv4 checksum setup failed: %w", err)
		}
	}

	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{
		ComputeChecksums: true,
		FixLengths:       true,
	}

	// Serialize UDP header + empty payload
	if err := gopacket.SerializeLayers(buf, opts, udp, gopacket.Payload([]byte{})); err != nil {
		return nil, fmt.Errorf("serialize failed: %w", err)
	}

	return buf.Bytes(), nil
}

// ScanUDPPort sends a UDP datagram and listens for ICMP Port Unreachable.
func ScanUDPPort(iface string, srcIP, dstIP net.IP, dstPort int, timeoutMs int, conn net.PacketConn) (string, error) {
	srcPort := uint16(rand.Intn(64511) + 1024)

	// Open pcap for listening only
	handle, err := pcap.OpenLive(iface, 65535, false, time.Duration(timeoutMs)*time.Millisecond)
	if err != nil {
		return "", fmt.Errorf("pcap open failed: %w", err)
	}
	defer handle.Close()

	// BPF filter: ICMPv4 or ICMPv6 from the target depending on IP version
	var filter string
	if IsIPv6(dstIP) {
		filter = fmt.Sprintf("icmp6 and src host %s", dstIP.String())
	} else {
		filter = fmt.Sprintf("icmp and src host %s", dstIP.String())
	}
	if err := handle.SetBPFFilter(filter); err != nil {
		return "", fmt.Errorf("filter failed: %w", err)
	}

	// Build the UDP header
	udpBytes, err := createUDPHeader(srcIP, dstIP, srcPort, uint16(dstPort))
	if err != nil {
		return "", err
	}

	// Send via the shared raw IP socket
	if _, err := conn.WriteTo(udpBytes, &net.IPAddr{IP: dstIP}); err != nil {
		return "", fmt.Errorf("send failed: %w", err)
	}

	// Read ICMP replies until deadline
	deadline := time.Now().Add(time.Duration(timeoutMs) * time.Millisecond)
	for time.Now().Before(deadline) {
		data, _, err := handle.ReadPacketData()
		if err != nil {
			continue
		}

		packet := gopacket.NewPacket(data, handle.LinkType(), gopacket.DecodeOptions{Lazy: true})

		if IsIPv6(dstIP) {
			// ICMPv6: Type 1 (Destination Unreachable) Code 4 (Port Unreachable)
			icmpLayer := packet.Layer(layers.LayerTypeICMPv6)
			if icmpLayer == nil {
				continue
			}
			icmp, _ := icmpLayer.(*layers.ICMPv6)
			if icmp.TypeCode == layers.CreateICMPv6TypeCode(1, 4) {
				return "closed", nil
			}
		} else {
			// ICMPv4: Type 3 (Destination Unreachable) Code 3 (Port Unreachable)
			icmpLayer := packet.Layer(layers.LayerTypeICMPv4)
			if icmpLayer == nil {
				continue
			}
			icmp, _ := icmpLayer.(*layers.ICMPv4)
			if icmp.TypeCode == layers.CreateICMPv4TypeCode(3, 3) {
				return "closed", nil
			}
		}
	}

	// No ICMP reply = open (per RFC 1122)
	return "open", nil
}
