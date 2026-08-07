package cpu

// location is an effective address already resolved to either a
// register or a memory address - resolved once per operand access
// (resolveEA) and reused for both the read and the write half of a
// read-modify-write instruction, since re-resolving would double-apply
// a postincrement/predecrement mode's side effect.
type location struct {
	isRegister bool
	regIsAddr  bool // only meaningful when isRegister
	reg        byte
	addr       uint32
}

func sizeBytes(size byte) uint32 {
	switch size {
	case sizeByte:
		return 1
	case sizeLong:
		return 4
	default:
		return 2
	}
}

// resolveEA decodes a 3-bit mode + 3-bit register field (mode 7 uses
// the register field to select among 5 more addressing modes) into a
// location, fetching any extension words and applying any
// pre/post-decrement side effect along the way.
func (c *CPU) resolveEA(mode, reg byte, size byte) location {
	switch mode {
	case 0:
		return location{isRegister: true, reg: reg}
	case 1:
		return location{isRegister: true, regIsAddr: true, reg: reg}
	case 2:
		return location{addr: c.regs.A[reg]}
	case 3: // (An)+
		addr := c.regs.A[reg]
		inc := sizeBytes(size)
		if reg == 7 && size == sizeByte {
			inc = 2 // A7 (the stack pointer) always stays word-aligned
		}
		c.regs.A[reg] += inc
		return location{addr: addr}
	case 4: // -(An)
		dec := sizeBytes(size)
		if reg == 7 && size == sizeByte {
			dec = 2
		}
		c.regs.A[reg] -= dec
		return location{addr: c.regs.A[reg]}
	case 5: // (d16,An)
		disp := int32(int16(c.fetchWord()))
		return location{addr: uint32(int32(c.regs.A[reg]) + disp)}
	case 6: // (d8,An,Xn)
		return location{addr: c.indexedAddr(c.regs.A[reg])}
	default: // mode 7: reg selects between several more modes
		switch reg {
		case 0: // (xxx).W
			addr := uint32(int32(int16(c.fetchWord())))
			return location{addr: addr}
		case 1: // (xxx).L
			return location{addr: c.fetchLong()}
		case 2: // (d16,PC)
			base := c.regs.PC
			disp := int32(int16(c.fetchWord()))
			return location{addr: uint32(int32(base) + disp)}
		case 3: // (d8,PC,Xn)
			return location{addr: c.indexedAddr(c.regs.PC)}
		default: // 4: #immediate - only ever valid as a source, see readEA
			return location{}
		}
	}
}

// indexedAddr implements the (d8,base,Xn) family shared by address- and
// PC-relative indexed modes: an extension word selects an index
// register (data or address) and whether to use it sign-extended as a
// word or as the full long, plus an 8-bit displacement.
func (c *CPU) indexedAddr(base uint32) uint32 {
	ext := c.fetchWord()
	xnIsAddr := ext&0x8000 != 0
	xnNum := byte(ext>>12) & 0x07
	longIndex := ext&0x0800 != 0
	disp := int32(int8(ext))

	var xn uint32
	if xnIsAddr {
		xn = c.regs.A[xnNum]
	} else {
		xn = c.regs.D[xnNum]
	}
	if !longIndex {
		xn = uint32(int32(int16(xn)))
	}

	return uint32(int32(base) + int32(xn) + disp)
}

// readEA reads size bytes from a freshly resolved location - immediate
// mode is handled here rather than in resolveEA, since it never
// resolves to an address at all.
func (c *CPU) readEA(mode, reg byte, size byte) uint32 {
	if mode == 7 && reg == 4 {
		return c.fetchImmediate(size)
	}
	return c.readLocation(c.resolveEA(mode, reg, size), size)
}

func (c *CPU) fetchImmediate(size byte) uint32 {
	switch size {
	case sizeByte:
		return uint32(byte(c.fetchWord()))
	case sizeLong:
		return c.fetchLong()
	default:
		return uint32(c.fetchWord())
	}
}

func (c *CPU) readLocation(loc location, size byte) uint32 {
	if loc.isRegister {
		if loc.regIsAddr {
			return c.regs.A[loc.reg]
		}
		return maskSize(c.regs.D[loc.reg], size)
	}
	switch size {
	case sizeByte:
		return uint32(c.bus.Read8(loc.addr))
	case sizeLong:
		return c.bus.Read32(loc.addr)
	default:
		return uint32(c.bus.Read16(loc.addr))
	}
}

// writeLocation writes size bytes into loc. Writing a byte or word into
// a *data* register only replaces those low bits, leaving the rest of
// the 32-bit register alone - real 68000 behavior software sometimes
// deliberately relies on (packing multiple byte values into one Dn).
// Address registers don't have this quirk: any write there sign-extends
// to fill all 32 bits.
func (c *CPU) writeLocation(loc location, size byte, v uint32) {
	if loc.isRegister {
		if loc.regIsAddr {
			c.regs.A[loc.reg] = signExtend(v, size)
			return
		}
		c.regs.D[loc.reg] = mergeSize(c.regs.D[loc.reg], v, size)
		return
	}
	switch size {
	case sizeByte:
		c.bus.Write8(loc.addr, byte(v))
	case sizeLong:
		c.bus.Write32(loc.addr, v)
	default:
		c.bus.Write16(loc.addr, uint16(v))
	}
}

func maskSize(v uint32, size byte) uint32 {
	switch size {
	case sizeByte:
		return v & 0xFF
	case sizeLong:
		return v
	default:
		return v & 0xFFFF
	}
}

func mergeSize(old, v uint32, size byte) uint32 {
	switch size {
	case sizeByte:
		return old&0xFFFFFF00 | v&0xFF
	case sizeLong:
		return v
	default:
		return old&0xFFFF0000 | v&0xFFFF
	}
}

func signExtend(v uint32, size byte) uint32 {
	switch size {
	case sizeByte:
		return uint32(int32(int8(v)))
	case sizeLong:
		return v
	default:
		return uint32(int32(int16(v)))
	}
}
