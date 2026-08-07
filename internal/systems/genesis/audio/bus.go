// Package audio wires the Genesis's sound coprocessor: the same Z80
// chip this project already implements from scratch for the Master
// System (internal/systems/sms/cpu), repurposed here to drive the
// YM2612 FM synth and the backward-compatible SN76489 PSG (reused
// directly from internal/systems/sms/psg - it's genuinely the same
// chip on real Genesis hardware).
package audio

import (
	"github.com/Duc-inc/espaze/internal/systems/genesis/ym2612"
	"github.com/Duc-inc/espaze/internal/systems/sms/psg"
)

// Bus is the Z80 coprocessor's 16-bit memory map: its own small RAM,
// the YM2612's register ports, the PSG, and (on real hardware) a
// bank-switched window into the 68000's ROM space. That last window
// isn't implemented - no driver this project has tested needs the Z80
// to read cartridge data, only to receive commands the 68k already
// wrote into shared RAM - so reads there return open-bus 0xFF and
// writes are dropped.
type Bus struct {
	ram    [0x2000]byte // 8KB, mirrored across $0000-$3FFF
	ym     *ym2612.YM2612
	sound  *psg.PSG
	bank   uint16
	halted bool // set while the 68k holds the bus request line
}

// New wires a Z80 bus around a shared YM2612 and PSG - the same chip
// instances the top-level Genesis core drains samples from.
func New(ym *ym2612.YM2612, sound *psg.PSG) *Bus {
	return &Bus{ym: ym, sound: sound}
}

// Reset clears the coprocessor's own RAM; the YM2612/PSG are reset
// independently by their owner.
func (b *Bus) Reset() { b.ram = [0x2000]byte{} }

// Read implements sms/cpu.Bus.
func (b *Bus) Read(addr uint16) byte {
	switch {
	case addr < 0x4000:
		return b.ram[addr&0x1FFF]
	case addr < 0x6000:
		return 0 // YM2612 status isn't modeled; always report "not busy"
	default:
		return 0xFF
	}
}

// Write implements sms/cpu.Bus.
func (b *Bus) Write(addr uint16, v byte) {
	switch {
	case addr < 0x4000:
		b.ram[addr&0x1FFF] = v
	case addr < 0x6000:
		b.writeYM2612(addr, v)
	case addr == 0x6000:
		b.bank = b.bank>>1 | uint16(v&1)<<8 // 9-bit bank register, shifted in one bit per write
	case addr >= 0x7F00 && addr < 0x8000:
		if addr&1 == 1 { // $7F11 (and its mirrors) is the PSG's single write port
			b.sound.Write(v)
		}
	}
}

func (b *Bus) writeYM2612(addr uint16, v byte) {
	switch addr & 0x03 {
	case 0:
		b.ym.WriteAddress1(v)
	case 1:
		b.ym.WriteData1(v)
	case 2:
		b.ym.WriteAddress2(v)
	case 3:
		b.ym.WriteData2(v)
	}
}

// In/Out implement sms/cpu.IOBus: the Genesis's Z80 coprocessor has no
// devices in I/O space, everything above is memory-mapped, so these
// are stubs the CPU's IN/OUT instructions never meaningfully hit.
func (b *Bus) In(port byte) byte     { return 0xFF }
func (b *Bus) Out(port byte, v byte) {}

// RAM exposes the coprocessor's shared RAM directly - the 68000 can
// read and write it through its own $A00000 window (see
// genesis/memory.Bus), which is how games actually load Z80 driver
// code and post commands to it.
func (b *Bus) RAM() *[0x2000]byte { return &b.ram }

// RequestBus/BusAcknowledged/SetReset model the two coprocessor
// control lines the 68000 can see at $A11100/$A11200. Holding the bus
// request just pauses this CPU's own Step calls (see genesis.go); real
// hardware's more nuanced arbitration isn't reproduced.
func (b *Bus) RequestBus(held bool)  { b.halted = held }
func (b *Bus) BusAcknowledged() bool { return b.halted }
func (b *Bus) Halted() bool          { return b.halted }
