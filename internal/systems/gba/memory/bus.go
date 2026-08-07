// Package memory implements the Game Boy Advance's 32-bit address
// space: cartridge ROM, the two work-RAM regions (256KB "EWRAM" +
// 32KB "IWRAM"), a simple SRAM save region, and the PPU/APU/DMA/timer/
// keypad/interrupt I/O registers - decoded by address region exactly
// like real hardware's coarse top-byte-based memory map.
package memory

import "github.com/Duc-inc/espaze/internal/emulation/input"

// PPU is the subset of ppu.PPU this bus dispatches register access to.
type PPU interface {
	WriteDISPCNT(v uint16)
	ReadDISPCNT() uint16
	WriteDISPSTAT(v uint16)
	WriteBGCNT(bg int, v uint16)
	WriteBGHOFS(bg int, v uint16)
	WriteBGVOFS(bg int, v uint16)
	ReadVRAM8(addr uint32) byte
	WriteVRAM8(addr uint32, v byte)
	ReadOAM8(addr uint32) byte
	WriteOAM8(addr uint32, v byte)
	ReadPalette8(addr uint32) byte
	WritePalette16(addr uint32, v uint16)
}

// APU is the subset of apu.APU this bus dispatches register access to.
type APU interface {
	WriteFIFOA(v byte)
	WriteFIFOB(v byte)
	WriteSoundCntH(v uint16)
}

// Bus implements gba/cpu.Bus - the ARM7TDMI's full 32-bit address space.
type Bus struct {
	rom   cartridge
	ewram [0x40000]byte
	iwram [0x8000]byte
	sram  sram

	video PPU
	sound APU
	dma   [4]dmaChannel
	tm    timers
	kp    keypad
	irq   interrupts
}

// New wires the bus to a loaded ROM and the PPU/APU it dispatches
// register access to.
func New(rom []byte, video PPU, sound APU) *Bus {
	b := &Bus{rom: newCartridge(rom), video: video, sound: sound}
	b.tm.irq = &b.irq
	return b
}

// SetButtons applies the latest input state to the keypad.
func (b *Bus) SetButtons(state input.State) { b.kp.setButtons(state) }

// RaiseVBlank lets the top-level core report the PPU's VBlank IRQ
// (the PPU itself doesn't know about the interrupt controller).
func (b *Bus) RaiseVBlank() { b.irq.raise(IFVBlank) }

// InterruptPending reports whether the CPU's IRQ line should be asserted.
func (b *Bus) InterruptPending() bool { return b.irq.pending() }

// StepTimers advances every timer by cpuCycles CPU cycles.
func (b *Bus) StepTimers(cpuCycles int) { b.tm.step(cpuCycles) }

func region(addr uint32) uint32 { return (addr >> 24) & 0xFF }

// Read8 implements cpu.Bus.
func (b *Bus) Read8(addr uint32) byte {
	switch region(addr) {
	case 0x02:
		return b.ewram[addr&0x3FFFF]
	case 0x03:
		return b.iwram[addr&0x7FFF]
	case 0x04:
		return byte(b.readIO(addr&^1) >> ((addr & 1) * 8))
	case 0x05:
		return b.video.ReadPalette8(addr)
	case 0x06:
		return b.video.ReadVRAM8(vramAddr(addr))
	case 0x07:
		return b.video.ReadOAM8(addr)
	case 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D:
		return b.rom.read8(addr & 0x01FFFFFF)
	case 0x0E:
		return b.sram.read(addr)
	default:
		return 0
	}
}

// Read16 implements cpu.Bus.
func (b *Bus) Read16(addr uint32) uint16 {
	addr &^= 1
	switch region(addr) {
	case 0x04:
		return b.readIO(addr)
	case 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D:
		return b.rom.read16(addr & 0x01FFFFFF)
	default:
		return uint16(b.Read8(addr)) | uint16(b.Read8(addr+1))<<8
	}
}

// Read32 implements cpu.Bus.
func (b *Bus) Read32(addr uint32) uint32 {
	addr &^= 3
	if region(addr) >= 0x08 && region(addr) <= 0x0D {
		return b.rom.read32(addr & 0x01FFFFFF)
	}
	return uint32(b.Read16(addr)) | uint32(b.Read16(addr+2))<<16
}

// Write8 implements cpu.Bus.
func (b *Bus) Write8(addr uint32, v byte) {
	switch region(addr) {
	case 0x02:
		b.ewram[addr&0x3FFFF] = v
	case 0x03:
		b.iwram[addr&0x7FFF] = v
	case 0x04:
		b.writeIO8(addr, v)
	case 0x05:
		word := uint16(v)<<8 | uint16(v)
		b.video.WritePalette16(addr&^1, word)
	case 0x06:
		b.video.WriteVRAM8(vramAddr(addr), v)
	case 0x07:
		b.video.WriteOAM8(addr, v)
	case 0x0E:
		b.sram.write(addr, v)
	}
}

// Write16 implements cpu.Bus.
func (b *Bus) Write16(addr uint32, v uint16) {
	addr &^= 1
	switch region(addr) {
	case 0x02:
		b.ewram[addr&0x3FFFF], b.ewram[(addr+1)&0x3FFFF] = byte(v), byte(v>>8)
	case 0x03:
		b.iwram[addr&0x7FFF], b.iwram[(addr+1)&0x7FFF] = byte(v), byte(v>>8)
	case 0x04:
		b.writeIO(addr, v)
	case 0x05:
		b.video.WritePalette16(addr, v)
	case 0x06:
		va := vramAddr(addr)
		b.video.WriteVRAM8(va, byte(v))
		b.video.WriteVRAM8(va+1, byte(v>>8))
	case 0x07:
		b.video.WriteOAM8(addr, byte(v))
		b.video.WriteOAM8(addr+1, byte(v>>8))
	}
}

// Write32 implements cpu.Bus.
func (b *Bus) Write32(addr uint32, v uint32) {
	addr &^= 3
	b.Write16(addr, uint16(v))
	b.Write16(addr+2, uint16(v>>16))
}

// vramAddr mirrors the 96KB VRAM region's own quirky over-size mirror
// (128KB of address space maps 96KB of real VRAM, repeating the last
// 32KB) - simplified here to a flat mask across the 96KB actually backed.
func vramAddr(addr uint32) uint32 {
	a := addr & 0x1FFFF
	if a >= 0x18000 {
		a -= 0x8000
	}
	return a
}
