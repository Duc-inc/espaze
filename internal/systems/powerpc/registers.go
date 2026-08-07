// Package powerpc implements a from-scratch interpreter for the
// 32-bit PowerPC core shared by the GameCube (IBM "Gekko"), Wii
// ("Broadway"), and Wii U ("Espresso", a triple-core Broadway
// derivative this project treats as a single core). Unlike this
// project's Neo Geo Pocket or SNES audio cores, PowerPC's instruction
// encoding IS extensively and reliably documented (IBM's own Power
// ISA manuals), so this covers real opcode values with reasonable
// confidence for the common integer/branch/load-store instruction
// classes.
//
// This package is deliberately NOT wired into the app as a playable
// system. A GameCube/Wii/Wii U "core" needs a full 3D graphics
// pipeline (the GX/GX2 API - vertex buffers, shaders, texture
// mapping, rasterization) and a disc filesystem reader (often
// encrypted, for Wii and Wii U) before any commercial game shows a
// single pixel; neither exists in this project, and building either
// to a meaningful degree is a multi-year undertaking real projects
// (Dolphin) have dedicated large teams to. Registering this CPU alone
// as a "system" would let someone load a real game and see nothing
// happen, which is worse than not offering it at all. This package
// exists as real, tested groundwork for whoever wants to build the
// rest someday - not as a shipped feature.
package powerpc

// Condition register and XER bit helpers operate on the standard
// PowerPC bit-numbering convention (bit 0 = most significant).

// registers holds the 32-bit PowerPC integer register file: 32
// general-purpose registers, the link register (subroutine return
// address), count register (loop counter / branch target), condition
// register (8 4-bit fields, CR0 set by "." instruction variants),
// XER (carry/overflow/summary), and the program counter.
type registers struct {
	GPR [32]uint32
	LR  uint32
	CTR uint32
	CR  uint32
	XER uint32
	PC  uint32
}

// XER bit positions.
const (
	XERCarry    uint32 = 1 << 29
	XEROverflow uint32 = 1 << 30
	XERSummary  uint32 = 1 << 31
)

func (r *registers) getXER(bit uint32) bool { return r.XER&bit != 0 }

func (r *registers) setXER(bit uint32, on bool) {
	if on {
		r.XER |= bit
	} else {
		r.XER &^= bit
	}
}

// setCR0 updates condition register field 0 (bits 0-3) from a signed
// comparison against zero, plus XER's summary overflow bit copied
// into bit 3 - the standard behavior of every "." (record) instruction
// variant.
func (r *registers) setCR0(result uint32) {
	var field uint32
	switch {
	case int32(result) < 0:
		field = 0x8
	case int32(result) > 0:
		field = 0x4
	default:
		field = 0x2
	}
	if r.getXER(XERSummary) {
		field |= 0x1
	}
	r.CR = r.CR&0x0FFFFFFF | field<<28
}
