// Package ngpc wires a from-scratch TLCS900H-architecture CPU, video
// chip, and a Z80 audio coprocessor (reusing this project's own SMS
// Z80 core and PSG directly) together into a core.Core implementation:
// the Neo Geo Pocket (Color). See each subpackage's own doc comment
// for the honesty caveat this platform needs more than any other in
// this project - the TLCS900H and its video chip are dramatically
// less publicly documented than every other system implemented here.
package ngpc

import (
	"fmt"

	"github.com/Duc-inc/espaze/internal/emulation/core"
	emuinput "github.com/Duc-inc/espaze/internal/emulation/input"
	"github.com/Duc-inc/espaze/internal/emulation/video"

	"github.com/Duc-inc/espaze/internal/systems/ngpc/audio"
	"github.com/Duc-inc/espaze/internal/systems/ngpc/cpu"
	"github.com/Duc-inc/espaze/internal/systems/ngpc/memory"
	"github.com/Duc-inc/espaze/internal/systems/ngpc/ppu"
	sms "github.com/Duc-inc/espaze/internal/systems/sms/cpu"
	"github.com/Duc-inc/espaze/internal/systems/sms/psg"
)

const systemID = "ngpc"

const cpuClockHz = 6144000.0
const cyclesPerFrame = 198 * 515 // matches ppu's own scanline timing assumption

const z80ClockHz = 3072000.0 // half the main CPU clock, this project's own choice

// NGPC wires the CPU/PPU/audio-coprocessor/PSG together.
type NGPC struct {
	bus   *memory.Bus
	proc  *cpu.CPU
	video *ppu.PPU
	sound *psg.PSG

	z80     *sms.CPU
	z80bus  *audio.Bus
	z80Left float64

	loaded bool
}

// New constructs a core with its stateless components ready; LoadROM
// still needs to be called before StepFrame does anything useful.
func New() core.Core {
	return &NGPC{video: ppu.New(), sound: psg.New()}
}

func init() {
	core.Register(Metadata(), New)
}

// Metadata describes the NGPC core for the registry and scanner.
func Metadata() core.Metadata {
	return core.Metadata{
		ID:              systemID,
		Name:            "Neo Geo Pocket Color",
		Extensions:      []string{".ngp", ".ngc"},
		FramesPerSecond: cpuClockHz / cyclesPerFrame,
		ScreenWidth:     ppu.Width,
		ScreenHeight:    ppu.Height,
	}
}

// Metadata implements core.Core.
func (n *NGPC) Metadata() core.Metadata { return Metadata() }

// LoadROM implements core.Core. NGPC cartridges are raw binary images
// - no header to parse.
func (n *NGPC) LoadROM(data []byte) error {
	if len(data) == 0 {
		return fmt.Errorf("ngpc: empty ROM")
	}

	n.video.Reset()
	n.sound.Reset()

	n.z80bus = audio.New(n.sound)
	n.z80 = sms.New(n.z80bus, n.z80bus)

	n.bus = memory.New(data, n.video, n.z80bus)
	n.proc = cpu.New(n.bus)

	n.z80Left = 0
	n.loaded = true
	return nil
}

// Reset implements core.Core.
func (n *NGPC) Reset() {
	if !n.loaded {
		return
	}
	n.video.Reset()
	n.sound.Reset()
	n.z80bus.Reset()
	n.z80.Reset()
	n.proc.Reset()
	n.z80Left = 0
}

// StepFrame implements core.Core: runs the CPU/PPU/Z80 coprocessor
// together for exactly one 198-line raster.
func (n *NGPC) StepFrame() error {
	if !n.loaded {
		return nil
	}

	spent := 0
	for spent < cyclesPerFrame {
		cycles := n.proc.Step()

		if irq := n.video.Step(cycles); irq&ppu.IRQVBlank != 0 {
			n.proc.TriggerIRQ(0x0004) // this project's own fixed VBlank vector
		}
		n.stepZ80(cycles)

		spent += cycles
	}

	return nil
}

// stepZ80 advances the audio coprocessor by its own clock's share of
// the main CPU cycles just spent, honoring the reset line.
func (n *NGPC) stepZ80(mainCycles int) {
	if n.bus.Z80ResetAsserted() {
		return
	}
	n.z80Left += float64(mainCycles) * (z80ClockHz / cpuClockHz)
	for n.z80Left > 0 {
		n.z80Left -= float64(n.z80.Step())
	}
}

// FrameBuffer implements core.Core.
func (n *NGPC) FrameBuffer() *video.FrameBuffer { return n.video.FrameBuffer() }

// DrainAudio implements core.Core.
func (n *NGPC) DrainAudio() ([]int16, int) {
	return n.sound.DrainSamples(), psg.SampleRate
}

// SetInput implements core.Core.
func (n *NGPC) SetInput(state emuinput.State) { n.bus.SetButtons(state) }
