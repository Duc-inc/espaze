package powerpc

// Bus is PowerPC's 32-bit, big-endian address space.
type Bus interface {
	Read8(addr uint32) byte
	Read16(addr uint32) uint16
	Read32(addr uint32) uint32
	Write8(addr uint32, v byte)
	Write16(addr uint32, v uint16)
	Write32(addr uint32, v uint32)
}
