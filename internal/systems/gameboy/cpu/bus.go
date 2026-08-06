package cpu

// Bus is the minimal interface the CPU needs from memory: a flat 16-bit
// address space. *memory.MMU implements this; the CPU package never
// imports the memory package, keeping the dependency one-directional.
type Bus interface {
	Read(addr uint16) byte
	Write(addr uint16, v byte)
}
