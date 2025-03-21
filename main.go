package main

import (
	"flag"
	"log"
	"log/slog"
	"os"
	"path/filepath"
)

func main() {
	// Define command line flags
	romPath := flag.String("rom", "", "Path to CHIP-8 ROM file")
	debug := flag.Bool("debug", false, "Enable debug logging")
	flag.Parse()

	// Set up logging
	logLevel := slog.LevelInfo
	if *debug {
		logLevel = slog.LevelDebug
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})))

	// If no ROM specified, use default
	if *romPath == "" {
		defaultROM := filepath.Join("roms", "pong.ch8")
		romPath = &defaultROM
		slog.Info("No ROM specified, using default", "path", *romPath)
	}

	// Initialize VM with ROM
	vm, err := NewWithROMPath(*romPath)
	if err != nil {
		log.Fatal("Failed to initialize ROM:", err)
	}

	// Start emulation
	if err := vm.Serve(); err != nil {
		log.Fatal("Emulation failed:", err)
	}
}
