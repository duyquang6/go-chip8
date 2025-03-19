package main

import (
	"context"
	"log"
	"math/rand"
	"os"
	"time"
)

type Chip8VM struct {
	mem [4096]byte
	V   [16]byte
	I   uint16
	pc  uint16

	// support depth 16 level callstack
	stack [16]uint16
	sp    byte

	delayTimer byte
	// trigger beep
	soundTimer byte

	// 64 x 32 pixels, monochrome pixel (1 pixel = 1 bit)
	// for simplify just use byte here
	display [64][32]byte

	keys [16]byte
}

var fontSet = [...]byte{
	0xF0, 0x90, 0x90, 0x90, 0xF0, // 0: 1111, 1001, 1001, 1001, 1111, skip 4 low bit, only used 4 high bit
	0x20, 0x60, 0x20, 0x20, 0x70, // 1: 0010, 0110, 0010, 0010, 0111
	0xF0, 0x10, 0xF0, 0x80, 0xF0, // 2: 1111, 0001, 1111, 1000, 1111
	0xF0, 0x10, 0xF0, 0x10, 0xF0, // 3
	0x90, 0x90, 0xF0, 0x10, 0x10, // 4
	0xF0, 0x80, 0xF0, 0x10, 0xF0, // 5
	0xF0, 0x80, 0xF0, 0x90, 0xF0, // 6
	0xF0, 0x10, 0x20, 0x40, 0x40, // 7
	0xF0, 0x90, 0xF0, 0x90, 0xF0, // 8
	0xF0, 0x90, 0xF0, 0x10, 0xF0, // 9
	0xF0, 0x90, 0xF0, 0x90, 0x90, // A
	0xE0, 0x90, 0xE0, 0x90, 0xE0, // B
	0xF0, 0x80, 0x80, 0x80, 0xF0, // C
	0xE0, 0x90, 0x90, 0x90, 0xE0, // D
	0xF0, 0x80, 0xF0, 0x80, 0xF0, // E
	0xF0, 0x80, 0xF0, 0x80, 0x80, // F
}

func NewWithROMPath(romPath string) *Chip8VM {
	payload, err := os.ReadFile(romPath)
	if err != nil {
		log.Println("load rom failed, err: ", err)
		return nil
	}
	return New(payload)
}

func New(romPayload []byte) *Chip8VM {
	vm := &Chip8VM{}
	copy(vm.mem[:80], fontSet[:])
	copy(vm.mem[0x200:len(romPayload)+0x200], romPayload)

	return vm
}

func (vm *Chip8VM) Serve(ctx context.Context) {
	cpuHz := 500
	timerHz := 60

	cpuCycleTime := 1 / float64(cpuHz)
	timerCycleTime := 1 / float64(timerHz)

	cpuTick := time.NewTicker(time.Duration(cpuCycleTime) * time.Second)
	timerTick := time.NewTicker(time.Duration(timerCycleTime) * time.Second)
	for {
		select {
		case <-cpuTick.C:
			vm.handleCycle()
		case <-timerTick.C:
			vm.handleTimer()
		case <-ctx.Done():
			log.Println("stopping....")
			return
		}
	}
}

func (vm *Chip8VM) handleTimer() {
	if vm.delayTimer > 0 {
		vm.delayTimer--
	}
	if vm.soundTimer > 0 {
		vm.soundTimer--
		// one beep
	}
}

func (vm *Chip8VM) handleCycle() {
	// fetch opcode, opcode is 2 byte
	opcode := uint16(vm.mem[vm.pc])<<8 | uint16(vm.mem[vm.pc+1])
	vm.pc += 2

	xx := opcode & 0xFF00 >> 8
	highX := opcode & 0xF000 >> 12
	lowX := xx & 0xF
	n := byte(opcode & 0xF)
	nn := byte(opcode & 0xFF)
	nnn := opcode & 0x0FFF

	log.Printf("Handle opcode %d \n", opcode)

	switch {
	case opcode == 0x00E0:
		// clear screen, set black
		vm.display = [64][32]byte{}
	case opcode == 0x00EE:
		// RET return from subroutine
		vm.sp--
		vm.pc = vm.stack[vm.sp]
	case highX == 1:
		// JP addr
		vm.pc = nnn
	case highX == 2:
		// CALL addr
		if vm.sp == 16 {
			panic("stack overflow, limit depth 16")
		}
		vm.stack[vm.sp] = vm.pc
		vm.sp++
		vm.pc = nnn
	case highX == 6:
		// LD Vx, byte
		vm.V[lowX] = nn
	case highX == 7:
		// ADD Vx, byte
		vm.V[lowX] += nn
	case highX == 0xA:
		//LD I, addr
		vm.I = nnn
	case highX == 0xC:
		// RND Vx, byte
		rnd := byte(rand.Intn(256))
		vm.V[lowX] = rnd & nn
	case highX == 0xD:
		// DRW Vx, Vy, nibble
		vx, vy := vm.V[lowX], vm.V[nn&0xF0>>4]
		height := n
		vm.V[0xF] = 0
		for row := range height {
			sprite := vm.mem[vm.I+uint16(row)]
			for col := range byte(8) {
				bit := sprite & (0x80 >> col)
				if bit > 0 {
					if vm.display[vx+row][vy+col] == 1 {
						vm.V[0xF] = 1
					}
					vm.display[vx+row][vy+col] ^= 1
				}
			}
		}
	default:
		log.Printf("opcode %X not implement yet", opcode)
	}
}
