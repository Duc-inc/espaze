package powerpc

// sprField decodes mfspr/mtspr's split 10-bit SPR number (its low 5
// bits live in the instruction's rA field, its high 5 bits in rB) -
// this project only implements the two most commonly used special
// registers, LR (8) and CTR (9).
func sprField(instr uint32) uint32 { return fieldRA(instr) | fieldRB(instr)<<5 }

func init() {
	setExt31(339, func(c *CPU, instr uint32) int { // mfspr
		rD := fieldRD(instr)
		switch sprField(instr) {
		case 8:
			c.regs.GPR[rD] = c.regs.LR
		case 9:
			c.regs.GPR[rD] = c.regs.CTR
		}
		return 2
	})
	setExt31(467, func(c *CPU, instr uint32) int { // mtspr
		rS := c.regs.GPR[fieldRD(instr)]
		switch sprField(instr) {
		case 8:
			c.regs.LR = rS
		case 9:
			c.regs.CTR = rS
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
