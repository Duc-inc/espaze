package display

// Super-CHIP adds a 128x64 "extended" mode on top of the original CHIP-8
// 64x32 screen; programs switch between them at runtime (00FE/00FF).
const (
	LowWidth   = 64
	LowHeight  = 32
	HighWidth  = 128
	HighHeight = 64
)

// Display is a 1-bit-per-pixel framebuffer whose size changes with the
// active resolution mode, XOR-drawn like the original interpreter.
type Display struct {
	extended bool
	width    int
	height   int
	pixels   []bool
}

// New returns a cleared display, booted in low-res (64x32) mode.
func New() *Display {
	d := &Display{}
	d.setMode(false)
	return d
}

func (d *Display) setMode(extended bool) {
	d.extended = extended
	if extended {
		d.width, d.height = HighWidth, HighHeight
	} else {
		d.width, d.height = LowWidth, LowHeight
	}
	d.pixels = make([]bool, d.width*d.height)
}

// SetExtended switches resolution mode, clearing the screen if it changed.
func (d *Display) SetExtended(on bool) {
	if on != d.extended {
		d.setMode(on)
	}
}

// Extended reports whether the display is currently in 128x64 mode.
func (d *Display) Extended() bool { return d.extended }

// Width and Height report the active resolution.
func (d *Display) Width() int  { return d.width }
func (d *Display) Height() int { return d.height }

// Clear turns every pixel off without changing resolution (CLS).
func (d *Display) Clear() {
	d.pixels = make([]bool, d.width*d.height)
}

// DrawSprite XORs a sprite onto the screen at (x, y), wrapping at the
// edges. When wide is true the sprite is 16 pixels wide with 2 bytes per
// row (the SCHIP DXY0 mode); otherwise it's the classic 8-pixels-wide,
// 1-byte-per-row CHIP-8 sprite. Reports whether any pixel was turned off.
func (d *Display) DrawSprite(x, y int, sprite []byte, wide bool) bool {
	spriteWidth := 8
	bytesPerRow := 1
	if wide {
		spriteWidth = 16
		bytesPerRow = 2
	}

	collision := false
	rows := len(sprite) / bytesPerRow
	for row := 0; row < rows; row++ {
		bits := rowBits(sprite, row, bytesPerRow)
		py := (y + row) % d.height
		for col := 0; col < spriteWidth; col++ {
			if bits&(0x8000>>col) == 0 {
				continue
			}
			px := (x + col) % d.width
			idx := py*d.width + px
			if d.pixels[idx] {
				collision = true
			}
			d.pixels[idx] = !d.pixels[idx]
		}
	}
	return collision
}

func rowBits(sprite []byte, row, bytesPerRow int) uint16 {
	if bytesPerRow == 1 {
		return uint16(sprite[row]) << 8
	}
	return uint16(sprite[row*2])<<8 | uint16(sprite[row*2+1])
}

// Pixels returns a copy of the active on/off grid, row-major.
func (d *Display) Pixels() []bool {
	out := make([]bool, len(d.pixels))
	copy(out, d.pixels)
	return out
}

// Restore overwrites the display with a previously captured snapshot,
// switching resolution to match if needed.
func (d *Display) Restore(extended bool, width, height int, pixels []bool) {
	d.extended = extended
	d.width = width
	d.height = height
	d.pixels = make([]bool, len(pixels))
	copy(d.pixels, pixels)
}
