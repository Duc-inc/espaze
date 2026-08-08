// Package hle provides high-level emulation (HLE) of the small set of
// Nintendo SDK/OS functions real GameCube code calls into constantly
// (reporting debug text, reading more data off the disc) - functions
// this project can't actually run, since doing so would mean
// interpreting Nintendo's real, copyrighted library code this project
// doesn't have. Instead, an HLE function does the real function's
// observable job directly in Go and returns as if it had run.
//
// The mechanism: Install writes a real `sc` (system call) instruction
// at the target address and routes powerpc.CPU's existing
// SyscallHandler hook back to this package's dispatch table, keyed by
// the faulting address (recovered from PC, which has already advanced
// past the trap by the time the handler runs). Replacing a function's
// entry point with a trap instruction that calls back into emulator
// code is a standard technique across emulators generally, not
// something specific to - or verified against - real GameCube
// software; this project makes no claim about which real SDK
// functions live at which addresses in any specific game, since that
// varies per game build. Whoever wires this up (e.g. the ipl package's
// Apploader boot path) is expected to know the addresses its own
// target expects to call.
package hle

import "github.com/Duc-inc/espaze/internal/systems/powerpc"

// MemoryAccess is the subset of main memory access an HLE function
// typically needs: reading a C string or buffer argument, or copying
// data (e.g. from a disc image) into place.
type MemoryAccess interface {
	Read8(addr uint32) byte
	Write8(addr uint32, v byte)
}

// Func is one high-level-emulated function: given the CPU (so it can
// read arguments from GPR3-GPR10, the real PowerPC/PowerOpen calling
// convention, and set a return value in GPR3) and memory access for
// pointer/buffer arguments, it does the real function's job and
// returns - Install handles actually returning to the caller
// afterward, so Func itself doesn't need to touch PC/LR.
type Func func(cpu *powerpc.CPU, mem MemoryAccess)

// scInstruction is the `sc` opcode (primary 17) Install plants at
// every HLE'd address.
const scInstruction = 17 << 26

// Table dispatches HLE calls by address.
type Table struct {
	funcs map[uint32]Func
	mem   MemoryAccess
}

// New returns an empty Table backed by mem for any function that
// needs to read/write memory.
func New(mem MemoryAccess) *Table {
	return &Table{funcs: make(map[uint32]Func), mem: mem}
}

// Install registers fn to run whenever execution reaches addr: it
// writes a trap instruction at addr and (re)wires cpu's
// SyscallHandler to this table's dispatch, so calling Install several
// times for the same cpu is safe and only needs to happen once per
// cpu in practice.
func (t *Table) Install(cpu *powerpc.CPU, addr uint32, fn Func) {
	t.funcs[addr] = fn
	cpu.SyscallHandler = func(c *powerpc.CPU) {
		callAddr := c.PC() - 4 // PC already advanced past the trap
		if hf, ok := t.funcs[callAddr]; ok {
			hf(c, t.mem)
		}
		c.SetPC(c.LR())
	}
	writeInstr(t.mem, addr, scInstruction)
}

func writeInstr(mem MemoryAccess, addr, instr uint32) {
	mem.Write8(addr, byte(instr>>24))
	mem.Write8(addr+1, byte(instr>>16))
	mem.Write8(addr+2, byte(instr>>8))
	mem.Write8(addr+3, byte(instr))
}

func readCString(mem MemoryAccess, addr uint32) string {
	var b []byte
	for i := 0; i < 4096; i++ { // sane cap against a malformed/unterminated pointer
		c := mem.Read8(addr + uint32(i))
		if c == 0 {
			break
		}
		b = append(b, c)
	}
	return string(b)
}
