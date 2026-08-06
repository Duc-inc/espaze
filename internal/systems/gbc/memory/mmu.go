package memory

import (
	"github.com/Duc-inc/espaze/internal/systems/gameboy/apu"
	"github.com/Duc-inc/espaze/internal/systems/gameboy/joypad"
	dmgmem "github.com/Duc-inc/espaze/internal/systems/gameboy/memory"
	"github.com/Duc-inc/espaze/internal/systems/gameboy/timer"
	"github.com/Duc-inc/espaze/internal/systems/gbc/ppu"
)

// Interrupt Flag/Enable bit positions - identical to DMG.
const (
	InterruptVBlank = 1 << 0
	InterruptSTAT   = 1 << 1
	InterruptTimer  = 1 << 2
	InterruptSerial = 1 << 3
	InterruptJoypad = 1 << 4
)

// MMU is the CGB's full 16-bit address bus. It's the DMG MMU's
// extension: the same cartridge/timer/joypad/APU wiring (reused
// directly from the gameboy package - none of that hardware changed on
// CGB), plus banked work RAM, the color PPU, HDMA, and the speed switch.
type MMU struct {
	mbc dmgmem.MBC
	ppu *ppu.PPU
	tmr *timer.Timer
	pad *joypad.Joypad
	snd *apu.APU

	wram  *wram
	hram  [0x7F]byte
	hdma  hdma
	speed speedSwitch

	ifReg byte
	ieReg byte

	serial [2]byte
}

// New wires the MMU to the cartridge controller and the components it
// forwards register access to.
func New(mbc dmgmem.MBC, p *ppu.PPU, t *timer.Timer, j *joypad.Joypad, a *apu.APU) *MMU {
	return &MMU{mbc: mbc, ppu: p, tmr: t, pad: j, snd: a, wram: newWRAM()}
}

// RequestInterrupt sets one or more bits in IF.
func (m *MMU) RequestInterrupt(bits byte) { m.ifReg |= bits }

// DoubleSpeed reports whether KEY1 currently has the CPU in double-speed
// mode - gbc.go uses this to halve how many real-time (PPU/timer/APU)
// cycles each CPU cycle is worth.
func (m *MMU) DoubleSpeed() bool { return m.speed.doubleSpeed }

// Read implements cpu.Bus.
func (m *MMU) Read(addr uint16) byte {
	switch {
	case addr <= 0x7FFF:
		return m.mbc.ReadROM(addr)
	case addr <= 0x9FFF:
		return m.ppu.ReadVRAM(addr)
	case addr <= 0xBFFF:
		return m.mbc.ReadRAM(addr)
	case addr <= 0xCFFF:
		return m.wram.readLow(addr - 0xC000)
	case addr <= 0xDFFF:
		return m.wram.readHigh(addr - 0xD000)
	case addr <= 0xEFFF:
		return m.wram.readLow(addr - 0xE000)
	case addr <= 0xFDFF:
		return m.wram.readHigh(addr - 0xF000)
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
	case addr <= 0xFF45:
		return m.ppu.ReadRegister(addr)
	case addr == 0xFF46:
		return 0xFF // OAM DMA is write-only
	case addr <= 0xFF4B:
		return m.ppu.ReadRegister(addr)
	case addr == 0xFF4D:
		return m.speed.readKEY1()
	case addr == 0xFF4F:
		return m.ppu.ReadRegister(addr)
	case addr <= 0xFF55:
		return 0xFF // HDMA source/dest are write-only; transfers always complete instantly (see hdma.go)
	case addr <= 0xFF67:
		return 0xFF
	case addr <= 0xFF6B:
		return m.ppu.ReadRegister(addr)
	case addr <= 0xFF6F:
		return 0xFF
	case addr == 0xFF70:
		return m.wram.readSVBK()
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
	case addr <= 0xCFFF:
		m.wram.writeLow(addr-0xC000, v)
	case addr <= 0xDFFF:
		m.wram.writeHigh(addr-0xD000, v)
	case addr <= 0xEFFF:
		m.wram.writeLow(addr-0xE000, v)
	case addr <= 0xFDFF:
		m.wram.writeHigh(addr-0xF000, v)
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
	case addr <= 0xFF3F:
		m.snd.WriteRegister(addr, v)
	case addr <= 0xFF45:
		m.ppu.WriteRegister(addr, v)
	case addr == 0xFF46:
		m.oamDMA(v)
	case addr <= 0xFF4B:
		m.ppu.WriteRegister(addr, v)
	case addr == 0xFF4D:
		m.speed.writeKEY1(v)
	case addr == 0xFF4F:
		m.ppu.WriteRegister(addr, v)
	case addr == 0xFF51:
		m.hdma.writeSrcHi(v)
	case addr == 0xFF52:
		m.hdma.writeSrcLo(v)
	case addr == 0xFF53:
		m.hdma.writeDstHi(v)
	case addr == 0xFF54:
		m.hdma.writeDstLo(v)
	case addr == 0xFF55:
		m.hdma.transfer(v, m.Read, m.ppu.WriteVRAM)
	case addr <= 0xFF67:
		// unmapped
	case addr <= 0xFF6B:
		m.ppu.WriteRegister(addr, v)
	case addr <= 0xFF6F:
		// unmapped
	case addr == 0xFF70:
		m.wram.writeSVBK(v)
	case addr <= 0xFF7F:
		// unmapped
	case addr <= 0xFFFE:
		m.hram[addr-0xFF80] = v
	default: // 0xFFFF
		m.ieReg = v
	}
}

// oamDMA copies 0xA0 bytes from (src*0x100) into OAM.
func (m *MMU) oamDMA(src byte) {
	base := uint16(src) << 8
	for i := uint16(0); i < 0xA0; i++ {
		m.ppu.WriteOAM(0xFE00+i, m.Read(base+i))
	}
}
