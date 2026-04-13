package main

import (
	"fmt"
	"os"
)

func main() {
	// setup Ctrl+C handler
	SetupSignalHandler()

	// os.Args[0] is the program name (e.g. ./ipk-L4-scan), so we skip it
	// and send only the actual parameters to our parser.
	args, err := ParseArgs(os.Args[1:])
	
	// If the parser returned an error (err != nil), print it and exit with code 1.
	if err != nil {
		fmt.Fprintf(os.Stderr, "Parsing error: %v\n", err)
		os.Exit(1)
	}

	//user ask for help (-h or --help)
	if args.Help {
		PrintHelp()
		os.Exit(0)
	}

	//user ask to only list network interfaces (-i without other params)
	if args.IfaceOnly {
		err := ListInterfaces()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error gathering network interfaces: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	// resolve the host to actual IPs
	ips, err := ResolveHost(args.Host)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	//launch the scanner
	Scan(args, ips)
}

// PrintHelp prints out instructions on how to use the scanner.
func PrintHelp() {
	fmt.Println("Usage:")
	fmt.Println("  ./ipk-L4-scan -i INTERFACE [-u PORTS] [-t PORTS] HOST [-w TIMEOUT]")
	fmt.Println("  ./ipk-L4-scan -i")
	fmt.Println("  ./ipk-L4-scan -h | --help")
	fmt.Println("\nOptions:")
	fmt.Println("  -h, --help    Show this help message and exit.")
	fmt.Println("  -i INTERFACE  Specify the network interface to use.")
	fmt.Println("  -i (alone)    List all active network interfaces and exit.")
	fmt.Println("  -t PORTS      TCP ports to scan (e.g. 22, 80-88, 80,443).")
	fmt.Println("  -u PORTS      UDP ports to scan.")
	fmt.Println("  -w TIMEOUT    Timeout in milliseconds for each port (default 1000).")
	fmt.Println("  HOST          Target hostname or IP address.")
}
