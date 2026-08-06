package gbc

import (
	"fmt"

	"github.com/Duc-inc/espaze/internal/emulation/core"
	emuinput "github.com/Duc-inc/espaze/internal/emulation/input"
	"github.com/Duc-inc/espaze/internal/emulation/video"

	"github.com/Duc-inc/espaze/internal/systems/gameboy/apu"
	"github.com/Duc-inc/espaze/internal/systems/gameboy/cpu"
	"github.com/Duc-inc/espaze/internal/systems/gameboy/joypad"
	dmgmem "github.com/Duc-inc/espaze/internal/systems/gameboy/memory"
	"github.com/Duc-inc/espaze/internal/systems/gameboy/timer"
	"github.com/Duc-inc/espaze/internal/systems/gbc/memory"
	"github.com/Duc-inc/espaze/internal/systems/gbc/ppu"
)

const systemID = "gbc"

// cyclesPerFrame is exactly one LCD refresh's worth of T-cycles, in
// real time - identical to DMG's, since the PPU/timer/APU always run at
// the same real-time rate regardless of the CPU's speed mode.
const cyclesPerFrame = 154 * 456
const cpuClockHz = 4194304.0

// GBC wires the CPU, MMU, PPU, timer, joypad and APU together into a
// core.Core implementation. It reuses the DMG core's CPU, timer,
// joypad and APU packages unchanged (none of that hardware differs on
// CGB) and adds only what actually changed: a color-capable PPU, banked
// work RAM, HDMA, and the double-speed switch, all behind its own MMU.
type GBC struct {
	mbc    dmgmem.MBC
	mmu    *memory.MMU
	proc   *cpu.CPU
	video  *ppu.PPU
	tmr    *timer.Timer
	pad    *joypad.Joypad
	sound  *apu.APU
	loaded bool
}

// New constructs a core with its stateless components ready; LoadROM
// still needs to be called before StepFrame does anything useful.
func New() core.Core {
	return &GBC{
		video: ppu.New(),
		tmr:   timer.New(),
		pad:   joypad.New(),
		sound: apu.New(),
	}
}

func init() {
	core.Register(Metadata(), New)
}

// Metadata describes the Game Boy Color core for the registry and scanner.
func Metadata() core.Metadata {
	return core.Metadata{
		ID:              systemID,
		Name:            "Game Boy Color",
		Extensions:      []string{".gbc"},
		FramesPerSecond: cpuClockHz / cyclesPerFrame,
		ScreenWidth:     ppu.Width,
		ScreenHeight:    ppu.Height,
	}
}

// Metadata implements core.Core.
func (gb *GBC) Metadata() core.Metadata { return Metadata() }

// LoadROM implements core.Core: parses the cartridge header (reusing
// the DMG parser and MBC0/1/5 controllers unchanged - a CGB cartridge's
// header and bank-switching hardware aren't CGB-specific), and builds a
// fresh MMU/CPU around it.
func (gb *GBC) LoadROM(data []byte) error {
	cart, err := dmgmem.ParseCartridge(data)
	if err != nil {
		return fmt.Errorf("gbc: %w", err)
	}

	gb.video.Reset()
	gb.tmr.Reset()
	gb.pad.Reset()
	gb.sound.Reset()

	gb.mbc = dmgmem.NewMBC(cart)
	gb.mmu = memory.New(gb.mbc, gb.video, gb.tmr, gb.pad, gb.sound)
	// cpu.NewCGB() pokes the post-boot I/O register values (LCDC, NR52,
	// and A=0x11 marking this as CGB hardware) through the MMU, so it
	// must run after the components above are wired up and reset, not
	// before - otherwise their own Reset calls would wipe out what it
	// just wrote.
	gb.proc = cpu.NewCGB(gb.mmu)
	gb.loaded = true
	return nil
}

// Reset implements core.Core.
func (gb *GBC) Reset() {
	if !gb.loaded {
		return
	}
	gb.video.Reset()
	gb.tmr.Reset()
	gb.pad.Reset()
	gb.sound.Reset()
	// Reset last: it re-pokes the post-boot I/O register values, which
	// must win over the component resets above.
	gb.proc.Reset()
}

// StepFrame implements core.Core: runs the CPU/PPU/timer/APU together
// for exactly one LCD refresh's worth of *real* T-cycles. In
// double-speed mode the CPU consumes T-cycles twice as fast as
// real-time, so the PPU/timer/APU (which never speed up) only advance
// by half of whatever the CPU just spent.
func (gb *GBC) StepFrame() error {
	if !gb.loaded {
		return nil
	}

	spent := 0
	for spent < cyclesPerFrame {
		cpuCycles := gb.proc.Step()
		realCycles := cpuCycles
		if gb.mmu.DoubleSpeed() {
			realCycles = cpuCycles / 2
		}

		interrupts := gb.video.Step(realCycles)
		if gb.tmr.Step(realCycles) {
			interrupts |= memory.InterruptTimer
		}
		gb.mmu.RequestInterrupt(interrupts)
		gb.sound.Step(realCycles)

		spent += realCycles
	}
	return nil
}

// FrameBuffer implements core.Core.
func (gb *GBC) FrameBuffer() *video.FrameBuffer { return gb.video.FrameBuffer() }

// DrainAudio implements core.Core.
func (gb *GBC) DrainAudio() ([]int16, int) {
	return gb.sound.DrainSamples(), apu.SampleRate
}

// SetInput implements core.Core.
func (gb *GBC) SetInput(state emuinput.State) { gb.pad.SetButtons(state) }
