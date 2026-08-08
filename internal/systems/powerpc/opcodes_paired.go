// opcodes_paired.go implements the Gekko/Broadway-specific "paired-
// single" extension: real GameCube/Wii CPUs add a second value
// alongside every FPR (registers.go's PS1), and most scalar double-
// precision FP instructions have a paired-single counterpart that
// operates on both slots at once - real GameCube game code leans on
// this heavily for vector/matrix math (including much of what
// internal/systems/gamecube/xf would eventually need to actually run
// as real PowerPC code, rather than this project's own Go
// implementation of the same math).
//
// This project's confidence here is lower than for the rest of the
// PowerPC package: the paired arithmetic instructions (ps_add and
// friends) are implemented with reasonable confidence, since Gekko
// reuses the exact same extended-opcode values as their scalar
// double-precision equivalents under primary opcode 63, just gated by
// primary opcode 4 instead - a well-documented, deliberate design
// choice. The paired load/store instructions (psq_l/psq_st) are only
// implemented in simplified form: real hardware's encoding also packs
// a quantize-register index and format selector into the instruction
// word, letting a load/store convert to/from 8/16-bit integer formats
// with a scale factor; this project always loads/stores both slots as
// plain 32-bit floats through an ordinary D-form effective address,
// since this project isn't confident enough in the real W/I sub-field
// bit layout to implement it without risking asserting precision it
// doesn't have.
package powerpc

import "math"

func init() {
	setExt4A(21, func(c *CPU, instr uint32) int { // ps_add
		frD, frA, frB := fieldRD(instr), fieldRA(instr), fieldRB(instr)
		c.regs.FPR[frD] = c.regs.FPR[frA] + c.regs.FPR[frB]
		c.regs.PS1[frD] = c.regs.PS1[frA] + c.regs.PS1[frB]
		return 4
	})
	setExt4A(20, func(c *CPU, instr uint32) int { // ps_sub
		frD, frA, frB := fieldRD(instr), fieldRA(instr), fieldRB(instr)
		c.regs.FPR[frD] = c.regs.FPR[frA] - c.regs.FPR[frB]
		c.regs.PS1[frD] = c.regs.PS1[frA] - c.regs.PS1[frB]
		return 4
	})
	setExt4A(25, func(c *CPU, instr uint32) int { // ps_mul (uses frC, not frB, like fmul)
		frD, frA, frC := fieldRD(instr), fieldRA(instr), fieldFRC(instr)
		c.regs.FPR[frD] = c.regs.FPR[frA] * c.regs.FPR[frC]
		c.regs.PS1[frD] = c.regs.PS1[frA] * c.regs.PS1[frC]
		return 4
	})
	setExt4A(18, func(c *CPU, instr uint32) int { // ps_div
		frD, frA, frB := fieldRD(instr), fieldRA(instr), fieldRB(instr)
		c.regs.FPR[frD] = c.regs.FPR[frA] / c.regs.FPR[frB]
		c.regs.PS1[frD] = c.regs.PS1[frA] / c.regs.PS1[frB]
		return 20
	})

	setExt4(72, func(c *CPU, instr uint32) int { // ps_mr
		frD, frB := fieldRD(instr), fieldRB(instr)
		c.regs.FPR[frD] = c.regs.FPR[frB]
		c.regs.PS1[frD] = c.regs.PS1[frB]
		return 4
	})
	setExt4(40, func(c *CPU, instr uint32) int { // ps_neg
		frD, frB := fieldRD(instr), fieldRB(instr)
		c.regs.FPR[frD] = -c.regs.FPR[frB]
		c.regs.PS1[frD] = -c.regs.PS1[frB]
		return 4
	})
	setExt4(264, func(c *CPU, instr uint32) int { // ps_abs
		frD, frB := fieldRD(instr), fieldRB(instr)
		c.regs.FPR[frD] = math.Abs(c.regs.FPR[frB])
		c.regs.PS1[frD] = math.Abs(c.regs.PS1[frB])
		return 4
	})

	setPrimary(56, func(c *CPU, instr uint32) int { // psq_l (simplified - see package doc)
		addr := c.effectiveAddr(instr)
		frD := fieldRD(instr)
		c.regs.FPR[frD] = float64(math.Float32frombits(c.read32(addr)))
		c.regs.PS1[frD] = float64(math.Float32frombits(c.read32(addr + 4)))
		return 4
	})
	setPrimary(60, func(c *CPU, instr uint32) int { // psq_st (simplified - see package doc)
		addr := c.effectiveAddr(instr)
		frD := fieldRD(instr)
		c.write32(addr, math.Float32bits(float32(c.regs.FPR[frD])))
		c.write32(addr+4, math.Float32bits(float32(c.regs.PS1[frD])))
		return 4
	})
}
