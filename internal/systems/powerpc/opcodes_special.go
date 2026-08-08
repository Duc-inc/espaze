package powerpc

// sprField decodes mfspr/mtspr's split 10-bit SPR number (its low 5
// bits live in the instruction's rA field, its high 5 bits in rB) -
// this project implements LR (8), CTR (9), SRR0/SRR1 (26/27, see
// exceptions.go), and the 8 BAT register pairs (528-543, real
// hardware's own SPR numbers for IBAT0-3/DBAT0-3 upper/lower halves -
// see mmu.go).
func sprField(instr uint32) uint32 { return fieldRA(instr) | fieldRB(instr)<<5 }

// batIndex reports which of the 8 bat entries an SPR number in
// 528-543 addresses, and whether it's the upper (BEPI/BL/valid) or
// lower (BRPN) half.
func batIndex(spr uint32) (index int, upper bool, ok bool) {
	if spr < 528 || spr > 543 {
		return 0, false, false
	}
	off := spr - 528
	return int(off / 2), off%2 == 0, true
}

func init() {
	setExt31(339, func(c *CPU, instr uint32) int { // mfspr
		rD := fieldRD(instr)
		spr := sprField(instr)
		switch spr {
		case 8:
			c.regs.GPR[rD] = c.regs.LR
		case 9:
			c.regs.GPR[rD] = c.regs.CTR
		case 26:
			c.regs.GPR[rD] = c.regs.SRR0
		case 27:
			c.regs.GPR[rD] = c.regs.SRR1
		default:
			if idx, upper, ok := batIndex(spr); ok {
				if upper {
					c.regs.GPR[rD] = encodeBATUpper(c.regs.bats[idx])
				} else {
					c.regs.GPR[rD] = encodeBATLower(c.regs.bats[idx])
				}
			}
		}
		return 2
	})
	setExt31(467, func(c *CPU, instr uint32) int { // mtspr
		rS := c.regs.GPR[fieldRD(instr)]
		spr := sprField(instr)
		switch spr {
		case 8:
			c.regs.LR = rS
		case 9:
			c.regs.CTR = rS
		case 26:
			c.regs.SRR0 = rS
		case 27:
			c.regs.SRR1 = rS
		default:
			if idx, upper, ok := batIndex(spr); ok {
				b := &c.regs.bats[idx]
				if upper {
					b.effective, b.length, b.valid = decodeBATUpper(rS)
				} else {
					b.real = decodeBATLower(rS)
				}
			}
		}
		return 2
	})
	setExt31(19, func(c *CPU, instr uint32) int { // mfcr
		c.regs.GPR[fieldRD(instr)] = c.regs.CR
		return 2
	})

	setPrimary(17, func(c *CPU, instr uint32) int { // sc
		if c.SyscallHandler != nil {
			c.SyscallHandler(c)
		}
		return 2
	})
}
