package main

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/gdamore/tcell/v2"
)

var (
	FONTSET = [...]byte{
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

	KEYMAP = map[rune]byte{
		'1': 0x1, '2': 0x2, '3': 0x3, '4': 0xC,
		'q': 0x4, 'w': 0x5, 'e': 0x6, 'r': 0xD,
		'a': 0x7, 's': 0x8, 'd': 0x9, 'f': 0xE,
		'z': 0xA, 'x': 0x0, 'c': 0xB, 'v': 0xF,
	}

	REVKEYMAP = map[byte]rune{}
)

func init() {
	for k, v := range KEYMAP {
		REVKEYMAP[v] = k
	}
}

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

func NewWithROMPath(romPath string) (*Chip8VM, error) {
	payload, err := os.ReadFile(romPath)
	if err != nil {
		return nil, fmt.Errorf("read ROM error: %w", err)
	}
	return New(payload)
}

func New(romPayload []byte) (*Chip8VM, error) {
	vm := &Chip8VM{
		pc: 0x200,
	}
	if len(romPayload) >= 3584 {
		return nil, errors.New("ROM payload too large")
	}
	copy(vm.mem[:80], FONTSET[:])
	copy(vm.mem[0x200:len(romPayload)+0x200], romPayload)

	return vm, nil
}

func (vm *Chip8VM) Serve() error {
	cpuHz := 500
	timerHz := 60

	cpuPeriod := time.Second / time.Duration(cpuHz)
	timerPeriod := time.Second / time.Duration(timerHz)
	stopSignal := make(chan struct{})

	go func() {
		lastCPU := time.Now()
		lastTimer := time.Now()

		for {
			select {
			case <-stopSignal:
				log.Println("received stop signal, exit handleCycle goroutine")
				return
			default:
				// do cpu cycle
				if time.Since(lastCPU) >= cpuPeriod {
					vm.handleCycle()
					lastCPU = lastCPU.Add(cpuPeriod)
				}

				if time.Since(lastTimer) >= timerPeriod {
					vm.handleTimer()
					lastTimer = lastTimer.Add(timerPeriod)
				}

				// Sleep to prevent busy-waiting; 500 µs balances responsiveness and CPU usage
				time.Sleep(500 * time.Microsecond)
			}
		}
	}()

	// Display
	screen, err := vm.initDisplay()
	if err != nil {
		return fmt.Errorf("init display got error: %w", err)
	}
	defer screen.Fini()

	keyLastPressedAt := make(map[byte]time.Time) // Track pressed keys
	for {
		ev := screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			if ev.Key() == tcell.KeyCtrlC || ev.Key() == tcell.KeyEsc {
				screen.PostEvent(tcell.NewEventInterrupt(nil))
				return nil
			}

			char := ev.Rune()
			if key, ok := KEYMAP[char]; ok {
				if ev.Modifiers() == tcell.ModNone && vm.keys[key] == 0 {
					vm.keys[key] = 1
					keyLastPressedAt[key] = time.Now()
					log.Printf("Key %c pressed (0x%X)", char, key)
				}
			}
		case *tcell.EventInterrupt:
			close(stopSignal)
			log.Println("send interrupt msg")
			return nil
		}

		// Auto release key after 100ms
		for key, pressed := range vm.keys {
			key := byte(key)
			if pressed == 1 && time.Since(keyLastPressedAt[key]) >= 100*time.Millisecond {
				vm.keys[key] = 0
				log.Printf("Key %c released (0x%X)", REVKEYMAP[key], key)
			}
		}
	}
}

