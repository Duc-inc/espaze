package powerpc

// Field extraction helpers, using PowerPC's own documented bit
// numbering (bit 0 = most significant bit of the 32-bit instruction
// word) and D-form/X-form/B-form/I-form layouts from the Power ISA
// manual.
func fieldRD(instr uint32) uint32    { return (instr >> 21) & 0x1F }
func fieldRA(instr uint32) uint32    { return (instr >> 16) & 0x1F }
func fieldRB(instr uint32) uint32    { return (instr >> 11) & 0x1F }
func fieldRC(instr uint32) bool      { return instr&1 != 0 }
func fieldExtOp(instr uint32) uint32 { return (instr >> 1) & 0x3FF }

func fieldSimm(instr uint32) int32  { return int32(int16(instr)) }
func fieldUimm(instr uint32) uint32 { return instr & 0xFFFF }

// fieldLI extracts I-form's 24-bit signed branch offset (bits 6-29),
// already shifted left 2 (branch targets are always word-aligned).
func fieldLI(instr uint32) int32 {
	raw := instr & 0x03FFFFFC
	if raw&0x02000000 != 0 {
		raw |= 0xFC000000
	}
	return int32(raw)
}

// fieldBD extracts B-form's 14-bit signed conditional-branch offset
// (bits 16-29), shifted left 2.
func fieldBD(instr uint32) int32 {
	raw := instr & 0x0000FFFC
	if raw&0x00008000 != 0 {
		raw |= 0xFFFF0000
	}
	return int32(raw)
}

func fieldBO(instr uint32) uint32 { return (instr >> 21) & 0x1F }
func fieldBI(instr uint32) uint32 { return (instr >> 16) & 0x1F }
func fieldAA(instr uint32) bool   { return instr&0x02 != 0 }
func fieldLK(instr uint32) bool   { return instr&0x01 != 0 }

// fieldFRC and fieldAFormXO are specific to A-form floating-point
// instructions (fadd/fsub/fmul/fdiv and similar): frD/frA/frB sit at
// the same bit positions fieldRD/fieldRA/fieldRB already extract, but
// A-form adds a third source register frC (bits 21-25) and its
// extended opcode is only 5 bits (bits 26-30), not the 10-bit field
// most other extended opcodes use.
func fieldFRC(instr uint32) uint32     { return (instr >> 6) & 0x1F }
func fieldAFormXO(instr uint32) uint32 { return (instr >> 1) & 0x1F }
