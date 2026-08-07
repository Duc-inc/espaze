package spc700

// Bus is the SPC700's 16-bit address space (its own 64KB "ARAM",
// entirely separate from the main CPU's memory - the two only talk
// through 4 shared I/O ports).
type Bus interface {
	Read8(addr uint16) byte
	Write8(addr uint16, v byte)
}
