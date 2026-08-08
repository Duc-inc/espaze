// Package vi implements the GameCube's Video Interface: the real
// memory-mapped registers a game reads/writes to configure the
// framebuffer address and get notified of VBlank, instead of this
// project's other cores' higher-level "just call SetPixel" approach.
// Register addresses and bit layouts come from a public hardware
// register reference (YAGCD chapter 5, "VI - Video Interface"),
// extracted and restated in this project's own words - timing values
// (sync pulse widths etc.) that don't affect observable behavior at
// this project's level of emulation aren't modeled.
package vi

const (
	Base = 0xCC002000
	Size = 0x100

	regDCR  = 0x02 // Display Configuration: enable, reset, format, interlace
	regTFBL = 0x1C // Top Field framebuffer base address
	regBFBL = 0x24 // Bottom Field framebuffer base address
	regDPV  = 0x2C // current vertical raster position (read-only)
	regDI0  = 0x30 // Display Interrupt 0-3, 4 bytes apart
)

const numDisplayInterrupts = 4

// VI holds the Video Interface's register state.
type VI struct {
	enabled bool
	format  byte // 0=NTSC,1=PAL,2=MPAL,3=Debug

	topFieldAddr    uint32
	bottomFieldAddr uint32

	vpos uint32 // current raster line, DPV's own counting scheme

	interrupts [numDisplayInterrupts]displayInterrupt
}

type displayInterrupt struct {
	active  bool
	enabled bool
	line    uint32
}

// New returns a VI with every register zeroed.
func New() *VI { return &VI{} }

// FramebufferAddr returns the top field's current framebuffer
// address - what a real game's own render loop sets via TFBL and the
// display hardware reads from continuously.
func (v *VI) FramebufferAddr() uint32 { return v.topFieldAddr }

// Enabled reports whether video timing generation is on (DCR's ENB bit).
func (v *VI) Enabled() bool { return v.enabled }

// AnyInterruptActive reports whether any of VI's 4 display interrupts
// is currently active (its status bit set, not yet cleared by a game
// writing DI0-3 back) - the level-triggered cause signal pi.PI's VI
// bit reports (see gamecube.go's Step).
func (v *VI) AnyInterruptActive() bool {
	for _, d := range v.interrupts {
		if d.active {
			return true
		}
	}
	return false
}

// Read32 reads one VI register at a block-relative offset.
func (v *VI) Read32(offset uint32) uint32 {
	switch {
	case offset == regDPV:
		return v.vpos
	case offset >= regDI0 && offset < regDI0+4*numDisplayInterrupts:
		return v.readDI(int((offset - regDI0) / 4))
	default:
		return 0
	}
}

// Write32 writes one VI register at a block-relative offset.
func (v *VI) Write32(offset uint32, val uint32) {
	switch {
	case offset == regDCR:
		v.enabled = val&1 != 0
		v.format = byte((val >> 8) & 0x3)
	case offset == regTFBL:
		v.topFieldAddr = val & 0x00FFFE00
	case offset == regBFBL:
		v.bottomFieldAddr = val & 0x00FFFE00
	case offset >= regDI0 && offset < regDI0+4*numDisplayInterrupts:
		v.writeDI(int((offset-regDI0)/4), val)
	}
}

func (v *VI) readDI(i int) uint32 {
	d := v.interrupts[i]
	word := d.line & 0x3FF
	if d.enabled {
		word |= 1 << 28
	}
	if d.active {
		word |= 1 << 31
	}
	return word
}

func (v *VI) writeDI(i int, val uint32) {
	d := &v.interrupts[i]
	d.line = val & 0x3FF
	d.enabled = val&(1<<28) != 0
	if val&(1<<31) == 0 {
		d.active = false // real hardware clears INT by writing it back as 0
	}
}

// linesPerFrame is this project's own simplified frame length (NTSC's
// own documented total field-line count), used only to wrap Step's
// raster position - not a claim about exact per-line timing.
const linesPerFrame = 525

// Step advances the raster position by one line and reports whether
// any enabled display interrupt (real hardware's VBlank notification
// mechanism) just fired.
func (v *VI) Step() (interrupted bool) {
	v.vpos++
	if v.vpos >= linesPerFrame {
		v.vpos = 1
	}
	for i := range v.interrupts {
		d := &v.interrupts[i]
		if d.enabled && d.line == v.vpos {
			d.active = true
			interrupted = true
		}
	}
	return interrupted
}
