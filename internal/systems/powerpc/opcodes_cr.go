// opcodes_cr.go implements the condition register logical
// instructions (crand/cror/crxor/crnand/crnor/creqv/crandc/crorc) and
// mcrf - standard PowerPC XL-form instructions real compiled code
// uses to combine multiple comparison results (e.g. "if (a && b)")
// into a single bit a later bc can test, and to copy one CR field to
// another.
package powerpc

// crBit reads/writes one CR bit by its real hardware bit index (0-31,
// IBM bit-numbering: 0 = most significant) - the same convention
// fieldBI/fieldBO already use for bc's condition bit.
func (r *registers) crBit(i uint32) bool { return r.CR>>(31-i)&1 != 0 }

func (r *registers) setCRBit(i uint32, v bool) {
	if v {
		r.CR |= 1 << (31 - i)
	} else {
		r.CR &^= 1 << (31 - i)
	}
}

func init() {
	// XL-form: BT<<21 | BA<<16 | BB<<11 | xo<<1 - the same bit positions
	// fieldRD/fieldRA/fieldRB already extract, reused here for CR bit
	// indices instead of GPR numbers.
	crOp := func(xo uint32, f func(a, b bool) bool) {
		setExt19(xo, func(c *CPU, instr uint32) int {
			bt, ba, bb := fieldRD(instr), fieldRA(instr), fieldRB(instr)
			c.regs.setCRBit(bt, f(c.regs.crBit(ba), c.regs.crBit(bb)))
			return 2
		})
	}
	crOp(257, func(a, b bool) bool { return a && b })    // crand
	crOp(449, func(a, b bool) bool { return a || b })    // cror
	crOp(193, func(a, b bool) bool { return a != b })    // crxor
	crOp(225, func(a, b bool) bool { return !(a && b) }) // crnand
	crOp(33, func(a, b bool) bool { return !(a || b) })  // crnor
	crOp(289, func(a, b bool) bool { return a == b })    // creqv
	crOp(129, func(a, b bool) bool { return a && !b })   // crandc
	crOp(417, func(a, b bool) bool { return a || !b })   // crorc

	setExt19(0, func(c *CPU, instr uint32) int { // mcrf
		bf := fieldRD(instr) >> 2  // top 3 bits of the rD-shaped field: destination field 0-7
		bfa := fieldRA(instr) >> 2 // same for the source field
		shiftFrom := (7 - bfa) * 4
		shiftTo := (7 - bf) * 4
		value := (c.regs.CR >> shiftFrom) & 0xF
		c.regs.CR = c.regs.CR&^(0xF<<shiftTo) | value<<shiftTo
		return 2
	})
}
