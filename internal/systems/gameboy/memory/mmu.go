package memory

import (
	"github.com/Duc-inc/espaze/internal/systems/gameboy/apu"
	"github.com/Duc-inc/espaze/internal/systems/gameboy/joypad"
	"github.com/Duc-inc/espaze/internal/systems/gameboy/ppu"
	"github.com/Duc-inc/espaze/internal/systems/gameboy/timer"
)

// Interrupt Flag/Enable bit positions, shared by every component that can
// raise one (see RequestInterrupt).
const (
	InterruptVBlank = 1 << 0
	InterruptSTAT   = 1 << 1
	InterruptTimer  = 1 << 2
	InterruptSerial = 1 << 3
	InterruptJoypad = 1 << 4
)

// MMU is the Game Boy's full 16-bit address bus: it owns WRAM/HRAM
// directly and routes everything else (cartridge, video, timer, joypad)
// to the component that actually implements it. It is the cpu.Bus the
// CPU package talks to.
type MMU struct {
	mbc MBC
	ppu *ppu.PPU
	tmr *timer.Timer
	pad *joypad.Joypad
	snd *apu.APU

	wram [0x2000]byte
	hram [0x7F]byte

	ifReg byte // 0xFF0F
	ieReg byte // 0xFFFF

	serial [2]byte // 0xFF01-0xFF02, stubbed
}

// New wires the MMU to the cartridge controller and the four components
// it needs to forward register access to.
func New(mbc MBC, p *ppu.PPU, t *timer.Timer, j *joypad.Joypad, a *apu.APU) *MMU {
	return &MMU{mbc: mbc, ppu: p, tmr: t, pad: j, snd: a}
}

// RequestInterrupt sets one or more bits in IF, called by gameboy.go after
// stepping the PPU/timer/joypad each cycle.
func (m *MMU) RequestInterrupt(bits byte) {
	m.ifReg |= bits
}

// Read implements cpu.Bus.
func (m *MMU) Read(addr uint16) byte {
	switch {
	case addr <= 0x7FFF:
		return m.mbc.ReadROM(addr)
	case addr <= 0x9FFF:
		return m.ppu.ReadVRAM(addr)
	case addr <= 0xBFFF:
		return m.mbc.ReadRAM(addr)
	case addr <= 0xDFFF:
		return m.wram[addr-0xC000]
	case addr <= 0xFDFF:
		return m.wram[addr-0xE000]
	case addr <= 0xFE9F:
		return m.ppu.ReadOAM(addr)
	case addr <= 0xFEFF:
		return 0xFF
	case addr == 0xFF00:
		return m.pad.ReadRegister()
	case addr <= 0xFF02:
		return m.serial[addr-0xFF01]
	case addr <= 0xFF07:
		return m.tmr.ReadRegister(addr)
	case addr == 0xFF0F:
		return m.ifReg | 0xE0
	case addr <= 0xFF3F:
		return m.snd.ReadRegister(addr)
	case addr <= 0xFF4B:
		return m.ppu.ReadRegister(addr)
	case addr <= 0xFF7F:
		return 0xFF
	case addr <= 0xFFFE:
		return m.hram[addr-0xFF80]
	default: // 0xFFFF
		return m.ieReg
	}
}

// Write implements cpu.Bus.
func (m *MMU) Write(addr uint16, v byte) {
	switch {
	case addr <= 0x7FFF:
		m.mbc.WriteROM(addr, v)
	case addr <= 0x9FFF:
		m.ppu.WriteVRAM(addr, v)
	case addr <= 0xBFFF:
		m.mbc.WriteRAM(addr, v)
	case addr <= 0xDFFF:
		m.wram[addr-0xC000] = v
	case addr <= 0xFDFF:
		m.wram[addr-0xE000] = v
	case addr <= 0xFE9F:
		m.ppu.WriteOAM(addr, v)
	case addr <= 0xFEFF:
		// unusable
	case addr == 0xFF00:
		m.pad.WriteRegister(v)
	case addr <= 0xFF02:
		m.serial[addr-0xFF01] = v
	case addr <= 0xFF07:
		m.tmr.WriteRegister(addr, v)
	case addr == 0xFF0F:
		m.ifReg = v & 0x1F
	case addr == 0xFF46:
		m.oamDMA(v)
	case addr <= 0xFF3F:
		m.snd.WriteRegister(addr, v)
	case addr <= 0xFF4B:
		m.ppu.WriteRegister(addr, v)
	case addr <= 0xFF7F:
		// unmapped IO
	case addr <= 0xFFFE:
		m.hram[addr-0xFF80] = v
	default: // 0xFFFF
		m.ieReg = v
	}
}

// oamDMA copies 0xA0 bytes from (src*0x100) into OAM, the way writing to
// FF46 does on real hardware (timing is simplified to be instantaneous).
func (m *MMU) oamDMA(src byte) {
	base := uint16(src) << 8
	for i := uint16(0); i < 0xA0; i++ {
		m.ppu.WriteOAM(0xFE00+i, m.Read(base+i))
	}
}
