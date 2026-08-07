// Package memory implements the ColecoVision's memory map as seen by
// its Z80 (reused directly from internal/systems/sms/cpu - the same
// chip): cartridge ROM, 1KB RAM, the TMS9918 VDP's port window, the
// PSG's write port, and the controller. This project has no BIOS
// (none is redistributable), so cartridge code runs directly from
// power-on rather than being handed off to by the real 8KB boot ROM.
package memory

import (
	"github.com/Duc-inc/espaze/internal/emulation/input"
	"github.com/Duc-inc/espaze/internal/systems/sms/psg"
	"github.com/Duc-inc/espaze/internal/systems/tms9918"
)

// Bus implements sms/cpu.Bus (memory) and sms/cpu.IOBus (ports).
type Bus struct {
	rom cartridge
	ram [0x0400]byte
	pad Controller

	video *tms9918.TMS9918
	sound *psg.PSG
}

// New wires the bus to a loaded ROM and the VDP/PSG it forwards
// access to.
func New(rom []byte, video *tms9918.TMS9918, sound *psg.PSG) *Bus {
	return &Bus{rom: newCartridge(rom), video: video, sound: sound}
}

// SetButtons applies the latest input state to the controller.
func (b *Bus) SetButtons(state input.State) { b.pad.SetButtons(state) }

// Read implements cpu.Bus.
func (b *Bus) Read(addr uint16) byte {
	switch {
	case addr < 0x6000:
		return b.rom.read(addr)
	case addr < 0x8000:
		return b.ram[addr&0x03FF]
	default:
		return b.rom.read(addr)
	}
}

// Write implements cpu.Bus. ROM (including the $6000-$7FFF alias some
// carts also decode as RAM-adjacent space) only actually writes into
// RAM; the $8000+ cartridge window is read-only.
func (b *Bus) Write(addr uint16, v byte) {
	if addr >= 0x6000 && addr < 0x8000 {
		b.ram[addr&0x03FF] = v
	}
}

// In implements cpu.IOBus.
func (b *Bus) In(port byte) byte {
	switch {
	case port&0xE0 == 0xA0:
		if port&1 == 0 {
			return b.video.ReadData()
		}
		return b.video.ReadStatus()
	case port&0xE0 == 0xE0:
		return b.pad.Read()
	default:
		return 0xFF
	}
}

// Out implements cpu.IOBus.
func (b *Bus) Out(port byte, v byte) {
	switch {
	case port&0xE0 == 0xA0:
		if port&1 == 0 {
			b.video.WriteData(v)
		} else {
			b.video.WriteControl(v)
		}
	case port&0xE0 == 0xE0:
		b.sound.Write(v)
	}
}
