package cpu

func opCLC(c *CPU, _ addrMode, _ uint16, _ bool) int { c.regs.setFlag(FlagCarry, false); return 0 }
func opSEC(c *CPU, _ addrMode, _ uint16, _ bool) int { c.regs.setFlag(FlagCarry, true); return 0 }
func opCLI(c *CPU, _ addrMode, _ uint16, _ bool) int { c.regs.setFlag(FlagInterrupt, false); return 0 }
func opSEI(c *CPU, _ addrMode, _ uint16, _ bool) int { c.regs.setFlag(FlagInterrupt, true); return 0 }
func opCLV(c *CPU, _ addrMode, _ uint16, _ bool) int { c.regs.setFlag(FlagOverflow, false); return 0 }
func opCLD(c *CPU, _ addrMode, _ uint16, _ bool) int { c.regs.setFlag(FlagDecimal, false); return 0 }
func opSED(c *CPU, _ addrMode, _ uint16, _ bool) int { c.regs.setFlag(FlagDecimal, true); return 0 }
