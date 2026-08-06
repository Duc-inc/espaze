package gameboy

import (
	"fmt"

	"github.com/Duc-inc/espaze/internal/emulation/core"
	emuinput "github.com/Duc-inc/espaze/internal/emulation/input"
	"github.com/Duc-inc/espaze/internal/emulation/video"

	"github.com/Duc-inc/espaze/internal/systems/gameboy/cpu"
	"github.com/Duc-inc/espaze/internal/systems/gameboy/joypad"
	"github.com/Duc-inc/espaze/internal/systems/gameboy/memory"
	"github.com/Duc-inc/espaze/internal/systems/gameboy/ppu"
	"github.com/Duc-inc/espaze/internal/systems/gameboy/timer"
)

const systemID = "gameboy"

// cyclesPerFrame is exactly one LCD refresh's worth of T-cycles on real
// hardware: 154 scanlines * 456 cycles. Driving the CPU/PPU/timer by
// this fixed budget every StepFrame reproduces the real ~59.73Hz refresh
// rate instead of an approximated 60Hz.
const cyclesPerFrame = 154 * 456
const cpuClockHz = 4194304.0

// GameBoy wires the CPU, MMU, PPU, timer and joypad together into a
// core.Core implementation. Audio is not implemented yet (see DrainAudio)
// - every other subsystem (CPU opcodes, PPU rendering, MBC0/MBC1
// banking, timer, joypad) is a from-scratch implementation of the
// documented DMG hardware behavior.
type GameBoy struct {
	mbc    memory.MBC
	mmu    *memory.MMU
	proc   *cpu.CPU
	video  *ppu.PPU
	tmr    *timer.Timer
	pad    *joypad.Joypad
	loaded bool
}

// New constructs a core with its stateless components ready; LoadROM
// still needs to be called before StepFrame does anything useful.
func New() core.Core {
	return &GameBoy{
		video: ppu.New(),
		tmr:   timer.New(),
		pad:   joypad.New(),
	}
}

func init() {
	core.Register(Metadata(), New)
}

// Metadata describes the Game Boy core for the registry and scanner.
func Metadata() core.Metadata {
	return core.Metadata{
		ID:              systemID,
		Name:            "Game Boy",
		Extensions:      []string{".gb"},
		FramesPerSecond: cpuClockHz / cyclesPerFrame,
		ScreenWidth:     ppu.Width,
		ScreenHeight:    ppu.Height,
	}
}

// Metadata implements core.Core.
func (gb *GameBoy) Metadata() core.Metadata { return Metadata() }

// LoadROM implements core.Core: parses the cartridge header, picks the
// right MBC, and builds a fresh MMU/CPU around it.
func (gb *GameBoy) LoadROM(data []byte) error {
	cart, err := memory.ParseCartridge(data)
	if err != nil {
		return fmt.Errorf("gameboy: %w", err)
	}

	gb.video.Reset()
	gb.tmr.Reset()
	gb.pad.Reset()

	gb.mbc = memory.NewMBC(cart)
	gb.mmu = memory.New(gb.mbc, gb.video, gb.tmr, gb.pad)
	// cpu.New() pokes the post-boot I/O register values (LCDC, BGP, ...)
	// through the MMU, so it must run after the components above are
	// wired up and reset, not before - otherwise their own Reset calls
	// would wipe out what it just wrote.
	gb.proc = cpu.New(gb.mmu)
	gb.loaded = true
	return nil
}

// Reset implements core.Core.
func (gb *GameBoy) Reset() {
	if !gb.loaded {
		return
	}
	gb.video.Reset()
	gb.tmr.Reset()
	gb.pad.Reset()
	// Reset last: it re-pokes the post-boot I/O register values, which
	// must win over the component resets above.
	gb.proc.Reset()
}

// StepFrame implements core.Core: runs the CPU/PPU/timer together for
// exactly one LCD refresh's worth of T-cycles.
func (gb *GameBoy) StepFrame() error {
	if !gb.loaded {
		return nil
	}

	spent := 0
	for spent < cyclesPerFrame {
		cycles := gb.proc.Step()

		interrupts := gb.video.Step(cycles)
		if gb.tmr.Step(cycles) {
			interrupts |= memory.InterruptTimer
		}
		gb.mmu.RequestInterrupt(interrupts)

		spent += cycles
	}
	return nil
}

// FrameBuffer implements core.Core.
func (gb *GameBoy) FrameBuffer() *video.FrameBuffer { return gb.video.FrameBuffer() }

// DrainAudio implements core.Core. The APU (4 sound channels) hasn't
// been built yet, so this always reports silence - a disclosed gap, not
// an oversight: every other subsystem is fully implemented.
func (gb *GameBoy) DrainAudio() ([]int16, int) { return nil, 44100 }

// SetInput implements core.Core.
func (gb *GameBoy) SetInput(state emuinput.State) { gb.pad.SetButtons(state) }
