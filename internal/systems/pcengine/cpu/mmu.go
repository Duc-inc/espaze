package cpu

// mmu is the HuC6280's memory mapping unit: 8 page registers (MPR0-7),
// each mapping one 8KB slice of the CPU's 16-bit logical address space
// to any 8KB page of a 21-bit (2MB) physical space - the mechanism
// PC Engine software uses to bank cartridge ROM, RAM, and hardware
// registers into a fixed set of logical windows. TAM/TMA (see
// opcodes_system.go) are the only way software changes it.
type mmu struct {
	mpr [8]byte
}

// translate converts a 16-bit logical address into its 21-bit physical
// address through the page currently mapped at that logical page.
func (m *mmu) translate(logical uint16) uint32 {
	page := logical >> 13
	offset := uint32(logical & 0x1FFF)
	return uint32(m.mpr[page])<<13 | offset
}

func (m *mmu) writePages(mask byte, value byte) {
	for page := 0; page < 8; page++ {
		if mask&(1<<uint(page)) != 0 {
			m.mpr[page] = value
		}
	}
}

func (m *mmu) readPage(mask byte) byte {
	for page := 0; page < 8; page++ {
		if mask&(1<<uint(page)) != 0 {
			return m.mpr[page]
		}
	}
	return 0
}
