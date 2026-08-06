package memory

import (
	"github.com/Duc-inc/espaze/internal/emulation/input"
	"github.com/Duc-inc/espaze/internal/systems/nes/apu"
	"github.com/Duc-inc/espaze/internal/systems/nes/ppu"
)

// Bus is the NES's full CPU-visible address space: RAM, PPU/APU
// registers, both controller ports, and the cartridge (via its mapper).
type Bus struct {
	ram    [0x0800]byte
	prgRAM [0x2000]byte

	ppu    *ppu.PPU
	sound  *apu.APU
	mapper Mapper

	controller1, controller2 Controller

	stallCycles int
}

// New builds a Bus wired to the given components.
func New(p *ppu.PPU, a *apu.APU, mapper Mapper) *Bus {
	return &Bus{ppu: p, sound: a, mapper: mapper}
}

// Read implements cpu.Bus and apu.MemoryReader (the DMC channel streams
// its samples through the same address space the CPU sees).
func (b *Bus) Read(addr uint16) byte {
	switch {
	case addr < 0x2000:
		return b.ram[addr%0x0800]
	case addr < 0x4000:
		return b.ppu.ReadRegister(0x2000 + (addr-0x2000)%8)
	case addr == 0x4015:
		return b.sound.ReadRegister(addr)
	case addr == 0x4016:
		return b.controller1.Read()
	case addr == 0x4017:
		return b.controller2.Read()
	case addr < 0x4018:
		return 0 // remaining APU registers are write-only
	case addr < 0x6000:
		return 0 // expansion area, unused by supported mappers
	case addr < 0x8000:
		return b.prgRAM[addr-0x6000]
	default:
		return b.mapper.ReadPRG(addr)
	}
}

// Write implements cpu.Bus.
func (b *Bus) Write(addr uint16, v byte) {
	switch {
	case addr < 0x2000:
		b.ram[addr%0x0800] = v
	case addr < 0x4000:
		b.ppu.WriteRegister(0x2000+(addr-0x2000)%8, v)
	case addr == 0x4014:
		b.oamDMA(v)
	case addr == 0x4016:
		b.controller1.Write(v)
		b.controller2.Write(v)
	case addr < 0x4018:
		b.sound.WriteRegister(addr, v)
	case addr < 0x6000:
		// expansion area, ignored
	case addr < 0x8000:
		b.prgRAM[addr-0x6000] = v
	default:
		b.mapper.WritePRG(addr, v)
	}
}

// oamDMA copies 256 bytes starting at page*0x100 into OAM, and stalls
// the CPU for the same 513 cycles real hardware does (514 on an odd CPU
// cycle - that extra one-cycle wobble isn't modeled here).
func (b *Bus) oamDMA(page byte) {
	base := uint16(page) << 8
	for i := 0; i < 256; i++ {
		b.ppu.WriteOAMByte(b.Read(base + uint16(i)))
	}
	b.stallCycles += 513
}

// TakeStallCycles returns and clears however many cycles OAM DMA has
// stalled the CPU for since the last call.
func (b *Bus) TakeStallCycles() int {
	c := b.stallCycles
	b.stallCycles = 0
	return c
}

// SetButtons feeds controller 1's held-button state (this project only
// wires up a single player).
func (b *Bus) SetButtons(state input.State) { b.controller1.SetButtons(state) }
