package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

// SetupSignalHandler creates a background listener for SIGINT and SIGTERM.
// When caught, the program exits immediately with code 0.
func SetupSignalHandler() {
	// Create a channel where the OS will send the signal alerts
	sigChan := make(chan os.Signal, 1)

	// Tell the OS to send SIGINT (Ctrl+C) and SIGTERM to our channel
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Background goroutine waits for the signal and exits cleanly
	go func() {
		sig := <-sigChan
		fmt.Fprintf(os.Stderr, "\nCaught %v, exiting.\n", sig)
		os.Exit(0)
	}()
}
