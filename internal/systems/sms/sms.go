package sms

import (
	"fmt"

	"github.com/Duc-inc/espaze/internal/emulation/core"
	emuinput "github.com/Duc-inc/espaze/internal/emulation/input"
	"github.com/Duc-inc/espaze/internal/emulation/video"

	"github.com/Duc-inc/espaze/internal/systems/sms/cpu"
	"github.com/Duc-inc/espaze/internal/systems/sms/memory"
	"github.com/Duc-inc/espaze/internal/systems/sms/psg"
	"github.com/Duc-inc/espaze/internal/systems/sms/vdp"
)

const systemID = "sms"

const cpuClockHz = 3579545.0 // NTSC master clock, shared by the Z80 and PSG
const cyclesPerFrame = 262 * 228

// SMS wires the Z80, VDP, PSG and bus together into a core.Core
// implementation - a from-scratch implementation of the documented
// Sega Master System hardware.
type SMS struct {
	bus       *memory.Bus
	proc      *cpu.CPU
	video     *vdp.VDP
	sound     *psg.PSG
	pausePrev bool
	loaded    bool
}

// New constructs a core with its stateless components ready; LoadROM
// still needs to be called before StepFrame does anything useful.
func New() core.Core {
	return &SMS{video: vdp.New(), sound: psg.New()}
}

func init() {
	core.Register(Metadata(), New)
}

// Metadata describes the Master System core for the registry and scanner.
func Metadata() core.Metadata {
	return core.Metadata{
		ID:              systemID,
		Name:            "Master System",
		Extensions:      []string{".sms"},
		FramesPerSecond: cpuClockHz / cyclesPerFrame,
		ScreenWidth:     vdp.Width,
		ScreenHeight:    vdp.Height,
	}
}

// Metadata implements core.Core.
func (s *SMS) Metadata() core.Metadata { return Metadata() }

// LoadROM implements core.Core. SMS cartridges are raw binary images -
// no header to parse, unlike NES/GBC - so this just hands the bytes
// straight to a fresh bus built around them.
func (s *SMS) LoadROM(data []byte) error {
	if len(data) == 0 {
		return fmt.Errorf("sms: empty ROM")
	}

	s.video.Reset()
	s.sound.Reset()
	s.bus = memory.New(data, s.video, s.sound)
	s.proc = cpu.New(s.bus, s.bus) // Bus satisfies both cpu.Bus (memory) and cpu.IOBus (ports)
	s.pausePrev = false
	s.loaded = true
	return nil
}

// Reset implements core.Core.
func (s *SMS) Reset() {
	if !s.loaded {
		return
	}
	s.video.Reset()
	s.sound.Reset()
	s.proc.Reset()
	s.pausePrev = false
}

// StepFrame implements core.Core: runs the CPU/VDP/PSG together for
// exactly one NTSC frame's worth of Z80 T-states.
func (s *SMS) StepFrame() error {
	if !s.loaded {
		return nil
	}

	spent := 0
	for spent < cyclesPerFrame {
		cycles := s.proc.Step()

		if irq := s.video.Step(cycles); irq&(vdp.IRQFrame|vdp.IRQLine) != 0 {
			s.proc.TriggerInterrupt(0xFF)
		}
		s.sound.Step(cycles)

		spent += cycles
	}

	// Pause is wired directly to the Z80's NMI line on real hardware,
	// not through either I/O port - and it's edge-triggered, so holding
	// it down doesn't keep re-firing the interrupt every frame.
	pausedNow := s.bus.PausePressed()
	if pausedNow && !s.pausePrev {
		s.proc.TriggerNMI()
	}
	s.pausePrev = pausedNow

	return nil
}

// FrameBuffer implements core.Core.
func (s *SMS) FrameBuffer() *video.FrameBuffer { return s.video.FrameBuffer() }

// DrainAudio implements core.Core.
func (s *SMS) DrainAudio() ([]int16, int) {
	return s.sound.DrainSamples(), psg.SampleRate
}

// SetInput implements core.Core.
func (s *SMS) SetInput(state emuinput.State) { s.bus.SetButtons(state) }
