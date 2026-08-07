package cpu

// Bus is the Z80's 16-bit memory address space: RAM, the mapper's ROM
// banking registers, and cartridge RAM all live behind it.
type Bus interface {
	Read(addr uint16) byte
	Write(addr uint16, v byte)
}

// IOBus is the Z80's *separate* 8-bit I/O port space (IN/OUT
// instructions never touch memory) - the VDP, PSG and joypads all live
// here instead of in the memory map.
type IOBus interface {
	In(port byte) byte
	Out(port byte, v byte)
}
