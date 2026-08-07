package cpu

// dispatchEntry is one opcode byte's handler; each handler resolves
// its own addressing mode and returns the cycle cost, mirroring how
// this project's other CPU cores account for variable-length
// instructions.
type dispatchEntry struct {
	execute func(c *CPU) int
}

var dispatchTable [256]dispatchEntry

func setOp(opcode byte, fn func(c *CPU) int) {
	dispatchTable[opcode] = dispatchEntry{execute: fn}
}
