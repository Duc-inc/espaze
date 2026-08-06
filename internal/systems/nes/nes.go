package nes

import (
	"fmt"

	"github.com/Duc-inc/espaze/internal/emulation/core"
	emuinput "github.com/Duc-inc/espaze/internal/emulation/input"
	"github.com/Duc-inc/espaze/internal/emulation/video"

	"github.com/Duc-inc/espaze/internal/systems/nes/apu"
	"github.com/Duc-inc/espaze/internal/systems/nes/cpu"
	"github.com/Duc-inc/espaze/internal/systems/nes/memory"
	"github.com/Duc-inc/espaze/internal/systems/nes/ppu"
)

const systemID = "nes"

const cpuClockHz = 1789773.0 // NTSC 2A03 clock

// cyclesPerFrame is the NTSC average of 341 PPU dots * 262 scanlines / 3
// dots-per-CPU-cycle (≈29780.67); using a fixed whole number here (like
// this project's other cores do for their own frame budgets) means
// audio/video drift very slightly from real hardware's exact
// frame-length wobble, imperceptibly over normal play.
const cyclesPerFrame = 29780

// NES wires the CPU, PPU, APU, mapper and bus together into a
// core.Core implementation - a from-scratch implementation of the
// documented NES/Famicom hardware behavior.
type NES struct {
	cart   *memory.Cartridge
	mapper memory.Mapper
	bus    *memory.Bus
	proc   *cpu.CPU
	video  *ppu.PPU
	sound  *apu.APU
	loaded bool
}

// New constructs a core with nothing loaded yet; LoadROM builds every
// component fresh once a cartridge (and therefore a mapper) is known.
func New() core.Core {
	return &NES{}
}

func init() {
	core.Register(Metadata(), New)
}

// Metadata describes the NES core for the registry and scanner.
func Metadata() core.Metadata {
	return core.Metadata{
		ID:              systemID,
		Name:            "NES",
		Extensions:      []string{".nes"},
		FramesPerSecond: cpuClockHz / cyclesPerFrame,
		ScreenWidth:     ppu.Width,
		ScreenHeight:    ppu.Height,
	}
}

// Metadata implements core.Core.
func (n *NES) Metadata() core.Metadata { return Metadata() }

// LoadROM implements core.Core: parses the iNES header, picks the right
// mapper, and builds a fresh PPU/APU/bus/CPU around it. The PPU and APU
// both need the mapper (CHR access, DMC sample fetches respectively)
// wired in before the CPU runs its first instruction, so construction
// order matters here.
func (n *NES) LoadROM(data []byte) error {
	cart, err := memory.ParseCartridge(data)
	if err != nil {
		return fmt.Errorf("nes: %w", err)
	}
	mapper, err := memory.NewMapper(cart)
	if err != nil {
		return fmt.Errorf("nes: %w", err)
	}

	n.cart = cart
	n.mapper = mapper
	n.video = ppu.New(mapper)
	n.sound = apu.New(nil) // DMC's memory link needs the bus, built next
	n.bus = memory.New(n.video, n.sound, mapper)
	n.sound.SetMemory(n.bus)
	n.proc = cpu.New(n.bus)
	n.proc.Reset()
	n.loaded = true
	return nil
}

// Reset implements core.Core.
func (n *NES) Reset() {
	if !n.loaded {
		return
	}
	n.video.Reset()
	n.sound.Reset() // preserves its DMC memory link, see apu.Reset
	n.proc.Reset()
}

// StepFrame implements core.Core: runs the CPU/PPU/APU together for
// roughly one NTSC frame's worth of CPU cycles.
func (n *NES) StepFrame() error {
	if !n.loaded {
		return nil
	}

	spent := 0
	for spent < cyclesPerFrame {
		cycles := n.proc.Step()
		cycles += n.bus.TakeStallCycles() // OAM DMA, if the instruction just triggered one

		if n.video.Step(cycles * 3) { // 3 PPU dots per CPU cycle (NTSC)
			n.proc.TriggerNMI()
		}
		n.sound.Step(cycles)
		if n.sound.IRQPending() {
			n.proc.TriggerIRQ()
		}

		spent += cycles
	}
	return nil
}

// FrameBuffer implements core.Core.
func (n *NES) FrameBuffer() *video.FrameBuffer { return n.video.FrameBuffer() }

// DrainAudio implements core.Core.
func (n *NES) DrainAudio() ([]int16, int) {
	return n.sound.DrainSamples(), apu.SampleRate
}

// SetInput implements core.Core.
func (n *NES) SetInput(state emuinput.State) { n.bus.SetButtons(state) }
