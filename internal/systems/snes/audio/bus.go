package audio

import "github.com/Duc-inc/espaze/internal/systems/snes/dsp"

// Bus is the SPC700's own 64KB address space: its ARAM, the DSP's
// address/data register ports, and the shared communication ports -
// mirroring this project's own simplified physical layout choices
// elsewhere (see the spc700 package's doc comment for why exact real
// hardware addresses aren't used).
type Bus struct {
	aram  [0x10000]byte
	sound *dsp.DSP
	ports *Ports

	dspAddr byte
}

// New wires an SPC700 bus around a shared DSP and communication ports.
func New(sound *dsp.DSP, ports *Ports) *Bus {
	return &Bus{sound: sound, ports: ports}
}

// Reset clears ARAM.
func (b *Bus) Reset() { b.aram = [0x10000]byte{} }

const (
	portBase = 0xF4
	portEnd  = 0xF8
)

// Read8 implements spc700.Bus.
func (b *Bus) Read8(addr uint16) byte {
	switch {
	case addr == 0xF2:
		return b.dspAddr
	case addr >= portBase && addr < portEnd:
		return b.ports.Read(int(addr - portBase))
	default:
		return b.aram[addr]
	}
}

// Write8 implements spc700.Bus.
func (b *Bus) Write8(addr uint16, v byte) {
	switch {
	case addr == 0xF2:
		b.dspAddr = v
	case addr == 0xF3:
		b.writeDSPRegister(v)
	case addr >= portBase && addr < portEnd:
		b.ports.Write(int(addr-portBase), v)
	default:
		b.aram[addr] = v
	}
}

// writeDSPRegister dispatches by this project's own simplified DSP
// register layout: bits4-6 select the channel, bits0-3 select which
// per-channel register (or, at two reserved global addresses, the
// key-on/key-off bitmasks).
func (b *Bus) writeDSPRegister(v byte) {
	switch b.dspAddr {
	case 0x4C: // KON
		for ch := 0; ch < 8; ch++ {
			if v&(1<<uint(ch)) != 0 {
				b.sound.KeyOn(ch)
			}
		}
	case 0x5C: // KOF
		for ch := 0; ch < 8; ch++ {
			if v&(1<<uint(ch)) != 0 {
				b.sound.KeyOff(ch)
			}
		}
	default:
		ch := int(b.dspAddr>>4) & 0x07
		switch b.dspAddr & 0x0F {
		case 0x00:
			b.sound.WriteVolume(ch, v)
		case 0x02:
			b.sound.WritePitchLow(ch, v)
		case 0x03:
			b.sound.WritePitchHigh(ch, v)
		case 0x04:
			b.sound.WriteWaveByte(ch, v)
		}
	}
}
