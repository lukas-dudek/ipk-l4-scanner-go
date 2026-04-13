package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/pflag"
)

// Args structure to store the results of parsing.
type Args struct {
	Interface string
	TCPPorts  []int
	UDPPorts  []int
	Host      string
	Timeout   int  // Milliseconds, default is 1000
	Help      bool // For -h or --help flags
	IfaceOnly bool // For -i flag without arguments (to list interfaces)
}

// ParseArgs uses the pflag library to process command line arguments.
// It takes a slice of strings (os.Args[1:]) and returns the parsed Args structure.
func ParseArgs(argv []string) (Args, error) {
	var args Args
	var tcpStrings []string
	var udpStrings []string

	// If the user passes ONLY "-i", they want to list interfaces.
	// Since pflag expects an argument for string flags, we intercept this here.
	if len(argv) == 1 && argv[0] == "-i" {
		args.IfaceOnly = true
		args.Timeout = 1000 // default timeout consistency
		return args, nil
	}

	// Create a new FlagSet. We use ContinueOnError so it doesn't automatically os.Exit()
	fs := pflag.NewFlagSet("scanner", pflag.ContinueOnError)

	// We don't want pflag to print its own error messages, we handle that in main.go
	fs.Usage = func() {}

	// Define our flags
	fs.StringVarP(&args.Interface, "interface", "i", "", "Network interface to use")
	fs.StringSliceVarP(&tcpStrings, "tcp", "t", nil, "TCP ports to scan")
	fs.StringSliceVarP(&udpStrings, "udp", "u", nil, "UDP ports to scan")
	fs.IntVarP(&args.Timeout, "timeout", "w", 1000, "Timeout for each scan in ms")
	fs.BoolVarP(&args.Help, "help", "h", false, "Show help message")

	err := fs.Parse(argv)
	if err != nil {
		return args, fmt.Errorf("flag parsing error: %w", err)
	}

	//user ask for help
	if args.Help {
		return args, nil
	}

	// Process the HOST parameter
	positional := fs.Args()

	// If there are no positional arguments, it's an error because we need a HOST.
	if len(positional) == 0 {
		return args, fmt.Errorf("missing target HOST")
	}
	// If there are too many, the user wrote something wrong.
	if len(positional) > 1 {
		return args, fmt.Errorf("too many positional arguments (expected only 1 HOST)")
	}
	args.Host = positional[0]

	//Validate that we actually have an interface assigned
	if args.Interface == "" {
		return args, fmt.Errorf("scanning requires an interface (-i)")
	}

	// Parse the TCP and UDP ports (they can be ranges or comma separated)
	// pflag's StringSlice splits by comma, but need to handle ranges
	for _, portStr := range tcpStrings {
		ports, err := parsePortRange(portStr)
		if err != nil {
			return args, fmt.Errorf("invalid TCP port: %w", err)
		}
		args.TCPPorts = append(args.TCPPorts, ports...)
	}

	for _, portStr := range udpStrings {
		ports, err := parsePortRange(portStr)
		if err != nil {
			return args, fmt.Errorf("invalid UDP port: %w", err)
		}
		args.UDPPorts = append(args.UDPPorts, ports...)
	}

	// We must have at least one port to scan.
	if len(args.TCPPorts) == 0 && len(args.UDPPorts) == 0 {
		return args, fmt.Errorf("at least one TCP (-t) or UDP (-u) port must be specified")
	}

	return args, nil
}

// Helper function to parse port strings (e.g. "80" or "1-1024")
func parsePortRange(s string) ([]int, error) {
	var results []int
	s = strings.TrimSpace(s)

	if s == "" {
		return nil, nil
	}

	// Handle ranges like "1-1024"
	if strings.Contains(s, "-") {
		parts := strings.Split(s, "-")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid range format: %s", s)
		}
		start, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
		end, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
		
		if err1 != nil || err2 != nil || start < 1 || end > 65535 || start > end {
			return nil, fmt.Errorf("invalid port range: %s", s)
		}
		
		for i := start; i <= end; i++ {
			results = append(results, i)
		}
		return results, nil
	}

	// Single port like "22"
	p, err := strconv.Atoi(s)
	if err != nil || p < 1 || p > 65535 {
		return nil, fmt.Errorf("invalid port: %s", s)
	}
	results = append(results, p)
	return results, nil
}
