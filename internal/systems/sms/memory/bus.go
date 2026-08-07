package memory

import (
	"github.com/Duc-inc/espaze/internal/emulation/input"
	"github.com/Duc-inc/espaze/internal/systems/sms/psg"
	"github.com/Duc-inc/espaze/internal/systems/sms/vdp"
)

// Bus is the SMS's full memory map ($0000-$FFFF, satisfying cpu.Bus)
// plus its I/O port space (satisfying cpu.IOBus): cartridge ROM/RAM via
// the Sega mapper, 8KB system RAM (mirrored through $E000-$FFFF, with
// the last 4 bytes of that mirror doubling as the mapper's own bank
// registers), and the VDP/PSG/joypad ports.
type Bus struct {
	mapper *mapper
	ram    [0x2000]byte

	video *vdp.VDP
	sound *psg.PSG
	pad   joypad
}

// New wires the bus to the cartridge ROM and the VDP/PSG it forwards
// port access to.
func New(rom []byte, video *vdp.VDP, sound *psg.PSG) *Bus {
	return &Bus{mapper: newMapper(rom), video: video, sound: sound}
}

// SetButtons feeds the joypad's held-button state.
func (b *Bus) SetButtons(state input.State) { b.pad.SetButtons(state) }

// PausePressed reports whether the Pause button (wired directly to the
// Z80's NMI line on real hardware, not through either I/O port) is
// currently held.
func (b *Bus) PausePressed() bool { return b.pad.state.Pressed(Pause) }

// Read implements cpu.Bus.
func (b *Bus) Read(addr uint16) byte {
	switch {
	case addr < 0xC000:
		return b.mapper.ReadROM(addr)
	case addr < 0xE000:
		return b.ram[addr-0xC000]
	default:
		return b.ram[addr-0xE000]
	}
}

// Write implements cpu.Bus.
func (b *Bus) Write(addr uint16, v byte) {
	switch {
	case addr < 0xC000:
		b.mapper.WriteROM(addr, v)
	case addr < 0xE000:
		b.ram[addr-0xC000] = v
	default:
		b.ram[addr-0xE000] = v
		if addr >= 0xFFFC {
			b.mapper.WriteControl(addr, v)
		}
	}
}

// In implements cpu.IOBus.
func (b *Bus) In(port byte) byte {
	switch {
	case port < 0x40:
		return 0xFF
	case port < 0x80:
		if port&1 == 0 {
			return byte(b.video.CurrentLine())
		}
		return 0xFF // horizontal counter: not implemented, no game this project targets needs it
	case port < 0xC0:
		if port&1 == 0 {
			return b.video.ReadData()
		}
		return b.video.ReadStatus()
	default:
		if port&1 == 0 {
			return b.pad.ReadPortDC()
		}
		return b.pad.ReadPortDD()
	}
}

// Out implements cpu.IOBus.
func (b *Bus) Out(port byte, v byte) {
	switch {
	case port < 0x40:
		// memory control / IO control registers - every game this
		// project targets leaves both at their power-on defaults
	case port < 0x80:
		b.sound.Write(v)
	case port < 0xC0:
		if port&1 == 0 {
			b.video.WriteData(v)
		} else {
			b.video.WriteControl(v)
		}
	default:
		// $C0-$FF writes aren't used by any real peripheral this
		// project supports
	}
}
