// Package memory implements the SNES's 24-bit address space as seen
// by its 65816: cartridge ROM (a simplified flat LoROM-style mapping,
// not the real chip's bank-switching variety), 128KB of work RAM, the
// PPU's register window, the shared audio communication ports, and
// the controller. This project's own physical layout, like its CPU
// and PPU packages, is a deliberate simplification rather than a
// reproduction of real hardware's exact memory map.
package memory

import "github.com/Duc-inc/espaze/internal/emulation/input"

// PPU is the subset of ppu.PPU this bus dispatches register access to.
type PPU interface {
	WriteVRAMAddrLow(v byte)
	WriteVRAMAddrHigh(v byte)
	WriteVRAMDataLow(v byte)
	WriteVRAMDataHigh(v byte)
	ReadVRAMLow() byte
	ReadVRAMHigh() byte
	WriteCGRAMAddr(v byte)
	WriteCGRAMData(v byte)
	WriteOAMByte(addr uint16, v byte)
	ReadOAMByte(addr uint16) byte
	WriteBGControl(layer int, v byte)
	WriteBGScrollH(layer int, v uint16)
	WriteBGScrollV(layer int, v uint16)
	WriteMainScreen(v byte)
}

// AudioPorts is the subset of audio.Ports the main CPU can see.
type AudioPorts interface {
	Read(i int) byte
	Write(i int, v byte)
}

// Bus implements snes/cpu.Bus.
type Bus struct {
	rom  cartridge
	wram [0x20000]byte

	video PPU
	ports AudioPorts
	pad1  Controller

	oamAddr uint16
}

// New wires the bus to a loaded ROM and the PPU/audio ports it
// dispatches register access to.
func New(rom []byte, video PPU, ports AudioPorts) *Bus {
	return &Bus{rom: newCartridge(rom), video: video, ports: ports}
}

// SetButtons applies the latest input state to the controller.
func (b *Bus) SetButtons(state input.State) { b.pad1.SetButtons(state) }

func bank(addr uint32) uint32   { return (addr >> 16) & 0xFF }
func offset(addr uint32) uint32 { return addr & 0xFFFF }

// Read8 implements cpu.Bus.
func (b *Bus) Read8(addr uint32) byte {
	bk, off := bank(addr)&0x7F, offset(addr)
	switch {
	case bk == 0x7E || bk == 0x7F:
		return b.wram[(bk-0x7E)<<16|off]
	case off < 0x2000:
		return b.wram[off]
	case off >= 0x2134 && off < 0x2136:
		if off == 0x2134 {
			return b.video.ReadVRAMLow()
		}
		return b.video.ReadVRAMHigh()
	case off >= 0x2140 && off < 0x2144:
		return b.ports.Read(int(off - 0x2140))
	case off == 0x4218:
		return b.pad1.ReadLow()
	case off == 0x4219:
		return b.pad1.ReadHigh()
	case off >= 0x8000:
		return b.rom.read8(bk*0x8000 + (off - 0x8000))
	default:
		return 0
	}
}

// Read16 composes two Read8 calls (65816 real hardware's own 8-bit
// data bus does the same).
func (b *Bus) Read16(addr uint32) uint16 {
	return uint16(b.Read8(addr)) | uint16(b.Read8(addr+1))<<8
}

// Write8 implements cpu.Bus.
func (b *Bus) Write8(addr uint32, v byte) {
	bk, off := bank(addr)&0x7F, offset(addr)
	switch {
	case bk == 0x7E || bk == 0x7F:
		b.wram[(bk-0x7E)<<16|off] = v
	case off < 0x2000:
		b.wram[off] = v
	case off == 0x2116:
		b.video.WriteVRAMAddrLow(v)
	case off == 0x2117:
		b.video.WriteVRAMAddrHigh(v)
	case off == 0x2118:
		b.video.WriteVRAMDataLow(v)
	case off == 0x2119:
		b.video.WriteVRAMDataHigh(v)
	case off == 0x2121:
		b.video.WriteCGRAMAddr(v)
	case off == 0x2122:
		b.video.WriteCGRAMData(v)
	case off == 0x2102:
		b.oamAddr = b.oamAddr&0xFF00 | uint16(v)
	case off == 0x2103:
		b.oamAddr = b.oamAddr&0x00FF | uint16(v)<<8
	case off == 0x2104:
		b.video.WriteOAMByte(b.oamAddr, v)
		b.oamAddr++
	case off >= 0x2107 && off < 0x210B:
		b.video.WriteBGControl(int(off-0x2107), v)
	case off >= 0x210D && off < 0x2115 && (off-0x210D)%2 == 0:
		layer := int(off-0x210D) / 2
		b.video.WriteBGScrollH(layer, uint16(v))
	case off >= 0x210D && off < 0x2115:
		layer := int(off-0x210E) / 2
		b.video.WriteBGScrollV(layer, uint16(v))
	case off == 0x212C:
		b.video.WriteMainScreen(v)
	case off >= 0x2140 && off < 0x2144:
		b.ports.Write(int(off-0x2140), v)
	}
}

// Write16 decomposes into two Write8 calls.
func (b *Bus) Write16(addr uint32, v uint16) {
	b.Write8(addr, byte(v))
	b.Write8(addr+1, byte(v>>8))
}
