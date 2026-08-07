// Package audio wires the Neo Geo Pocket (Color)'s sound coprocessor:
// a Z80 (reused directly from internal/systems/sms/cpu - the same
// chip, present on real NGPC hardware specifically for sound driver
// backward-compatibility) driving a T6W28 PSG, which is close enough
// to the SN76489 this project already implements for internal/systems/sms/psg
// that this project reuses it directly rather than modeling the T6W28's
// small differences (independent per-channel stereo panning being the
// main one - not reproduced here).
package audio

import "github.com/Duc-inc/espaze/internal/systems/sms/psg"

// Bus is the Z80 coprocessor's 16-bit memory map: its own RAM plus
// the PSG's write port.
type Bus struct {
	ram   [0x4000]byte
	sound *psg.PSG
}

// New wires a Z80 bus around a shared PSG instance.
func New(sound *psg.PSG) *Bus { return &Bus{sound: sound} }

// Reset clears the coprocessor's own RAM.
func (b *Bus) Reset() { b.ram = [0x4000]byte{} }

// Read implements sms/cpu.Bus.
func (b *Bus) Read(addr uint16) byte { return b.ram[addr&0x3FFF] }

// Write implements sms/cpu.Bus.
func (b *Bus) Write(addr uint16, v byte) {
	if addr >= 0x4000 {
		return
	}
	b.ram[addr&0x3FFF] = v
}

// In/Out implement sms/cpu.IOBus: the PSG's single write port lives
// at $00 in this project's I/O space.
func (b *Bus) In(port byte) byte { return 0xFF }
func (b *Bus) Out(port byte, v byte) {
	if port == 0x00 {
		b.sound.Write(v)
	}
}
