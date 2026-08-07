// Package memory implements the PC Engine's 21-bit physical address
// space the HuC6280's MMU translates logical addresses into:
// cartridge ROM, 8KB work RAM, and the VDC/VCE/PSG/timer/IRQ/joypad
// registers. This project's own choice of exactly where each device's
// registers sit isn't independently verified against real hardware's
// physical map - see the cpu package's own doc comment for the same
// caveat about opcode assignments.
package memory

import "github.com/Duc-inc/espaze/internal/emulation/input"

// VDC/VCE/PSG are the minimal interfaces this bus dispatches register
// access to - defined here rather than imported concretely so this
// package doesn't need to depend on their exact types.
type VDC interface {
	SelectRegister(reg byte)
	WriteDataLow(b byte)
	WriteDataHigh(b byte)
	ReadDataLow() byte
	ReadDataHigh() byte
}

type VCE interface {
	WriteAddressLow(b byte)
	WriteAddressHigh(b byte)
	WriteDataLow(b byte)
	WriteDataHigh(b byte)
}

type PSG interface {
	SelectChannel(v byte)
	WriteMainVolumeLeft(v byte)
	WriteMainVolumeRight(v byte)
	WriteFreqLow(v byte)
	WriteFreqHigh(v byte)
	WriteControl(v byte)
	WritePan(v byte)
	WriteWaveData(v byte)
	WriteNoiseControl(v byte)
}

// TimerIRQ is the HuC6280's own built-in timer/interrupt controller -
// see cpu.CPU's exported wrapper methods.
type TimerIRQ interface {
	WriteTimerReload(v byte)
	WriteTimerControl(v byte)
	WriteIRQMask(v byte)
	ReadIRQStatus() byte
}

// Bus implements cpu.Bus (Read/Write(uint32)) for the physical address space.
type Bus struct {
	rom  cartridge
	ram  [0x2000]byte
	vdc  VDC
	vce  VCE
	psg  PSG
	tirq TimerIRQ
	pad  Controller
}

// New wires the bus to a loaded ROM and every hardware device it
// dispatches register access to. tirq is the CPU's own built-in timer
// /interrupt controller - since building the CPU itself requires a
// Bus, callers construct one with SetTimerIRQ called right after, before
// any CPU cycles run.
func New(rom []byte, vdc VDC, vce VCE, sound PSG, tirq TimerIRQ) *Bus {
	return &Bus{rom: newCartridge(rom), vdc: vdc, vce: vce, psg: sound, tirq: tirq}
}

// SetTimerIRQ wires the CPU's built-in timer/interrupt controller in
// after construction, breaking the CPU<->Bus construction cycle.
func (b *Bus) SetTimerIRQ(t TimerIRQ) { b.tirq = t }

// SetButtons applies the latest input state to the pad.
func (b *Bus) SetButtons(state input.State) { b.pad.SetButtons(state) }

// Read implements cpu.Bus.
func (b *Bus) Read(addr uint32) byte {
	switch {
	case addr < 0x100000:
		return b.rom.read(addr)
	case addr >= 0x1F0000 && addr < 0x1F2000:
		return b.ram[addr-0x1F0000]
	case addr >= 0x1FE000 && addr < 0x1FE400:
		return b.readVDC(addr & 0x0F)
	case addr >= 0x1FEC00 && addr < 0x1FED00:
		if addr&0x0F == 0 {
			return b.tirq.ReadIRQStatus() // simplification: INTIM readback isn't exposed, reuse the status port slot
		}
		return 0xFF
	case addr >= 0x1FF000 && addr < 0x1FF100:
		return b.pad.Read()
	case addr >= 0x1FF400 && addr < 0x1FF500:
		if addr&0x0F == 3 {
			return b.tirq.ReadIRQStatus()
		}
		return 0xFF
	default:
		return 0xFF
	}
}

func (b *Bus) readVDC(offset uint32) byte {
	switch offset {
	case 2:
		return b.vdc.ReadDataLow()
	case 3:
		return b.vdc.ReadDataHigh()
	default:
		return 0xFF
	}
}

// Write implements cpu.Bus.
func (b *Bus) Write(addr uint32, v byte) {
	switch {
	case addr < 0x100000:
		// ROM: writes dropped.
	case addr >= 0x1F0000 && addr < 0x1F2000:
		b.ram[addr-0x1F0000] = v
	case addr >= 0x1FE000 && addr < 0x1FE400:
		b.writeVDC(addr&0x0F, v)
	case addr >= 0x1FE400 && addr < 0x1FE800:
		b.writeVCE(addr&0x0F, v)
	case addr >= 0x1FE800 && addr < 0x1FEC00:
		b.writePSG(addr&0x0F, v)
	case addr >= 0x1FEC00 && addr < 0x1FED00:
		b.writeTimer(addr&0x0F, v)
	case addr >= 0x1FF000 && addr < 0x1FF100:
		b.pad.Write(v)
	case addr >= 0x1FF400 && addr < 0x1FF500:
		if addr&0x0F == 2 {
			b.tirq.WriteIRQMask(v)
		}
	}
}

func (b *Bus) writeVDC(offset uint32, v byte) {
	switch offset {
	case 0:
		b.vdc.SelectRegister(v)
	case 2:
		b.vdc.WriteDataLow(v)
	case 3:
		b.vdc.WriteDataHigh(v)
	}
}

func (b *Bus) writeVCE(offset uint32, v byte) {
	switch offset {
	case 0:
		b.vce.WriteAddressLow(v)
	case 2:
		b.vce.WriteAddressHigh(v)
	case 4:
		b.vce.WriteDataLow(v)
	case 6:
		b.vce.WriteDataHigh(v)
	}
}

func (b *Bus) writePSG(offset uint32, v byte) {
	switch offset {
	case 0:
		b.psg.SelectChannel(v)
	case 1:
		b.psg.WriteMainVolumeLeft(v)
	case 2:
		b.psg.WriteMainVolumeRight(v)
	case 4:
		b.psg.WriteFreqLow(v)
	case 5:
		b.psg.WriteFreqHigh(v)
	case 6:
		b.psg.WriteControl(v)
	case 7:
		b.psg.WritePan(v)
	case 8:
		b.psg.WriteWaveData(v)
	case 9:
		b.psg.WriteNoiseControl(v)
	}
}

func (b *Bus) writeTimer(offset uint32, v byte) {
	switch offset {
	case 0:
		b.tirq.WriteTimerReload(v)
	case 1:
		b.tirq.WriteTimerControl(v)
	}
}
