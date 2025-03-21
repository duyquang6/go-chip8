# Go CHIP-8 Emulator

A CHIP-8 emulator implemented in Go using [`tcell`](go.mod) for display rendering.

## Features

- Complete CHIP-8 instruction set implementation
- Real-time display rendering at 30Hz
- Keyboard input support with configurable key mappings
- Timer and sound support (500Hz CPU, 60Hz timers)
- Debug logging

## Prerequisites

- Go 1.24.1 or later
- A terminal that supports TCell

## Installation

```bash
git clone https://github.com/duyquang6/go-chip-ei.git
cd go-chip-ei
go build
```

## Usage

Run the emulator with:

```bash
./go-chip-ei
```

By default, it loads the [`roms/pong.ch8`](roms/pong.ch8) ROM.

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

The emulator uses structured logging (slog) for debugging. Logs are output in JSON format with source information included.

## License

This project is open source and available under the MIT License.