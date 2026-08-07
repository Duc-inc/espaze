package cpu

// shiftLeftReg/shiftRightReg/arithShiftRightReg/rotateRightReg
// implement the barrel shifter's four operations when the shift
// amount comes from a register (the low byte of it, per ARM's own
// rule) rather than an immediate - shared by Thumb's format 4 ALU
// shifts and ARM's register-specified operand2. Each updates the
// Carry flag in cpsr to the last bit shifted out, matching real
// hardware, except when the shift amount is 0 (Carry is left alone).
func setCarry(cpsr *uint32, out bool) {
	if out {
		*cpsr |= FlagC
	} else {
		*cpsr &^= FlagC
	}
}

func shiftLeftReg(v, amount uint32, cpsr *uint32) uint32 {
	amount &= 0xFF
	switch {
	case amount == 0:
		return v
	case amount < 32:
		setCarry(cpsr, v&(1<<(32-amount)) != 0)
		return v << amount
	case amount == 32:
		setCarry(cpsr, v&1 != 0)
		return 0
	default:
		setCarry(cpsr, false)
		return 0
	}
}

func shiftRightReg(v, amount uint32, cpsr *uint32) uint32 {
	amount &= 0xFF
	switch {
	case amount == 0:
		return v
	case amount < 32:
		setCarry(cpsr, v&(1<<(amount-1)) != 0)
		return v >> amount
	case amount == 32:
		setCarry(cpsr, v&0x80000000 != 0)
		return 0
	default:
		setCarry(cpsr, false)
		return 0
	}
}

func arithShiftRightReg(v, amount uint32, cpsr *uint32) uint32 {
	amount &= 0xFF
	if amount == 0 {
		return v
	}
	if amount >= 32 {
		neg := int32(v) < 0
		setCarry(cpsr, neg)
		if neg {
			return 0xFFFFFFFF
		}
		return 0
	}
	setCarry(cpsr, v&(1<<(amount-1)) != 0)
	return uint32(int32(v) >> amount)
}

func rotateRightReg(v, amount uint32, cpsr *uint32) uint32 {
	amount &= 0xFF
	if amount == 0 {
		return v
	}
	rot := amount % 32
	if rot == 0 {
		setCarry(cpsr, v&0x80000000 != 0)
		return v
	}
	result := v>>rot | v<<(32-rot)
	setCarry(cpsr, result&0x80000000 != 0)
	return result
}
