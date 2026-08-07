// Package pcengine wires a from-scratch HuC6280 CPU, HuC6270 VDC,
// HuC6260 VCE, and built-in PSG together into a core.Core
// implementation - the NEC PC Engine / TurboGrafx-16. See each
// subpackage's own doc comment for what's simplified and why; the
// cpu and memory packages both flag that some register/opcode byte
// assignments are this project's best-effort reconstruction rather
// than independently verified against real hardware.
package pcengine

import (
	"fmt"

	"github.com/Duc-inc/espaze/internal/emulation/core"
	emuinput "github.com/Duc-inc/espaze/internal/emulation/input"
	"github.com/Duc-inc/espaze/internal/emulation/video"

	"github.com/Duc-inc/espaze/internal/systems/pcengine/cpu"
	"github.com/Duc-inc/espaze/internal/systems/pcengine/memory"
	"github.com/Duc-inc/espaze/internal/systems/pcengine/psg"
	"github.com/Duc-inc/espaze/internal/systems/pcengine/vce"
	"github.com/Duc-inc/espaze/internal/systems/pcengine/vdc"
)

const systemID = "pcengine"

const cpuClockHz = 7159090.0
const cyclesPerFrame = 262 * 455 // matches vdc's own scanline timing assumption

// PCEngine wires the reused-nowhere-else HuC6280 core to a
// from-scratch VDC/VCE/PSG.
type PCEngine struct {
	bus   *memory.Bus
	proc  *cpu.CPU
	video *vdc.VDC
	color *vce.VCE
	sound *psg.PSG

	loaded bool
}

// New constructs a core with its stateless components ready; LoadROM
// still needs to be called before StepFrame does anything useful.
func New() core.Core {
	color := vce.New()
	return &PCEngine{color: color, video: vdc.New(color), sound: psg.New()}
}

func init() {
	core.Register(Metadata(), New)
}

// Metadata describes the PC Engine core for the registry and scanner.
func Metadata() core.Metadata {
	return core.Metadata{
		ID:              systemID,
		Name:            "PC Engine / TurboGrafx-16",
		Extensions:      []string{".pce"},
		FramesPerSecond: cpuClockHz / cyclesPerFrame,
		ScreenWidth:     vdc.Width,
		ScreenHeight:    vdc.Height,
	}
}

// Metadata implements core.Core.
func (p *PCEngine) Metadata() core.Metadata { return Metadata() }

// LoadROM implements core.Core. HuCard images are raw binary - no
// header to parse.
func (p *PCEngine) LoadROM(data []byte) error {
	if len(data) == 0 {
		return fmt.Errorf("pcengine: empty ROM")
	}

	p.video.Reset()
	p.color.Reset()
	p.sound.Reset()
	p.bus = memory.New(data, p.video, p.color, p.sound, nil)
	p.proc = cpu.New(p.bus)
	p.bus.SetTimerIRQ(p.proc)
	p.loaded = true
	return nil
}

// Reset implements core.Core.
func (p *PCEngine) Reset() {
	if !p.loaded {
		return
	}
	p.video.Reset()
	p.color.Reset()
	p.sound.Reset()
	p.proc.Reset()
}

// StepFrame implements core.Core: runs the CPU/VDC/PSG together for
// exactly one 262-line raster.
func (p *PCEngine) StepFrame() error {
	if !p.loaded {
		return nil
	}

	spent := 0
	for spent < cyclesPerFrame {
		cycles := p.proc.Step()

		if irq := p.video.Step(cycles); irq&(vdc.IRQVBlank|vdc.IRQLine) != 0 {
			p.proc.TriggerIRQ2()
		}
		p.sound.Step(cycles)

		spent += cycles
	}

	return nil
}

// FrameBuffer implements core.Core.
func (p *PCEngine) FrameBuffer() *video.FrameBuffer { return p.video.FrameBuffer() }

// DrainAudio implements core.Core.
func (p *PCEngine) DrainAudio() ([]int16, int) {
	return p.sound.DrainSamples(), psg.SampleRate
}

// SetInput implements core.Core.
func (p *PCEngine) SetInput(state emuinput.State) { p.bus.SetButtons(state) }
