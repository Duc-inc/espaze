package cpu

// Bus is the TLCS900H's 24-bit address space (this project only
// actually addresses the low 16MB most NGPC software uses).
type Bus interface {
	Read8(addr uint32) byte
	Read16(addr uint32) uint16
	Write8(addr uint32, v byte)
	Write16(addr uint32, v uint16)
}
