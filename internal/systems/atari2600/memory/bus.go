// Package memory implements the Atari 2600's address space as seen by
// its 6507 CPU (a 6502 variant with only 13 address lines bonded out,
// A0-A12): the TIA's registers, the RIOT's RAM/timer/ports, and
// cartridge ROM, decoded the same coarse way real hardware does (only
// a few address bits are actually checked, so each device mirrors
// across large chunks of the space - this reproduces the standard,
// well-documented mirroring every 2600 game relies on).
package memory

import (
	"github.com/Duc-inc/espaze/internal/emulation/input"
	"github.com/Duc-inc/espaze/internal/systems/atari2600/riot"
	"github.com/Duc-inc/espaze/internal/systems/atari2600/tia"
)

// Bus implements the 6507's Bus interface (Read/Write(uint16)),
// satisfying internal/systems/nes/cpu.Bus structurally - this project
// reuses that 6502 core directly for the 2600's 6507.
type Bus struct {
	rom   cartridge
	video *tia.TIA
	riot  *riot.RIOT
}

// New wires the bus to a loaded ROM and the TIA/RIOT it forwards
// register access to.
func New(rom []byte, video *tia.TIA, r *riot.RIOT) *Bus {
	return &Bus{rom: newCartridge(rom), video: video, riot: r}
}

// SetButtons feeds the joystick's held-button state to both chips: the
// RIOT for the d-pad, the TIA for the fire button (wired to its own
// INPT4 port on real hardware, not the RIOT).
func (b *Bus) SetButtons(state input.State) {
	b.riot.SetButtons(state)
	b.video.SetButton(0, state.Pressed(riot.Fire))
}

// Read implements cpu.Bus.
func (b *Bus) Read(addr uint16) byte {
	a := addr & 0x1FFF
	switch {
	case a&0x1000 != 0:
		return b.rom.read(a & 0x0FFF)
	case a&0x0080 == 0:
		return b.video.ReadRegister(byte(a & 0x0F))
	case a&0x0200 == 0:
		return b.riot.ReadRAM(byte(a))
	default:
		return b.readRIOTPort(byte(a & 0x1F))
	}
}

func (b *Bus) readRIOTPort(offset byte) byte {
	switch offset {
	case 0x00:
		return b.riot.ReadSWCHA()
	case 0x02:
		return b.riot.ReadSWCHB()
	case 0x04:
		return b.riot.ReadINTIM()
	default:
		return 0xFF
	}
}

// Write implements cpu.Bus.
func (b *Bus) Write(addr uint16, v byte) {
	a := addr & 0x1FFF
	switch {
	case a&0x1000 != 0:
		// ROM: writes dropped.
	case a&0x0080 == 0:
		b.video.WriteRegister(byte(a&0x3F), v)
	case a&0x0200 == 0:
		b.riot.WriteRAM(byte(a), v)
	default:
		b.writeRIOTPort(byte(a&0x1F), v)
	}
}

func (b *Bus) writeRIOTPort(offset byte, v byte) {
	if offset&0x14 == 0x14 { // TIM1T/TIM8T/TIM64T/T1024T
		b.riot.WriteTimer(offset, v)
	}
	// SWACNT/SWCHA/SWBCNT writes (port direction/output) aren't
	// modeled - every game this project targets only reads the ports.
}
