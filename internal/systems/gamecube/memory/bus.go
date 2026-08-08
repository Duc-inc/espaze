// Package memory implements the GameCube's physical memory map as
// seen by its Gekko CPU (internal/systems/powerpc): 24MB of MEM1 main
// RAM, plus a hardware register window (0xCC000000+) that routes to
// whatever peripherals Attach wires in (vi/si/di/ai packages). An
// address in that window with no attached peripheral is accepted (so
// code that pokes an unmodeled register doesn't fault) but has no
// effect. This package on its own can run arbitrary PowerPC code and
// touch RAM; whether it can display/read/hear anything real depends
// on which peripherals the caller has attached.
package memory

const (
	mem1Base = 0x00000000
	mem1Size = 24 * 1024 * 1024

	hwRegBase = 0xCC000000
	hwRegSize = 0x00010000
)

// Peripheral is a memory-mapped I/O device addressable within the
// hardware register window, word-addressed the way real GameCube
// peripherals are (vi.VI, si.SI, di.DI, ai.AI all implement this).
type Peripheral interface {
	Read32(offset uint32) uint32
	Write32(offset uint32, val uint32)
}

type region struct {
	base, size uint32
	dev        Peripheral
}

// Bus implements powerpc.Bus.
type Bus struct {
	mem1        [mem1Size]byte
	peripherals []region
}

// New returns a Bus with MEM1 zeroed and no peripherals attached.
func New() *Bus { return &Bus{} }

// Reset clears MEM1. Attached peripherals keep their own state -
// callers that want those reset too should reset them directly.
func (b *Bus) Reset() { b.mem1 = [mem1Size]byte{} }

// Attach wires a peripheral into the hardware register window at
// [base, base+size).
func (b *Bus) Attach(base, size uint32, dev Peripheral) {
	b.peripherals = append(b.peripherals, region{base, size, dev})
}

func (b *Bus) inMEM1(addr uint32) bool {
	return addr >= mem1Base && addr < mem1Base+mem1Size
}

func (b *Bus) peripheralAt(addr uint32) (Peripheral, uint32) {
	for _, r := range b.peripherals {
		if addr >= r.base && addr < r.base+r.size {
			return r.dev, r.base
		}
	}
	return nil, 0
}

// Read8 implements powerpc.Bus.
func (b *Bus) Read8(addr uint32) byte {
	if b.inMEM1(addr) {
		return b.mem1[addr-mem1Base]
	}
	if dev, base := b.peripheralAt(addr); dev != nil {
		off := addr - base
		word := dev.Read32(off &^ 3)
		return byte(word >> (24 - 8*(off&3)))
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
	default:
		if dev, base := b.peripheralAt(addr); dev != nil {
			off := addr - base
			wordOff := off &^ 3
			shift := 24 - 8*(off&3)
			word := dev.Read32(wordOff)
			word = word&^(0xFF<<shift) | uint32(v)<<shift
			dev.Write32(wordOff, word)
		}
		// Unmodeled hardware register: accepted, no effect.
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
