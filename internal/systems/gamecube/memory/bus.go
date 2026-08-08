// Package memory implements the GameCube's physical memory map as
// seen by its Gekko CPU (internal/systems/powerpc): 24MB of MEM1 main
// RAM, and a stubbed-out hardware register window (0xCC000000+) that
// currently does nothing - filling that in with real behavior for the
// Flipper GPU's Command Processor, the DSP, and the other peripherals
// is future work, not attempted here. This package on its own can run
// arbitrary PowerPC code and touch RAM; it cannot boot or display
// anything from a real GameCube disc image, since nothing here reads
// one yet.
package memory

const (
	mem1Base = 0x00000000
	mem1Size = 24 * 1024 * 1024

	hwRegBase = 0xCC000000
	hwRegSize = 0x00010000
)

// Bus implements powerpc.Bus.
type Bus struct {
	mem1 [mem1Size]byte
}

// New returns a Bus with MEM1 zeroed.
func New() *Bus { return &Bus{} }

// Reset clears MEM1.
func (b *Bus) Reset() { b.mem1 = [mem1Size]byte{} }

func (b *Bus) inMEM1(addr uint32) bool {
	return addr >= mem1Base && addr < mem1Base+mem1Size
}

func (b *Bus) inHWRegs(addr uint32) bool {
	return addr >= hwRegBase && addr < hwRegBase+hwRegSize
}

// Read8 implements powerpc.Bus.
func (b *Bus) Read8(addr uint32) byte {
	if b.inMEM1(addr) {
		return b.mem1[addr-mem1Base]
	}
	return 0
}

// Read16 implements powerpc.Bus.
func (b *Bus) Read16(addr uint32) uint16 {
	return uint16(b.Read8(addr))<<8 | uint16(b.Read8(addr+1))
}

// Read32 implements powerpc.Bus.
func (b *Bus) Read32(addr uint32) uint32 {
	return uint32(b.Read16(addr))<<16 | uint32(b.Read16(addr+2))
}

// Write8 implements powerpc.Bus.
func (b *Bus) Write8(addr uint32, v byte) {
	switch {
	case b.inMEM1(addr):
		b.mem1[addr-mem1Base] = v
	case b.inHWRegs(addr):
		// Accepted (so code that pokes hardware registers doesn't
		// fault) but has no effect yet - see this package's own doc
		// comment.
	}
}

// Write16 implements powerpc.Bus.
func (b *Bus) Write16(addr uint32, v uint16) {
	b.Write8(addr, byte(v>>8))
	b.Write8(addr+1, byte(v))
}

// Write32 implements powerpc.Bus.
func (b *Bus) Write32(addr uint32, v uint32) {
	b.Write16(addr, uint16(v>>16))
	b.Write16(addr+2, uint16(v))
}
