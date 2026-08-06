package cpu

// opFunc executes one decoded instruction and returns the T-cycles spent.
type opFunc func(c *CPU) int

// mainTable and cbTable are populated by init() functions spread across
// opcodes_*.go, grouped by instruction family so no single file has to
// spell out all 256 (or 512, with CB) entries by hand.
var mainTable [256]opFunc
var cbTable [256]opFunc

func (c *CPU) execute(opcode byte) int {
	if fn := mainTable[opcode]; fn != nil {
		return fn(c)
	}
	return 4 // undefined opcode: treated as a harmless no-op rather than crashing
}

func (c *CPU) executeCB(opcode byte) int {
	if fn := cbTable[opcode]; fn != nil {
		return fn(c)
	}
	return 8
}
