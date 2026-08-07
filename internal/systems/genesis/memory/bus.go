// Package memory implements the Genesis's 68000-visible address
// space: cartridge ROM, 64KB work RAM, the VDP's port window, the
// controller, and the Z80 coprocessor's bus-request/reset lines plus a
// pass-through window into its RAM.
package memory

import (
	"github.com/Duc-inc/espaze/internal/emulation/input"
	"github.com/Duc-inc/espaze/internal/systems/genesis/vdp"
)

// Z80Bridge is the subset of the audio coprocessor's bus the 68000
// needs to see: its RAM (for loading driver code and posting
// commands) and its two control lines.
type Z80Bridge interface {
	Read(addr uint16) byte
	Write(addr uint16, v byte)
	RequestBus(held bool)
	BusAcknowledged() bool
}

// Bus implements genesis/cpu.Bus - the 68000's full 24-bit address space.
type Bus struct {
	rom  cartridge
	wram [0x10000]byte
	vdp  *vdp.VDP
	pad1 Controller
	z80  Z80Bridge

	z80Reset bool
}

// New builds a Bus around a loaded ROM and the VDP/Z80 it shares the
// address space with.
func New(rom []byte, video *vdp.VDP, z80 Z80Bridge) *Bus {
	return &Bus{rom: newCartridge(rom), vdp: video, z80: z80}
}

// SetButtons applies the latest input state to the port-A controller.
func (b *Bus) SetButtons(state input.State) { b.pad1.SetButtons(state) }

func (b *Bus) inZ80Window(addr uint32) bool  { return addr >= 0xA00000 && addr <= 0xA0FFFF }
func (b *Bus) inIOWindow(addr uint32) bool   { return addr >= 0xA10000 && addr <= 0xA1001F }
func (b *Bus) inVDPWindow(addr uint32) bool  { return addr >= 0xC00000 && addr <= 0xC0001F }
func (b *Bus) inCartWindow(addr uint32) bool { return addr < 0x400000 }
func (b *Bus) inWRAMWindow(addr uint32) bool { return addr >= 0xE00000 }

// Read8 implements genesis/cpu.Bus.
func (b *Bus) Read8(addr uint32) byte {
	addr &= 0xFFFFFF
	switch {
	case b.inCartWindow(addr):
		return b.rom.read8(addr)
	case b.inZ80Window(addr):
		return b.z80.Read(uint16(addr))
	case b.inIOWindow(addr):
		return b.readIO(addr)
	case addr == 0xA11100 || addr == 0xA11101:
		if b.z80.BusAcknowledged() {
			return 0x00
		}
		return 0x01
	case b.inVDPWindow(addr):
		return byte(b.vdpRead16(addr) >> ((1 - addr&1) * 8))
	case b.inWRAMWindow(addr):
		return b.wram[addr&0xFFFF]
	default:
		return 0xFF
	}
}

// Read16 implements genesis/cpu.Bus.
func (b *Bus) Read16(addr uint32) uint16 {
	addr &= 0xFFFFFE
	switch {
	case b.inCartWindow(addr):
		return b.rom.read16(addr)
	case b.inVDPWindow(addr):
		return b.vdpRead16(addr)
	case b.inWRAMWindow(addr):
		i := addr & 0xFFFF
		return uint16(b.wram[i])<<8 | uint16(b.wram[(i+1)&0xFFFF])
	default:
		return uint16(b.Read8(addr))<<8 | uint16(b.Read8(addr+1))
	}
}

// Read32 implements genesis/cpu.Bus.
func (b *Bus) Read32(addr uint32) uint32 {
	return uint32(b.Read16(addr))<<16 | uint32(b.Read16(addr+2))
}

func (b *Bus) vdpRead16(addr uint32) uint16 {
	switch addr & 0x1F {
	case 0, 1, 2, 3:
		return b.vdp.ReadData()
	case 4, 5, 6, 7:
		return b.vdp.ReadStatus()
	default:
		return 0 // H/V counter isn't modeled
	}
}

func (b *Bus) readIO(addr uint32) byte {
	switch addr & 0x1F {
	case 0x01:
		return 0xA0 // hardware version register: no TMSS, domestic NTSC
	case 0x03:
		return b.pad1.Read()
	default:
		return 0xFF // port B/EXT aren't implemented
	}
}

// Write8 implements genesis/cpu.Bus.
func (b *Bus) Write8(addr uint32, v byte) {
	addr &= 0xFFFFFF
	switch {
	case b.inCartWindow(addr):
		// ROM: writes dropped, no SRAM modeled.
	case b.inZ80Window(addr):
		b.z80.Write(uint16(addr), v)
	case b.inIOWindow(addr):
		b.writeIO(addr, v)
	case addr == 0xA11100 || addr == 0xA11101:
		b.z80.RequestBus(v&0x01 != 0)
	case addr == 0xA11200 || addr == 0xA11201:
		b.z80Reset = v&0x01 == 0
	case b.inVDPWindow(addr):
		b.vdpWrite16(addr, uint16(v)<<8|uint16(v))
	case b.inWRAMWindow(addr):
		b.wram[addr&0xFFFF] = v
	}
}

// Write16 implements genesis/cpu.Bus.
func (b *Bus) Write16(addr uint32, v uint16) {
	addr &= 0xFFFFFE
	switch {
	case b.inCartWindow(addr):
	case b.inVDPWindow(addr):
		b.vdpWrite16(addr, v)
	case b.inWRAMWindow(addr):
		i := addr & 0xFFFF
		b.wram[i] = byte(v >> 8)
		b.wram[(i+1)&0xFFFF] = byte(v)
	default:
		b.Write8(addr, byte(v>>8))
		b.Write8(addr+1, byte(v))
	}
}

// Write32 implements genesis/cpu.Bus.
func (b *Bus) Write32(addr uint32, v uint32) {
	b.Write16(addr, uint16(v>>16))
	b.Write16(addr+2, uint16(v))
}

func (b *Bus) vdpWrite16(addr uint32, v uint16) {
	switch addr & 0x1F {
	case 0, 1, 2, 3:
		b.vdp.WriteData(v)
	case 4, 5, 6, 7:
		b.vdp.WriteControl(v)
	}
}

func (b *Bus) writeIO(addr uint32, v byte) {
	if addr&0x1F == 0x03 {
		b.pad1.Write(v)
	}
}

// Z80ResetAsserted reports whether the 68000 currently holds the Z80
// in reset, so the top-level core can skip stepping it.
func (b *Bus) Z80ResetAsserted() bool { return b.z80Reset }
