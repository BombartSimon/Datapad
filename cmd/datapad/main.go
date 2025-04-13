package main

import (
	"datapad/internal/tui"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	// Define command line options
	var storagePath string
	flag.StringVar(&storagePath, "storage", "", "Path to notes storage folder (optional)")
	flag.Parse()

	// If no path is provided, use a default folder in the home directory
	if storagePath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: unable to determine home directory: %v\n", err)
			os.Exit(1)
		}
		storagePath = filepath.Join(homeDir, ".datapad")
	}

	// Launch the TUI application
	if err := tui.App(storagePath); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
