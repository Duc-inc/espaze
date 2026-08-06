package chip8

import (
	emuaudio "github.com/Duc-inc/espaze/internal/emulation/audio"
	"github.com/Duc-inc/espaze/internal/emulation/core"
	emuinput "github.com/Duc-inc/espaze/internal/emulation/input"
	"github.com/Duc-inc/espaze/internal/emulation/video"

	"github.com/Duc-inc/espaze/internal/systems/chip8/cpu"
	"github.com/Duc-inc/espaze/internal/systems/chip8/display"
	chip8input "github.com/Duc-inc/espaze/internal/systems/chip8/input"
	"github.com/Duc-inc/espaze/internal/systems/chip8/memory"
	"github.com/Duc-inc/espaze/internal/systems/chip8/rom"
	"github.com/Duc-inc/espaze/internal/systems/chip8/timer"
)

// systemID is how this core registers itself and how save states are tagged.
const systemID = "chip8"

// cyclesPerFrame approximates the ~700Hz instruction rate most CHIP-8
// programs assume, run in bursts once per 60Hz output frame.
const cyclesPerFrame = 11

// sampleRate is the audio rate used for the synthesized beep tone.
const sampleRate = 44100

// Chip8 wires together the CPU, memory, display, keypad and timers into a
// single core.Core implementation - the only thing the rest of the app
// interacts with directly.
type Chip8 struct {
	mem     *memory.Memory
	disp    *display.Display
	keys    *chip8input.Keypad
	delay   *timer.Timer
	sound   *timer.Timer
	proc    *cpu.CPU
	audio   *emuaudio.Buffer
	beepPos int
}

// New constructs a fresh, unloaded CHIP-8 core. Matches core.Factory.
func New() core.Core {
	mem := memory.New()
	disp := display.New()
	keys := chip8input.New()
	delay := timer.New()
	sound := timer.New()
	return &Chip8{
		mem:   mem,
		disp:  disp,
		keys:  keys,
		delay: delay,
		sound: sound,
		proc:  cpu.New(mem, disp, keys, delay, sound),
		audio: emuaudio.NewBuffer(),
	}
}

func init() {
	core.Register(Metadata(), New)
}

// Metadata describes the CHIP-8 core for the registry and library scanner.
func Metadata() core.Metadata {
	return core.Metadata{
		ID:              systemID,
		Name:            "CHIP-8",
		Extensions:      []string{".ch8", ".c8", ".chip8"},
		FramesPerSecond: 60,
		ScreenWidth:     display.Width,
		ScreenHeight:    display.Height,
	}
}

// Metadata implements core.Core.
func (c *Chip8) Metadata() core.Metadata { return Metadata() }

// LoadROM implements core.Core.
func (c *Chip8) LoadROM(data []byte) error {
	c.Reset()
	return rom.Load(c.mem, data)
}

// Reset implements core.Core.
func (c *Chip8) Reset() {
	c.mem.Reset()
	c.disp.Clear()
	c.keys.Reset()
	c.delay.Set(0)
	c.sound.Set(0)
	c.proc.Reset()
}

// StepFrame implements core.Core: run a burst of CPU cycles, then tick
// both timers once, matching the original interpreter's 60Hz cadence.
func (c *Chip8) StepFrame() error {
	for i := 0; i < cyclesPerFrame; i++ {
		if err := c.proc.Step(); err != nil {
			return err
		}
	}
	c.delay.Tick()
	c.sound.Tick()
	c.synthesizeAudio()
	return nil
}

// FrameBuffer implements core.Core, converting the 1-bit display into RGBA.
func (c *Chip8) FrameBuffer() *video.FrameBuffer {
	fb := video.NewFrameBuffer(display.Width, display.Height)
	pixels := c.disp.Pixels()
	for i, on := range pixels {
		x := i % display.Width
		y := i / display.Width
		if on {
			fb.SetPixel(x, y, 0xE6, 0xE6, 0xE6, 0xFF)
		} else {
			fb.SetPixel(x, y, 0x0B, 0x0B, 0x0B, 0xFF)
		}
	}
	return fb
}

// DrainAudio implements core.Core.
func (c *Chip8) DrainAudio() ([]int16, int) {
	return c.audio.Drain(), sampleRate
}

// SetInput implements core.Core, mapping generic bits 0x0-0xF straight onto
// the 16-key hex keypad.
func (c *Chip8) SetInput(state emuinput.State) {
	for key := uint8(0); key < 16; key++ {
		c.keys.Set(key, state.Pressed(key))
	}
}