func (*Chip8VM) initDisplay() (tcell.Screen, error) {
	screen, err := tcell.NewScreen()
	if err != nil {
		return nil, err
	}
	if err := screen.Init(); err != nil {
		return nil, err
	}
	return screen, nil
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

	x := byte((opcode & 0x0F00) >> 8)
	y := byte((opcode & 0x00F0) >> 4)
	n := byte(opcode & 0x000F)
	nn := byte(opcode & 0x00FF)
	nnn := uint16(opcode & 0x0FFF)

	log.Printf("Handle opcode %04X \n", opcode)

	switch {
	case opcode == 0x00E0:
		// clear screen, set black
		vm.display = [64][32]byte{}
	case opcode == 0x00EE:
		// RET return from subroutine
		if vm.sp == 0 {
			panic("stack underflow")
		}
		vm.sp--
		vm.pc = vm.stack[vm.sp]
	case opcode&0xF000 == 0x0000:
		// SYS addr
	case opcode&0xF000 == 0x1000:
		// JP addr
		vm.pc = nnn
	case opcode&0xF000 == 0x2000:
		// CALL addr
		if vm.sp == 16 {
			panic("stack overflow, limit depth 16")
		}
		vm.stack[vm.sp] = vm.pc
		vm.sp++
		vm.pc = nnn
	case opcode&0xF000 == 0x3000:
		// SE Vx, byte (skip if equal)
		if vm.V[x] == nn {
			vm.pc += 2
		}
	case opcode&0xF000 == 0x4000:
		// SNE Vx, byte
		if vm.V[x] != nn {
			vm.pc += 2
		}
	case opcode&0xF00F == 0x5000:
		// SE Vx, Vy
		if vm.V[x] == vm.V[y] {
			vm.pc += 2
		}
	case opcode&0xF000 == 0x6000:
		// LD Vx, byte
		vm.V[x] = nn
	case opcode&0xF000 == 0x7000:
		// ADD Vx, byte
		vm.V[x] = byte(uint16(vm.V[x]) + uint16(nn)&0xFF)
	// Arithm logic
	case opcode&0xF00F == 0x8000:
		// LD Vx, Vy
		vm.V[x] = vm.V[y]
	case opcode&0xF00F == 0x8001:
		// OR Vx, Vy
		vm.V[x] |= vm.V[y]
	case opcode&0xF00F == 0x8002:
		// AND Vx, Vy
		vm.V[x] &= vm.V[y]
	case opcode&0xF00F == 0x8003:
		vm.V[x] ^= vm.V[y]
	case opcode&0xF00F == 0x8004:
		sum := uint16(vm.V[x]) + uint16(vm.V[y])
		vm.V[0xF] = byte(sum >> 8)
		vm.V[x] = byte(sum & 0xFF)
	case opcode&0xF00F == 0x8005:
		diff := uint16(vm.V[x]) - uint16(vm.V[y])
		vm.V[0xF] = byte((diff>>15 ^ 1))
		vm.V[x] = byte(diff & 0xFF)
	case opcode&0xF00F == 0x8006:
		vm.V[0xF] = vm.V[x] & 1
		vm.V[x] >>= 1
	case opcode&0xF00F == 0x8007:
		diff := uint16(vm.V[y]) - uint16(vm.V[x])
		vm.V[0xF] = byte((diff>>15 ^ 1))
		vm.V[x] = byte(diff & 0xFF)
	case opcode&0xF00F == 0x800E:
		vm.V[0xF] = (vm.V[x] >> 7) & 1
		vm.V[x] = (vm.V[x] << 1) & 0xFF
	case opcode&0xF00F == 0x9000:
		// SNE Vx, Vy
		if vm.V[x] != vm.V[y] {
			vm.pc += 2
		}
	case opcode&0xF000 == 0xA000:
		// LD I, addr
		vm.I = nnn
	case opcode&0xF000 == 0xB000:
		// JP V0, addr + nnn
		vm.pc = uint16(vm.V[0]) + nnn
	case opcode&0xF000 == 0xC000:
		// RND Vx, byte
		rnd := byte(rand.Intn(256))
		vm.V[x] = rnd & nn
	case opcode&0xF000 == 0xD000:
		// DRW Vx, Vy, nibble
		vm.V[0xF] = 0
		for row := range n {
			sprite := vm.mem[vm.I+uint16(row)]
			for col := range byte(8) {
				if sprite&(0x80>>col) > 0 {
					r, c := (vm.V[x]+row)%64, (vm.V[y]+col)%32
					vm.display[r][c] ^= 1
					if vm.display[r][c] == 0 {
						vm.V[0xF] = 1
					}
				}
			}
		}
	case opcode&0xF0FF == 0xE09E:
		// SKP Vx (skip if key is pressed)
		vx := vm.V[x]
		if vm.keys[vx] == 1 {
			vm.pc += 2
		}
	case opcode&0xF0FF == 0xE0A1:
		// SKNP Vx (skip if key is pressed)
		if vm.keys[vm.V[x]] == 0 {
			vm.pc += 2
		}
	case opcode&0xF0FF == 0xF007:
		// LD Vx, Dt
		vm.V[x] = vm.delayTimer
	case opcode&0xF0FF == 0xF00A:
		// LD Vx, K wait till key pressed
		found := false
		for k, isPress := range vm.keys {
			if isPress == 1 {
				vm.V[x] = byte(k)
				found = true
				break
			}
		}
		// loop
		if !found {
			vm.pc -= 2
		}
	case opcode&0xF0FF == 0xF015:
		// LD DT, Vx
		vm.delayTimer = vm.V[x]
	case opcode&0xF0FF == 0xF018:
		// LD ST, Vx
		vm.soundTimer = vm.V[x]
	case opcode&0xF0FF == 0xF01E:
		// ADD I, Vx
		vm.I = (vm.I + uint16(vm.V[x])) & 0xFFF
	case opcode&0xF0FF == 0xF029:
		// LD F, Vx
		vm.I = uint16(vm.V[x] * 5)
	case opcode&0xF0FF == 0xF033:
		// LD B, Vx
		num := vm.V[x]
		vm.mem[vm.I] = num / 100
		vm.mem[vm.I+1] = (num % 100) / 10
		vm.mem[vm.I+2] = (num % 10)
	case opcode&0xF0FF == 0xF055:
		// LD [I], Vx
		for i := range x + 1 {
			vm.mem[vm.I+uint16(i)] = vm.V[i]
		}
	case opcode&0xF0FF == 0xF065:
		// LD Vx, [I]
		for i := range x + 1 {
			vm.V[i] = vm.mem[vm.I+uint16(i)]
		}
	default:
		log.Printf("opcode %X not implement yet", opcode)
	}
}
