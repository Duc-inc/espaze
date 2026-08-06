package schip

import (
	emuaudio "github.com/Duc-inc/espaze/internal/emulation/audio"
	"github.com/Duc-inc/espaze/internal/emulation/core"
	emuinput "github.com/Duc-inc/espaze/internal/emulation/input"
	"github.com/Duc-inc/espaze/internal/emulation/video"

	"github.com/Duc-inc/espaze/internal/systems/chip8/input"
	"github.com/Duc-inc/espaze/internal/systems/chip8/memory"
	"github.com/Duc-inc/espaze/internal/systems/chip8/rom"
	"github.com/Duc-inc/espaze/internal/systems/chip8/timer"
	"github.com/Duc-inc/espaze/internal/systems/schip/cpu"
	"github.com/Duc-inc/espaze/internal/systems/schip/display"
)

// systemID is how this core registers itself and how save states are tagged.
const systemID = "schip"

// cyclesPerFrame is higher than base CHIP-8's: Super-CHIP programs were
// written assuming the faster HP48 interpreter.
const cyclesPerFrame = 30

// sampleRate is the audio rate used for the synthesized beep tone.
const sampleRate = 44100

// Schip wires together the CPU, memory, display, keypad and timers into a
// core.Core implementation. It's built almost entirely out of CHIP-8's own
// components (memory, timer, keypad, ROM loader) - only the display
// (resolution switching) and CPU (extended opcode table) are new,
// demonstrating that a second system plugs into the same architecture
// without changing a single line of the first one.
type Schip struct {
	mem     *memory.Memory
	disp    *display.Display
	keys    *input.Keypad
	delay   *timer.Timer
	sound   *timer.Timer
	proc    *cpu.CPU
	audio   *emuaudio.Buffer
	beepPos int
}

// New constructs a fresh, unloaded Super-CHIP core. Matches core.Factory.
func New() core.Core {
	mem := memory.New()
	loadBigFont(mem)

	disp := display.New()
	keys := input.New()
	delay := timer.New()
	sound := timer.New()
	return &Schip{
		mem:   mem,
		disp:  disp,
		keys:  keys,
		delay: delay,
		sound: sound,
		proc:  cpu.New(mem, disp, keys, delay, sound),
		audio: emuaudio.NewBuffer(),
	}
}

// loadBigFont installs the 10-byte-per-digit font FX30 reads from, right
// alongside the small font that memory.New() already loaded.
func loadBigFont(mem *memory.Memory) {
	for i, b := range cpu.BigFontSet {
		mem.Write(cpu.BigFontStart+uint16(i), b)
	}
}

func init() {
	core.Register(Metadata(), New)
}

// Metadata describes the Super-CHIP core for the registry and scanner.
func Metadata() core.Metadata {
	return core.Metadata{
		ID:              systemID,
		Name:            "Super-CHIP",
		Extensions:      []string{".sc8", ".schip8"},
		FramesPerSecond: 60,
		ScreenWidth:     display.LowWidth,
		ScreenHeight:    display.LowHeight,
	}
}

// Metadata implements core.Core.
func (s *Schip) Metadata() core.Metadata { return Metadata() }

// LoadROM implements core.Core.
func (s *Schip) LoadROM(data []byte) error {
	s.Reset()
	return rom.Load(s.mem, data)
}

// Reset implements core.Core.
func (s *Schip) Reset() {
	s.mem.Reset()
	loadBigFont(s.mem)
	s.disp.SetExtended(false)
	s.disp.Clear()
	s.keys.Reset()
	s.delay.Set(0)
	s.sound.Set(0)
	s.proc.Reset()
}

// StepFrame implements core.Core.
func (s *Schip) StepFrame() error {
	for i := 0; i < cyclesPerFrame; i++ {
		if err := s.proc.Step(); err != nil {
			return err
		}
	}
	s.delay.Tick()
	s.sound.Tick()
	s.synthesizeAudio()
	return nil
}

// FrameBuffer implements core.Core, converting the active-resolution
// 1-bit display into RGBA. Size varies frame to frame with the display's
// resolution mode - callers must read Width/Height from each frame.
func (s *Schip) FrameBuffer() *video.FrameBuffer {
	width, height := s.disp.Width(), s.disp.Height()
	fb := video.NewFrameBuffer(width, height)
	pixels := s.disp.Pixels()
	for i, on := range pixels {
		x := i % width
		y := i / width
		if on {
			fb.SetPixel(x, y, 0xE6, 0xE6, 0xE6, 0xFF)
		} else {
			fb.SetPixel(x, y, 0x0B, 0x0B, 0x0B, 0xFF)
		}
	}
	return fb
}

// DrainAudio implements core.Core.
func (s *Schip) DrainAudio() ([]int16, int) {
	return s.audio.Drain(), sampleRate
}

// SetInput implements core.Core, mapping generic bits 0x0-0xF straight
// onto the 16-key hex keypad, same layout as base CHIP-8.
func (s *Schip) SetInput(state emuinput.State) {
	for key := uint8(0); key < 16; key++ {
		s.keys.Set(key, state.Pressed(key))
	}
}
