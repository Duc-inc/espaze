package cpu

import (
	"fmt"
	"math/rand"

	"github.com/Duc-inc/espaze/internal/systems/chip8/input"
	"github.com/Duc-inc/espaze/internal/systems/chip8/memory"
	"github.com/Duc-inc/espaze/internal/systems/chip8/timer"
	"github.com/Duc-inc/espaze/internal/systems/schip/display"
)

// CPU is the Super-CHIP instruction interpreter. It reuses CHIP-8's
// memory, timer and keypad components unchanged (they don't depend on
// screen resolution) and only owns what's genuinely different: registers,
// the call stack, the RPL user-flag file, and the extended opcode table.
type CPU struct {
	regs  registers
	stack stack
	rpl   [8]byte

	mem    *memory.Memory
	disp   *display.Display
	keys   *input.Keypad
	delay  *timer.Timer
	sound  *timer.Timer
	rand   *rand.Rand
	waitVX *uint8
	halted bool
}

// New wires a CPU to the shared components it will read from and mutate.
func New(mem *memory.Memory, disp *display.Display, keys *input.Keypad, delay, sound *timer.Timer) *CPU {
	return &CPU{
		regs:  newRegisters(),
		mem:   mem,
		disp:  disp,
		keys:  keys,
		delay: delay,
		sound: sound,
		rand:  rand.New(rand.NewSource(1)),
	}
}

// Reset returns registers, the call stack and RPL flags to their
// post-boot state. It does not touch memory, display or timers.
func (c *CPU) Reset() {
	c.regs.reset()
	c.stack.reset()
	c.rpl = [8]byte{}
	c.waitVX = nil
	c.halted = false
}

// Step fetches, decodes and executes exactly one instruction. A no-op
// once the program has executed 00FD (exit).
func (c *CPU) Step() error {
	if c.halted {
		return nil
	}

	if c.waitVX != nil {
		key, ok := c.keys.AnyDown()
		if !ok {
			return nil
		}
		c.regs.V[*c.waitVX] = key
		c.waitVX = nil
	}

	if int(c.regs.PC) >= memory.Size-1 {
		return fmt.Errorf("cpu: program counter out of range: 0x%04X", c.regs.PC)
	}

	opcode := uint16(c.mem.Read(c.regs.PC))<<8 | uint16(c.mem.Read(c.regs.PC+1))
	c.regs.PC += 2

	return c.execute(opcode)
}

// Snapshot captures every piece of CPU state needed to resume later.
type Snapshot struct {
	V      [16]byte
	I      uint16
	PC     uint16
	Stack  [stackDepth]uint16
	SP     uint8
	RPL    [8]byte
	Waited bool
	WaitVX uint8
	Halted bool
}

// Snapshot returns the CPU's current state for serialization.
func (c *CPU) Snapshot() Snapshot {
	s := Snapshot{
		V:      c.regs.V,
		I:      c.regs.I,
		PC:     c.regs.PC,
		Stack:  c.stack.data,
		SP:     c.stack.sp,
		RPL:    c.rpl,
		Halted: c.halted,
	}
	if c.waitVX != nil {
		s.Waited = true
		s.WaitVX = *c.waitVX
	}
	return s
}

// Restore overwrites CPU state with a previously captured Snapshot.
func (c *CPU) Restore(s Snapshot) {
	c.regs.V = s.V
	c.regs.I = s.I
	c.regs.PC = s.PC
	c.stack.data = s.Stack
	c.stack.sp = s.SP
	c.rpl = s.RPL
	c.halted = s.Halted
	if s.Waited {
		vx := s.WaitVX
		c.waitVX = &vx
	} else {
		c.waitVX = nil
	}
}
