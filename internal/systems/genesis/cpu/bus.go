package cpu

// Bus is the 68000's memory interface. The real chip has a 16-bit
// external data bus (byte accesses are still possible, just less
// efficient), and word/long accesses must land on an even address;
// this project doesn't model the resulting bus-error trap for
// misaligned access; a misaligned Read/Write is just Read/Write.
type Bus interface {
	Read8(addr uint32) byte
	Read16(addr uint32) uint16
	Read32(addr uint32) uint32
	Write8(addr uint32, v byte)
	Write16(addr uint32, v uint16)
	Write32(addr uint32, v uint32)
}
