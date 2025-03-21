# Go CHIP-8 Emulator

A CHIP-8 emulator implemented in Go using [`tcell`](go.mod) for display rendering.

## Features

- Complete CHIP-8 instruction set implementation
- Real-time display rendering at 30Hz
- Keyboard input support with configurable key mappings
- Timer and sound support (500Hz CPU, 60Hz timers)
- Debug logging with structured output
- Command-line interface with customizable options

## Prerequisites

- Go 1.24.1 or later
- A terminal that supports TCell

## Installation

```bash
git clone https://github.com/duyquang6/go-chip-8.git
cd go-chip-8
go build -o chip8
```

## Usage

Run the emulator with various options:

```bash
# Run with default ROM (pong)
./chip8

# Run with a specific ROM
./chip8 --rom path/to/rom.ch8

# Run with debug logging enabled
./chip8 --rom path/to/rom.ch8 --debug

# Show available options
./chip8 --help
```

By default, it loads the [`roms/pong.ch8`](roms/pong.ch8) ROM.

## Command-line Options

- `--rom`: Path to CHIP-8 ROM file (default: roms/pong.ch8)
- `--debug`: Enable debug logging
- `--help`: Display help information

## Keyboard Mapping

```
Original CHIP-8    Keyboard Map
1 2 3 C           1 2 3 4
4 5 6 D    -->    Q W E R
7 8 9 E           A S D F
A 0 B F           Z X C V
```

## Architecture

- [`main.go`](main.go) - Entry point and ROM loading
- [`vm.go`](vm.go) - CHIP-8 CPU implementation including:
  - Memory management (4KB)
  - Display buffer (64x32 pixels)
  - CPU registers and stack
  - Instruction execution
  - Timer handling
  - Input processing

## Technical Details

- CPU frequency: 500Hz
- Timer frequency: 60Hz
- Display refresh: 30Hz
- Memory: 4KB RAM
- Display: 64x32 monochrome pixels
- 16 general purpose registers (V0-VF)
- 16-level stack depth

## Controls

- ESC or Ctrl+C to exit
- Keys are automatically released after 100ms

## Development

The emulator uses structured logging (slog) for debugging. When debug mode is enabled (`--debug`), detailed logs are output in text format with source information included.

## License

This project is open source and available under the MIT License.