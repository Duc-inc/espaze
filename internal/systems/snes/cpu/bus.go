package cpu

// Bus is the 65816's 24-bit address space (bank<<16 | offset), an
// 8-bit data bus like every real 6502-family chip - this project
// composes wider reads/writes from Read8/Write8 the same way its
// other CPU cores do.
type Bus interface {
	Read8(addr uint32) byte
	Write8(addr uint32, v byte)
}
