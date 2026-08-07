// Package genesis wires a from-scratch Motorola 68000, the Genesis
// VDP, the YM2612 FM synth, and a Z80 audio coprocessor (reusing this
// project's own SMS Z80 core and PSG directly - both genuinely the
// same chips on real hardware) into a core.Core implementation.
package genesis

import (
	"fmt"

	"github.com/Duc-inc/espaze/internal/emulation/core"
	emuinput "github.com/Duc-inc/espaze/internal/emulation/input"
	"github.com/Duc-inc/espaze/internal/emulation/video"

	"github.com/Duc-inc/espaze/internal/systems/genesis/audio"
	"github.com/Duc-inc/espaze/internal/systems/genesis/cpu"
	"github.com/Duc-inc/espaze/internal/systems/genesis/memory"
	"github.com/Duc-inc/espaze/internal/systems/genesis/vdp"
	"github.com/Duc-inc/espaze/internal/systems/genesis/ym2612"
	sms "github.com/Duc-inc/espaze/internal/systems/sms/cpu"
	"github.com/Duc-inc/espaze/internal/systems/sms/psg"
)

const systemID = "genesis"

// m68kClockHz/cyclesPerFrame reproduce the VDP's own NTSC assumption
// (488 cycles/line * 262 lines - see genesis/vdp.cyclesPerLine).
const m68kClockHz = 7670000.0
const cyclesPerFrame = 488 * 262

// z80ClockHz is the audio coprocessor's clock - the same 3.579545MHz
// master clock this project's SMS core already runs its Z80 at.
const z80ClockHz = 3579545.0

// Genesis wires every component together into a core.Core
// implementation.
type Genesis struct {
	cpu   *cpu.CPU
	bus   *memory.Bus
	video *vdp.VDP
	ym    *ym2612.YM2612
	sound *psg.PSG

	z80     *sms.CPU
	z80bus  *audio.Bus
	z80Left float64

	prevZ80Reset bool
	loaded       bool
}

// New constructs a core with its stateless components ready; LoadROM
// still needs to be called before StepFrame does anything useful.
func New() core.Core {
	return &Genesis{video: vdp.New(), ym: ym2612.New(), sound: psg.New()}
}

func init() {
	core.Register(Metadata(), New)
}

// Metadata describes the Genesis core for the registry and scanner.
func Metadata() core.Metadata {
	return core.Metadata{
		ID:              systemID,
		Name:            "Genesis / Mega Drive",
		Extensions:      []string{".md", ".bin", ".gen"},
		FramesPerSecond: m68kClockHz / cyclesPerFrame,
		ScreenWidth:     vdp.Width,
		ScreenHeight:    vdp.Height,
	}
}

// Metadata implements core.Core.
func (g *Genesis) Metadata() core.Metadata { return Metadata() }

// LoadROM implements core.Core. Genesis cartridges are raw binary
// images - no header to parse, like this project's SMS core.
func (g *Genesis) LoadROM(data []byte) error {
	if len(data) == 0 {
		return fmt.Errorf("genesis: empty ROM")
	}

	g.video.Reset()
	g.ym.Reset()
	g.sound.Reset()

	g.z80bus = audio.New(g.ym, g.sound)
	g.z80 = sms.New(g.z80bus, g.z80bus)

	g.bus = memory.New(data, g.video, g.z80bus)
	g.video.SetMemory(g.bus)
	g.cpu = cpu.New(g.bus)

	g.z80Left = 0
	g.prevZ80Reset = false
	g.loaded = true
	return nil
}

// Reset implements core.Core.
func (g *Genesis) Reset() {
	if !g.loaded {
		return
	}
	g.video.Reset()
	g.ym.Reset()
	g.sound.Reset()
	g.z80bus.Reset()
	g.z80.Reset()
	g.cpu.Reset()
	g.z80Left = 0
	g.prevZ80Reset = false
}

// StepFrame implements core.Core: runs the 68000, VDP, YM2612, and the
// Z80 audio coprocessor together for one NTSC frame.
func (g *Genesis) StepFrame() error {
	if !g.loaded {
		return nil
	}

	spent := 0
	for spent < cyclesPerFrame {
		cycles := g.cpu.Step()

		if irq := g.video.Step(cycles); irq&vdp.IRQFrame != 0 {
			g.cpu.TriggerIRQ(6) // vblank is a fixed level-6 autovector on real hardware
		}
		g.ym.Step(cycles)
		g.stepZ80(cycles)

		spent += cycles
	}

	return nil
}

// stepZ80 advances the audio coprocessor by its own clock's share of
// the 68000 cycles just spent, honoring the reset and bus-request
// lines the 68000 controls.
func (g *Genesis) stepZ80(m68kCycles int) {
	resetNow := g.bus.Z80ResetAsserted()
	if g.prevZ80Reset && !resetNow {
		g.z80.Reset()
	}
	g.prevZ80Reset = resetNow

	if resetNow || g.z80bus.Halted() {
		return
	}

	g.z80Left += float64(m68kCycles) * (z80ClockHz / m68kClockHz)
	for g.z80Left > 0 {
		g.z80Left -= float64(g.z80.Step())
	}
}

// FrameBuffer implements core.Core.
func (g *Genesis) FrameBuffer() *video.FrameBuffer { return g.video.FrameBuffer() }

// DrainAudio implements core.Core: mixes the YM2612's samples with the
// backward-compatible PSG's, both generated at the same 44100Hz -
// small frame-to-frame length differences between the two (from
// rounding each chip's own cycle-to-sample conversion) are resolved by
// just mixing however many samples both streams have in common.
func (g *Genesis) DrainAudio() ([]int16, int) {
	fm := g.ym.DrainSamples()
	psgSamples := g.sound.DrainSamples()

	n := len(fm)
	if len(psgSamples) < n {
		n = len(psgSamples)
	}

	mixed := make([]int16, n)
	for i := 0; i < n; i++ {
		sum := int32(fm[i]) + int32(psgSamples[i])
		switch {
		case sum > 32767:
			sum = 32767
		case sum < -32768:
			sum = -32768
		}
		mixed[i] = int16(sum)
	}
	return mixed, ym2612.SampleRate
}

// SetInput implements core.Core.
func (g *Genesis) SetInput(state emuinput.State) { g.bus.SetButtons(state) }
