package cpu

// Bus is the physical, post-MMU-translation 21-bit address space the
// HuC6280 actually reads and writes - cartridge ROM, work RAM, and the
// VDC/VCE/PSG/interrupt-controller registers all live somewhere in it,
// decoded by the top-level pcengine package.
type Bus interface {
	Read(addr uint32) byte
	Write(addr uint32, v byte)
}
