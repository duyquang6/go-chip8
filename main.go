package main

import (
	"log"
	"log/slog"
	"os"
)

func main() {
	vm, err := NewWithROMPath("/var/home/ligt/me/go-chip-ei/roms/pong.ch8")
	// set log level debug
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})))

	if err != nil {
		log.Fatal("init rom failed:", err)
	}

	if err := vm.Serve(); err != nil {
		log.Fatal("serve failed:", err)
	}
}
