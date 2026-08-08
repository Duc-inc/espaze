// exceptions.go implements the one exception class this project needs
// to deliver GameCube peripheral interrupts (VI VBlank, AI sample
// match, etc.) into running code: the external interrupt. MSR's EE
// bit gates delivery, SRR0/SRR1 save the interrupted PC/MSR, and
// execution resumes at the standard Power ISA external-interrupt
// vector (0x500) - all facts from the general PowerPC architecture,
// not GameCube-specific. No other exception class (DSI, alignment,
// program, ...) is modeled, consistent with this project not
// implementing full address translation or exhaustive instruction
// validation.
package powerpc

// MSREE is the Machine State Register's External interrupt Enable
// bit (Power ISA bit 16, IBM bit-numbering).
const MSREE uint32 = 1 << 15

// ExternalInterruptVector is the standard Power ISA external
// interrupt exception vector address.
const ExternalInterruptVector = 0x00000500

// RaiseExternalInterrupt delivers the external interrupt exception if
// MSR[EE] is set. Unlike real hardware, an interrupt raised while EE
// is clear is dropped rather than latched pending - callers here
// (VI/AI Step, see gamecube.go) poll every CPU step, so a dropped
// edge is observable at most one step later.
func (c *CPU) RaiseExternalInterrupt() {
	if c.regs.MSR&MSREE == 0 {
		return
	}
	c.regs.SRR0 = c.regs.PC
	c.regs.SRR1 = c.regs.MSR
	c.regs.MSR &^= MSREE
	c.regs.PC = ExternalInterruptVector
}

func init() {
	setExt31(83, func(c *CPU, instr uint32) int { // mfmsr
		c.regs.GPR[fieldRD(instr)] = c.regs.MSR
		return 2
	})
	setExt31(146, func(c *CPU, instr uint32) int { // mtmsr
		c.regs.MSR = c.regs.GPR[fieldRD(instr)]
		return 2
	})
	setExt19(50, func(c *CPU, instr uint32) int { // rfi
		c.regs.PC = c.regs.SRR0
		c.regs.MSR = c.regs.SRR1
		return 2
	})
}
