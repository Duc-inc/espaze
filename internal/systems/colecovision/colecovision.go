// Package colecovision wires the ColecoVision together entirely out
// of components this project already implements from scratch: the Z80
// and PSG (reused directly from internal/systems/sms - the same
// chips) and the TMS9918 VDP (internal/systems/tms9918, also shared
// with this project's SG-1000 and MSX1 cores). No new CPU or video
// chip needed writing for this platform.
package colecovision

import (
	"fmt"

	"github.com/Duc-inc/espaze/internal/emulation/core"
	emuinput "github.com/Duc-inc/espaze/internal/emulation/input"
	"github.com/Duc-inc/espaze/internal/emulation/video"

	"github.com/Duc-inc/espaze/internal/systems/colecovision/memory"
	"github.com/Duc-inc/espaze/internal/systems/sms/cpu"
	"github.com/Duc-inc/espaze/internal/systems/sms/psg"
	"github.com/Duc-inc/espaze/internal/systems/tms9918"
)

const systemID = "colecovision"

const cpuClockHz = 3579545.0
const cyclesPerFrame = 262 * 228

// ColecoVision wires the reused Z80/PSG to the shared TMS9918 VDP.
type ColecoVision struct {
	bus   *memory.Bus
	proc  *cpu.CPU
	video *tms9918.TMS9918
	sound *psg.PSG

	loaded bool
}

// New constructs a core with its stateless components ready; LoadROM
// still needs to be called before StepFrame does anything useful.
func New() core.Core {
	return &ColecoVision{video: tms9918.New(), sound: psg.New()}
}

func init() {
	core.Register(Metadata(), New)
}

// Metadata describes the ColecoVision core for the registry and scanner.
func Metadata() core.Metadata {
	return core.Metadata{
		ID:              systemID,
		Name:            "ColecoVision",
		Extensions:      []string{".col"},
		FramesPerSecond: cpuClockHz / cyclesPerFrame,
		ScreenWidth:     tms9918.Width,
		ScreenHeight:    tms9918.Height,
	}
}

// Metadata implements core.Core.
func (c *ColecoVision) Metadata() core.Metadata { return Metadata() }

// LoadROM implements core.Core. ColecoVision cartridges are raw
// binary images - no header to parse.
func (c *ColecoVision) LoadROM(data []byte) error {
	if len(data) == 0 {
		return fmt.Errorf("colecovision: empty ROM")
	}

	c.video.Reset()
	c.sound.Reset()
	c.bus = memory.New(data, c.video, c.sound)
	c.proc = cpu.New(c.bus, c.bus)
	c.loaded = true
	return nil
}

// Reset implements core.Core.
func (c *ColecoVision) Reset() {
	if !c.loaded {
		return
	}
	c.video.Reset()
	c.sound.Reset()
	c.proc.Reset()
}

// StepFrame implements core.Core.
func (c *ColecoVision) StepFrame() error {
	if !c.loaded {
		return nil
	}

	spent := 0
	for spent < cyclesPerFrame {
		cycles := c.proc.Step()
		if irq := c.video.Step(cycles); irq&tms9918.IRQVBlank != 0 {
			c.proc.TriggerInterrupt(0xFF)
		}
		c.sound.Step(cycles)
		spent += cycles
	}
	return nil
}

// FrameBuffer implements core.Core.
func (c *ColecoVision) FrameBuffer() *video.FrameBuffer { return c.video.FrameBuffer() }

// DrainAudio implements core.Core.
func (c *ColecoVision) DrainAudio() ([]int16, int) {
	return c.sound.DrainSamples(), psg.SampleRate
}

// SetInput implements core.Core.
func (c *ColecoVision) SetInput(state emuinput.State) { c.bus.SetButtons(state) }
