package cpu

// opcodeEntry describes one byte's worth of the 6502's instruction set:
// which addressing mode decodes its operand, its base cycle cost, and
// whether that cost gets +1 when resolveOperand crosses a page boundary.
type opcodeEntry struct {
	name             string
	mode             addrMode
	cycles           int
	extraOnPageCross bool
	execute          func(c *CPU, mode addrMode, addr uint16, pageCrossed bool) int
}

// opcodeTable covers every official 6502 opcode. The 105 byte values
// with no official meaning are left as their zero value here and filled
// in by the init() below as a 2-cycle NOP - real hardware does *something*
// well-defined with each of them, but nothing this project's cores or
// commercial NES games routinely depend on, so they're deferred.
var opcodeTable = [256]opcodeEntry{
	0x69: {"ADC", modeImmediate, 2, false, opADC},
	0x65: {"ADC", modeZeroPage, 3, false, opADC},
	0x75: {"ADC", modeZeroPageX, 4, false, opADC},
	0x6D: {"ADC", modeAbsolute, 4, false, opADC},
	0x7D: {"ADC", modeAbsoluteX, 4, true, opADC},
	0x79: {"ADC", modeAbsoluteY, 4, true, opADC},
	0x61: {"ADC", modeIndirectX, 6, false, opADC},
	0x71: {"ADC", modeIndirectY, 5, true, opADC},

	0x29: {"AND", modeImmediate, 2, false, opAND},
	0x25: {"AND", modeZeroPage, 3, false, opAND},
	0x35: {"AND", modeZeroPageX, 4, false, opAND},
	0x2D: {"AND", modeAbsolute, 4, false, opAND},
	0x3D: {"AND", modeAbsoluteX, 4, true, opAND},
	0x39: {"AND", modeAbsoluteY, 4, true, opAND},
	0x21: {"AND", modeIndirectX, 6, false, opAND},
	0x31: {"AND", modeIndirectY, 5, true, opAND},

	0x0A: {"ASL", modeAccumulator, 2, false, opASL},
	0x06: {"ASL", modeZeroPage, 5, false, opASL},
	0x16: {"ASL", modeZeroPageX, 6, false, opASL},
	0x0E: {"ASL", modeAbsolute, 6, false, opASL},
	0x1E: {"ASL", modeAbsoluteX, 7, false, opASL},

	0x90: {"BCC", modeRelative, 2, false, opBCC},
	0xB0: {"BCS", modeRelative, 2, false, opBCS},
	0xF0: {"BEQ", modeRelative, 2, false, opBEQ},
	0x30: {"BMI", modeRelative, 2, false, opBMI},
	0xD0: {"BNE", modeRelative, 2, false, opBNE},
	0x10: {"BPL", modeRelative, 2, false, opBPL},
	0x50: {"BVC", modeRelative, 2, false, opBVC},
	0x70: {"BVS", modeRelative, 2, false, opBVS},

	0x24: {"BIT", modeZeroPage, 3, false, opBIT},
	0x2C: {"BIT", modeAbsolute, 4, false, opBIT},

	0x00: {"BRK", modeImplied, 7, false, opBRK},

	0x18: {"CLC", modeImplied, 2, false, opCLC},
	0xD8: {"CLD", modeImplied, 2, false, opCLD},
	0x58: {"CLI", modeImplied, 2, false, opCLI},
	0xB8: {"CLV", modeImplied, 2, false, opCLV},
	0x38: {"SEC", modeImplied, 2, false, opSEC},
	0xF8: {"SED", modeImplied, 2, false, opSED},
	0x78: {"SEI", modeImplied, 2, false, opSEI},

	0xC9: {"CMP", modeImmediate, 2, false, opCMP},
	0xC5: {"CMP", modeZeroPage, 3, false, opCMP},
	0xD5: {"CMP", modeZeroPageX, 4, false, opCMP},
	0xCD: {"CMP", modeAbsolute, 4, false, opCMP},
	0xDD: {"CMP", modeAbsoluteX, 4, true, opCMP},
	0xD9: {"CMP", modeAbsoluteY, 4, true, opCMP},
	0xC1: {"CMP", modeIndirectX, 6, false, opCMP},
	0xD1: {"CMP", modeIndirectY, 5, true, opCMP},

	0xE0: {"CPX", modeImmediate, 2, false, opCPX},
	0xE4: {"CPX", modeZeroPage, 3, false, opCPX},
	0xEC: {"CPX", modeAbsolute, 4, false, opCPX},

	0xC0: {"CPY", modeImmediate, 2, false, opCPY},
	0xC4: {"CPY", modeZeroPage, 3, false, opCPY},
	0xCC: {"CPY", modeAbsolute, 4, false, opCPY},

	0xC6: {"DEC", modeZeroPage, 5, false, opDEC},
	0xD6: {"DEC", modeZeroPageX, 6, false, opDEC},
	0xCE: {"DEC", modeAbsolute, 6, false, opDEC},
	0xDE: {"DEC", modeAbsoluteX, 7, false, opDEC},
	0xCA: {"DEX", modeImplied, 2, false, opDEX},
	0x88: {"DEY", modeImplied, 2, false, opDEY},

	0x49: {"EOR", modeImmediate, 2, false, opEOR},
	0x45: {"EOR", modeZeroPage, 3, false, opEOR},
	0x55: {"EOR", modeZeroPageX, 4, false, opEOR},
	0x4D: {"EOR", modeAbsolute, 4, false, opEOR},
	0x5D: {"EOR", modeAbsoluteX, 4, true, opEOR},
	0x59: {"EOR", modeAbsoluteY, 4, true, opEOR},
	0x41: {"EOR", modeIndirectX, 6, false, opEOR},
	0x51: {"EOR", modeIndirectY, 5, true, opEOR},

	0xE6: {"INC", modeZeroPage, 5, false, opINC},
	0xF6: {"INC", modeZeroPageX, 6, false, opINC},
	0xEE: {"INC", modeAbsolute, 6, false, opINC},
	0xFE: {"INC", modeAbsoluteX, 7, false, opINC},
	0xE8: {"INX", modeImplied, 2, false, opINX},
	0xC8: {"INY", modeImplied, 2, false, opINY},

	0x4C: {"JMP", modeAbsolute, 3, false, opJMP},
	0x6C: {"JMP", modeIndirect, 5, false, opJMP},
	0x20: {"JSR", modeAbsolute, 6, false, opJSR},

	0xA9: {"LDA", modeImmediate, 2, false, opLDA},
	0xA5: {"LDA", modeZeroPage, 3, false, opLDA},
	0xB5: {"LDA", modeZeroPageX, 4, false, opLDA},
	0xAD: {"LDA", modeAbsolute, 4, false, opLDA},
	0xBD: {"LDA", modeAbsoluteX, 4, true, opLDA},
	0xB9: {"LDA", modeAbsoluteY, 4, true, opLDA},
	0xA1: {"LDA", modeIndirectX, 6, false, opLDA},
	0xB1: {"LDA", modeIndirectY, 5, true, opLDA},

	0xA2: {"LDX", modeImmediate, 2, false, opLDX},
	0xA6: {"LDX", modeZeroPage, 3, false, opLDX},
	0xB6: {"LDX", modeZeroPageY, 4, false, opLDX},
	0xAE: {"LDX", modeAbsolute, 4, false, opLDX},
	0xBE: {"LDX", modeAbsoluteY, 4, true, opLDX},

	0xA0: {"LDY", modeImmediate, 2, false, opLDY},
	0xA4: {"LDY", modeZeroPage, 3, false, opLDY},
	0xB4: {"LDY", modeZeroPageX, 4, false, opLDY},
	0xAC: {"LDY", modeAbsolute, 4, false, opLDY},
	0xBC: {"LDY", modeAbsoluteX, 4, true, opLDY},

	0x4A: {"LSR", modeAccumulator, 2, false, opLSR},
	0x46: {"LSR", modeZeroPage, 5, false, opLSR},
	0x56: {"LSR", modeZeroPageX, 6, false, opLSR},
	0x4E: {"LSR", modeAbsolute, 6, false, opLSR},
	0x5E: {"LSR", modeAbsoluteX, 7, false, opLSR},

	0xEA: {"NOP", modeImplied, 2, false, opNOP},

	0x09: {"ORA", modeImmediate, 2, false, opORA},
	0x05: {"ORA", modeZeroPage, 3, false, opORA},
	0x15: {"ORA", modeZeroPageX, 4, false, opORA},
	0x0D: {"ORA", modeAbsolute, 4, false, opORA},
	0x1D: {"ORA", modeAbsoluteX, 4, true, opORA},
	0x19: {"ORA", modeAbsoluteY, 4, true, opORA},
	0x01: {"ORA", modeIndirectX, 6, false, opORA},
	0x11: {"ORA", modeIndirectY, 5, true, opORA},

	0x48: {"PHA", modeImplied, 3, false, opPHA},
	0x08: {"PHP", modeImplied, 3, false, opPHP},
	0x68: {"PLA", modeImplied, 4, false, opPLA},
	0x28: {"PLP", modeImplied, 4, false, opPLP},

	0x2A: {"ROL", modeAccumulator, 2, false, opROL},
	0x26: {"ROL", modeZeroPage, 5, false, opROL},
	0x36: {"ROL", modeZeroPageX, 6, false, opROL},
	0x2E: {"ROL", modeAbsolute, 6, false, opROL},
	0x3E: {"ROL", modeAbsoluteX, 7, false, opROL},

	0x6A: {"ROR", modeAccumulator, 2, false, opROR},
	0x66: {"ROR", modeZeroPage, 5, false, opROR},
	0x76: {"ROR", modeZeroPageX, 6, false, opROR},
	0x6E: {"ROR", modeAbsolute, 6, false, opROR},
	0x7E: {"ROR", modeAbsoluteX, 7, false, opROR},

	0x40: {"RTI", modeImplied, 6, false, opRTI},
	0x60: {"RTS", modeImplied, 6, false, opRTS},

	0xE9: {"SBC", modeImmediate, 2, false, opSBC},
	0xE5: {"SBC", modeZeroPage, 3, false, opSBC},
	0xF5: {"SBC", modeZeroPageX, 4, false, opSBC},
	0xED: {"SBC", modeAbsolute, 4, false, opSBC},
	0xFD: {"SBC", modeAbsoluteX, 4, true, opSBC},
	0xF9: {"SBC", modeAbsoluteY, 4, true, opSBC},
	0xE1: {"SBC", modeIndirectX, 6, false, opSBC},
	0xF1: {"SBC", modeIndirectY, 5, true, opSBC},

	0x85: {"STA", modeZeroPage, 3, false, opSTA},
	0x95: {"STA", modeZeroPageX, 4, false, opSTA},
	0x8D: {"STA", modeAbsolute, 4, false, opSTA},
	0x9D: {"STA", modeAbsoluteX, 5, false, opSTA},
	0x99: {"STA", modeAbsoluteY, 5, false, opSTA},
	0x81: {"STA", modeIndirectX, 6, false, opSTA},
	0x91: {"STA", modeIndirectY, 6, false, opSTA},

	0x86: {"STX", modeZeroPage, 3, false, opSTX},
	0x96: {"STX", modeZeroPageY, 4, false, opSTX},
	0x8E: {"STX", modeAbsolute, 4, false, opSTX},

	0x84: {"STY", modeZeroPage, 3, false, opSTY},
	0x94: {"STY", modeZeroPageX, 4, false, opSTY},
	0x8C: {"STY", modeAbsolute, 4, false, opSTY},

	0xAA: {"TAX", modeImplied, 2, false, opTAX},
	0xA8: {"TAY", modeImplied, 2, false, opTAY},
	0xBA: {"TSX", modeImplied, 2, false, opTSX},
	0x8A: {"TXA", modeImplied, 2, false, opTXA},
	0x9A: {"TXS", modeImplied, 2, false, opTXS},
	0x98: {"TYA", modeImplied, 2, false, opTYA},
}

func init() {
	for i := range opcodeTable {
		if opcodeTable[i].execute == nil {
			opcodeTable[i] = opcodeEntry{name: "ILLEGAL", mode: modeImplied, cycles: 2, execute: opNOP}
		}
	}
}
