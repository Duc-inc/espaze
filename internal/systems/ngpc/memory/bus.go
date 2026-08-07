// Package memory implements the Neo Geo Pocket (Color)'s address
// space as seen by its TLCS900H: cartridge ROM, work RAM, the PPU's
// VRAM/sprite table/palette windows, the Z80 sound coprocessor's
// shared RAM window, and the controller/Z80-control I/O registers.
// This project's own address layout, like its CPU and PPU packages,
// is a self-consistent simplification rather than a verified
// reproduction of real hardware's physical memory map.
package memory

import "github.com/Duc-inc/espaze/internal/emulation/input"

// PPU is the subset of ppu.PPU this bus dispatches register access to.
type PPU interface {
	WriteControl(v byte)
	WriteScrollX(v byte)
	WriteScrollY(v byte)
	ReadVRAM(addr uint32) byte
	WriteVRAM(addr uint32, v byte)
	ReadSprite(addr uint32) byte
	WriteSprite(addr uint32, v byte)
	WritePaletteLow(index byte, v byte)
	WritePaletteHigh(index byte, v byte)
}

// Z80Bridge is the subset of the audio coprocessor's bus the main CPU
// can see: its shared RAM window and a reset control line.
type Z80Bridge interface {
	Read(addr uint16) byte
	Write(addr uint16, v byte)
}

const (
	wramBase    = 0x200000
	vramBase    = 0x204000
	spriteBase  = 0x208000
	paletteBase = 0x209000
	z80Base     = 0x20A000
	ioBase      = 0x20B000
)

// Bus implements ngpc/cpu.Bus.
type Bus struct {
	rom cartridge

	wram [0x4000]byte

	video    PPU
	z80      Z80Bridge
	pad      Controller
	z80Reset bool
}

// New wires the bus to a loaded ROM and the PPU/Z80 bridge it
// dispatches register access to.
func New(rom []byte, video PPU, z80 Z80Bridge) *Bus {
	return &Bus{rom: newCartridge(rom), video: video, z80: z80}
}

// SetButtons applies the latest input state to the controller.
func (b *Bus) SetButtons(state input.State) { b.pad.SetButtons(state) }

// Z80ResetAsserted reports whether the main CPU currently holds the
// Z80 coprocessor in reset.
func (b *Bus) Z80ResetAsserted() bool { return b.z80Reset }

// Read8 implements cpu.Bus.
func (b *Bus) Read8(addr uint32) byte {
	switch {
	case addr < 0x200000:
		return b.rom.read8(addr)
	case addr >= wramBase && addr < wramBase+0x4000:
		return b.wram[addr-wramBase]
	case addr >= vramBase && addr < vramBase+0x4000:
		return b.video.ReadVRAM(addr - vramBase)
	case addr >= spriteBase && addr < spriteBase+0x100:
		return b.video.ReadSprite(addr - spriteBase)
	case addr >= z80Base && addr < z80Base+0x100:
		return b.z80.Read(uint16(addr - z80Base))
	case addr == ioBase:
		return b.pad.Read()
	default:
		return 0xFF
	}
}

// Read16 implements cpu.Bus.
func (b *Bus) Read16(addr uint32) uint16 {
	if addr < 0x200000 {
		return b.rom.read16(addr)
	}
	return uint16(b.Read8(addr)) | uint16(b.Read8(addr+1))<<8
}

// Write8 implements cpu.Bus.
func (b *Bus) Write8(addr uint32, v byte) {
	switch {
	case addr < 0x200000:
		// ROM: writes dropped.
	case addr >= wramBase && addr < wramBase+0x4000:
		b.wram[addr-wramBase] = v
	case addr >= vramBase && addr < vramBase+0x4000:
		b.video.WriteVRAM(addr-vramBase, v)
	case addr >= spriteBase && addr < spriteBase+0x100:
		b.video.WriteSprite(addr-spriteBase, v)
	case addr >= paletteBase && addr < paletteBase+0x40:
		idx := byte((addr - paletteBase) / 2)
		if (addr-paletteBase)&1 == 0 {
			b.video.WritePaletteLow(idx, v)
		} else {
			b.video.WritePaletteHigh(idx, v)
		}
	case addr >= z80Base && addr < z80Base+0x100:
		b.z80.Write(uint16(addr-z80Base), v)
	case addr == ioBase:
		// input port is read-only
	case addr == ioBase+1:
		b.video.WriteControl(v)
	case addr == ioBase+2:
		b.video.WriteScrollX(v)
	case addr == ioBase+3:
		b.video.WriteScrollY(v)
	case addr == ioBase+4:
		b.z80Reset = v&0x01 == 0
	}
}

// Write16 implements cpu.Bus.
func (b *Bus) Write16(addr uint32, v uint16) {
	b.Write8(addr, byte(v))
	b.Write8(addr+1, byte(v>>8))
}
