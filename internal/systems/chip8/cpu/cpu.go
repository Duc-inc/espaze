package cpu

import (
	"fmt"
	"math/rand"

	"github.com/Duc-inc/espaze/internal/systems/chip8/display"
	"github.com/Duc-inc/espaze/internal/systems/chip8/input"
	"github.com/Duc-inc/espaze/internal/systems/chip8/memory"
	"github.com/Duc-inc/espaze/internal/systems/chip8/timer"
)

const programStart = memory.ProgramStart

// CPU is the CHIP-8 instruction interpreter. It owns nothing but the
// registers and call stack; RAM, the screen, the keypad and the timers
// are all shared components wired in from the outside so the rest of
// the system can observe/drive them independently of the CPU.
type CPU struct {
	regs  registers
	stack stack

	mem    *memory.Memory
	disp   *display.Display
	keys   *input.Keypad
	delay  *timer.Timer
	sound  *timer.Timer
	rand   *rand.Rand
	waitVX *uint8 // set while blocked in Fx0A, nil otherwise
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

// Reset returns registers and the call stack to their post-boot state.
// It does not touch memory, display or timers - callers own that.
func (c *CPU) Reset() {
	c.regs.reset()
	c.stack.reset()
	c.waitVX = nil
}

// Step fetches, decodes and executes exactly one instruction.
func (c *CPU) Step() error {
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

// ProgramCounter exposes PC, mainly for save states.
func (c *CPU) ProgramCounter() uint16 { return c.regs.PC }

// Snapshot captures every piece of CPU state needed to resume later.
type Snapshot struct {
	V      [16]byte
	I      uint16
	PC     uint16
	Stack  [stackDepth]uint16
	SP     uint8
	Waited bool
	WaitVX uint8
}

// Snapshot returns the CPU's current state for serialization.
func (c *CPU) Snapshot() Snapshot {
	s := Snapshot{
		V:     c.regs.V,
		I:     c.regs.I,
		PC:    c.regs.PC,
		Stack: c.stack.data,
		SP:    c.stack.sp,
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
	if s.Waited {
		vx := s.WaitVX
		c.waitVX = &vx
	} else {
		c.waitVX = nil
	}
}
