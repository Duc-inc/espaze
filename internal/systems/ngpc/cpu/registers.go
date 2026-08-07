package cpu

// Status register flag bits, in the TLCS900H's own documented
// positions within the low byte of SR.
const (
	FlagCarry     byte = 1 << 0
	FlagHalfCarry byte = 1 << 1
	FlagOverflow  byte = 1 << 2
	FlagSign      byte = 1 << 3
	FlagZero      byte = 1 << 6
	FlagNegative  byte = 1 << 7 // "S" (subtract) flag some references call N
)

// registers holds one register bank's worth of the TLCS900H's general
// registers - XWA/XBC/XDE/XHL, each a 32-bit register with 16-bit (WA/
// BC/DE/HL) and 8-bit (W/A, B/C, D/E, H/L) views, exactly like real
// hardware. Real hardware actually has *two* banks, switched via the
// RFP bits in SR, plus a second full register file for the "maximum
// mode" interrupt system; this project implements a single active
// bank only - a deliberate simplification, since bank-switching is
// mostly used by interrupt handlers wanting a scratch set of
// registers without saving/restoring, not by normal game logic.
type registers struct {
	XWA, XBC, XDE, XHL uint32
	XIX, XIY, XIZ, XSP uint32
	PC                 uint32
	SR                 uint16 // low byte: flags; high byte: interrupt mask + bank select
}

func (r *registers) getFlag(flag byte) bool { return byte(r.SR)&flag != 0 }

func (r *registers) setFlag(flag byte, on bool) {
	if on {
		r.SR |= uint16(flag)
	} else {
		r.SR &^= uint16(flag)
	}
}

func (r *registers) setSZ8(v byte) {
	r.setFlag(FlagZero, v == 0)
	r.setFlag(FlagSign, v&0x80 != 0)
}

func (r *registers) setSZ16(v uint16) {
	r.setFlag(FlagZero, v == 0)
	r.setFlag(FlagSign, v&0x8000 != 0)
}

// reg8/setReg8 and reg16/setReg16 index into the 8 general registers'
// 8-bit and 16-bit views by the TLCS900H's own conventional register
// codes: 0=W/WA,1=A,2=B/BC,3=C,4=D/DE,5=E,6=H/HL,7=L for 8-bit views
// (W,A from XWA; B,C from XBC; D,E from XDE; H,L from XHL); 0-3 select
// WA/BC/DE/HL for 16-bit views.
func (r *registers) reg8(code byte) byte {
	pairs := [4]*uint32{&r.XWA, &r.XBC, &r.XDE, &r.XHL}
	p := pairs[(code>>1)&0x03]
	if code&1 == 0 {
		return byte(*p >> 8)
	}
	return byte(*p)
}

func (r *registers) setReg8(code byte, v byte) {
	pairs := [4]*uint32{&r.XWA, &r.XBC, &r.XDE, &r.XHL}
	p := pairs[(code>>1)&0x03]
	if code&1 == 0 {
		*p = *p&0xFFFF00FF | uint32(v)<<8
	} else {
		*p = *p&0xFFFFFF00 | uint32(v)
	}
}

func (r *registers) reg16(code byte) uint16 {
	pairs := [4]*uint32{&r.XWA, &r.XBC, &r.XDE, &r.XHL}
	return uint16(*pairs[code&0x03])
}

func (r *registers) setReg16(code byte, v uint16) {
	pairs := [4]*uint32{&r.XWA, &r.XBC, &r.XDE, &r.XHL}
	p := pairs[code&0x03]
	*p = *p&0xFFFF0000 | uint32(v)
}
