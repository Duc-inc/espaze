// Package gamegear implements Sega's Game Gear on top of this
// project's own SMS core: the same Z80, VDP, PSG, cartridge mapper,
// and joypad wiring, since the Game Gear is genuinely the same
// hardware family in a handheld shell. What's actually different -
// the 160x144 cropped viewport onto the VDP's full 256x192 plane, and
// a dedicated Start button on its own I/O port instead of Pause/NMI -
// is implemented here.
package gamegear

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

const systemID = "gamegear"

// Width/Height are the Game Gear's actual LCD resolution - a centered
// crop of the VDP's full 256x192 internal plane.
const (
	Width  = 160
	Height = 144
)

const offsetX = (vdp.Width - Width) / 2
const offsetY = (vdp.Height - Height) / 2

const cpuClockHz = 3579545.0 // same NTSC master clock as the SMS
const cyclesPerFrame = 262 * 228

// Start is the input.State bit for the Game Gear's Start button.
// Up/Down/Left/Right/Button1/Button2 reuse sms/memory's own constants
// (bits 0-5) directly since the joypad ports are identical hardware.
const Start = 6

// GameGear wires the reused SMS Z80/VDP/PSG/mapper together with the
// Start-button I/O port and the cropped output viewport.
type GameGear struct {
	bus    *memory.Bus
	io     *ioBus
	proc   *cpu.CPU
	video  *vdp.VDP
	sound  *psg.PSG
	frame  *video.FrameBuffer
	loaded bool
}

// New constructs a core with its stateless components ready; LoadROM
// still needs to be called before StepFrame does anything useful.
func New() core.Core {
	return &GameGear{video: vdp.New(), sound: psg.New(), frame: video.NewFrameBuffer(Width, Height)}
}

func init() {
	core.Register(Metadata(), New)
}

// Metadata describes the Game Gear core for the registry and scanner.
func Metadata() core.Metadata {
	return core.Metadata{
		ID:              systemID,
		Name:            "Game Gear",
		Extensions:      []string{".gg"},
		FramesPerSecond: cpuClockHz / cyclesPerFrame,
		ScreenWidth:     Width,
		ScreenHeight:    Height,
	}
}

// Metadata implements core.Core.
func (g *GameGear) Metadata() core.Metadata { return Metadata() }

// LoadROM implements core.Core. Game Gear cartridges use the same
// Sega mapper format as the SMS's - no header to parse.
func (g *GameGear) LoadROM(data []byte) error {
	if len(data) == 0 {
		return fmt.Errorf("gamegear: empty ROM")
	}

	g.video.Reset()
	g.sound.Reset()
	g.bus = memory.New(data, g.video, g.sound)
	g.io = &ioBus{mem: g.bus}
	g.proc = cpu.New(g.bus, g.io)
	g.loaded = true
	return nil
}

// Reset implements core.Core.
func (g *GameGear) Reset() {
	if !g.loaded {
		return
	}
	g.video.Reset()
	g.sound.Reset()
	g.proc.Reset()
}

// StepFrame implements core.Core: runs the CPU/VDP/PSG for exactly one
// NTSC frame's worth of Z80 T-states, then crops the VDP's full plane
// down to the Game Gear's actual viewport.
func (g *GameGear) StepFrame() error {
	if !g.loaded {
		return nil
	}

	spent := 0
	for spent < cyclesPerFrame {
		cycles := g.proc.Step()

		if irq := g.video.Step(cycles); irq&(vdp.IRQFrame|vdp.IRQLine) != 0 {
			g.proc.TriggerInterrupt(0xFF)
		}
		g.sound.Step(cycles)

		spent += cycles
	}

	g.cropFrame()
	return nil
}

// cropFrame copies the VDP's centered 160x144 window into the Game
// Gear's own output frame buffer, row by row.
func (g *GameGear) cropFrame() {
	src := g.video.FrameBuffer()
	rowBytes := Width * 4
	for y := 0; y < Height; y++ {
		srcStart := ((y+offsetY)*src.Width + offsetX) * 4
		dstStart := y * rowBytes
		copy(g.frame.Pixels[dstStart:dstStart+rowBytes], src.Pixels[srcStart:srcStart+rowBytes])
	}
}

// FrameBuffer implements core.Core.
func (g *GameGear) FrameBuffer() *video.FrameBuffer { return g.frame }

// DrainAudio implements core.Core.
func (g *GameGear) DrainAudio() ([]int16, int) {
	return g.sound.DrainSamples(), psg.SampleRate
}

// SetInput implements core.Core. Up/Down/Left/Right/Button1/Button2
// reuse the SMS joypad's exact bit layout; Start is the Game Gear's
// own addition, read through the ioBus wrapper instead of the shared
// joypad ports.
func (g *GameGear) SetInput(state emuinput.State) {
	g.bus.SetButtons(state)
	g.io.SetStart(state.Pressed(Start))
}
