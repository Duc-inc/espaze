package powerpc

// Real hardware's memory-ordering barriers (sync, isync, eieio) exist
// to control instruction/data reordering across multiple cores and
// caches. This interpreter executes everything serially with no
// reordering or caching to control, so all three are genuine, correct
// no-ops here - not simplifications, just what a barrier means when
// there's nothing to hold back.
func init() {
	setExt31(598, func(c *CPU, instr uint32) int { return 2 }) // sync
	setExt31(854, func(c *CPU, instr uint32) int { return 2 }) // eieio
	setExt19(150, func(c *CPU, instr uint32) int { return 2 }) // isync
}
