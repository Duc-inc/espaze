package powerpc

// primaryTable dispatches on the instruction's 6-bit primary opcode
// (bits 0-5) - the top level of PowerPC's own documented instruction
// format.
var primaryTable = map[uint32]func(c *CPU, instr uint32) int{}

// ext31Table/ext19Table dispatch primary opcode 31's (general integer/
// logical) and 19's (branch-to-LR/CTR) extended opcode field
// (bits 21-30) - the second level real PowerPC decode uses for these
// two opcode groups.
var ext31Table = map[uint32]func(c *CPU, instr uint32) int{}
var ext19Table = map[uint32]func(c *CPU, instr uint32) int{}

// ext63Table/ext63ATable dispatch primary opcode 63's (double-
// precision floating point) two different extended-opcode widths:
// X-form instructions (fmr/fneg/fabs/fcmpu) use the usual 10-bit
// field, but A-form arithmetic (fadd/fsub/fmul/fdiv) only has 5 bits
// of opcode, checked first since the two schemes would otherwise
// collide on some values.
var ext63Table = map[uint32]func(c *CPU, instr uint32) int{}
var ext63ATable = map[uint32]func(c *CPU, instr uint32) int{}

func setPrimary(op uint32, fn func(c *CPU, instr uint32) int)  { primaryTable[op] = fn }
func setExt31(extOp uint32, fn func(c *CPU, instr uint32) int) { ext31Table[extOp] = fn }
func setExt19(extOp uint32, fn func(c *CPU, instr uint32) int) { ext19Table[extOp] = fn }
func setExt63(extOp uint32, fn func(c *CPU, instr uint32) int) { ext63Table[extOp] = fn }
func setExt63A(xo uint32, fn func(c *CPU, instr uint32) int)   { ext63ATable[xo] = fn }

func init() {
	setPrimary(31, func(c *CPU, instr uint32) int {
		if fn, ok := ext31Table[fieldExtOp(instr)]; ok {
			return fn(c, instr)
		}
		return 2
	})
	setPrimary(19, func(c *CPU, instr uint32) int {
		if fn, ok := ext19Table[fieldExtOp(instr)]; ok {
			return fn(c, instr)
		}
		return 2
	})
	setPrimary(63, func(c *CPU, instr uint32) int {
		if fn, ok := ext63ATable[fieldAFormXO(instr)]; ok {
			return fn(c, instr)
		}
		if fn, ok := ext63Table[fieldExtOp(instr)]; ok {
			return fn(c, instr)
		}
		return 2
	})
}
