package cpu

// Bus is everything the CPU can read and write: RAM, PPU/APU registers,
// and cartridge space, all mapped into one 16-bit address space by
// whoever wires the CPU up (the nes package, not this one - the CPU
// itself has no idea what's behind an address).
type Bus interface {
	Read(addr uint16) byte
	Write(addr uint16, v byte)
}
