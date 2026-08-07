package cpu

// Flag bits within F.
const (
	FlagC  byte = 1 << 0 // carry
	FlagN  byte = 1 << 1 // subtract
	FlagPV byte = 1 << 2 // parity/overflow
	FlagX  byte = 1 << 3 // undocumented, copies bit 3 of the ALU result
	FlagH  byte = 1 << 4 // half-carry
	FlagY  byte = 1 << 5 // undocumented, copies bit 5 of the ALU result
	FlagZ  byte = 1 << 6 // zero
	FlagS  byte = 1 << 7 // sign
)

// registers holds the Z80's entire visible state: the main and shadow
// 8080-heritage register sets (swappable via EX AF,AF' / EXX), the two
// index registers, and the special-purpose ones.
type registers struct {
	A, F byte
	B, C byte
	D, E byte
	H, L byte

	A2, F2 byte
	B2, C2 byte
	D2, E2 byte
	H2, L2 byte

	IX, IY uint16
	SP, PC uint16
	I, R   byte

	IFF1, IFF2 bool // interrupt enable flip-flops
	IM         byte // interrupt mode: 0, 1, or 2
}

func (r *registers) setFlag(flag byte, on bool) {
	if on {
		r.F |= flag
	} else {
		r.F &^= flag
	}
}

func (r *registers) getFlag(flag byte) bool { return r.F&flag != 0 }

func (r *registers) BC() uint16 { return uint16(r.B)<<8 | uint16(r.C) }
func (r *registers) DE() uint16 { return uint16(r.D)<<8 | uint16(r.E) }
func (r *registers) HL() uint16 { return uint16(r.H)<<8 | uint16(r.L) }
func (r *registers) AF() uint16 { return uint16(r.A)<<8 | uint16(r.F) }

func (r *registers) SetBC(v uint16) { r.B, r.C = byte(v>>8), byte(v) }
func (r *registers) SetDE(v uint16) { r.D, r.E = byte(v>>8), byte(v) }
func (r *registers) SetHL(v uint16) { r.H, r.L = byte(v>>8), byte(v) }
func (r *registers) SetAF(v uint16) { r.A, r.F = byte(v>>8), byte(v) }

// setSZ sets Sign and Zero from a just-computed byte result - the most
// common flag pattern, shared by nearly every 8-bit ALU/load operation
// that touches flags at all.
func (r *registers) setSZ(v byte) {
	r.setFlag(FlagS, v&0x80 != 0)
	r.setFlag(FlagZ, v == 0)
}

// setYX copies the undocumented Y/X flags from bits 5/3 of v - real
// hardware does this after nearly every operation that sets S/Z, so
// callers that need flag-perfect behavior (some games' copy-protection
// checks do) get it by calling this alongside setSZ.
func (r *registers) setYX(v byte) {
	r.setFlag(FlagY, v&0x20 != 0)
	r.setFlag(FlagX, v&0x08 != 0)
}

func parity(v byte) bool {
	v ^= v >> 4
	v ^= v >> 2
	v ^= v >> 1
	return v&1 == 0
}
