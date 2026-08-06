package display

// Width and Height are the native CHIP-8 monochrome screen dimensions.
const (
	Width  = 64
	Height = 32
)

// Display is a 1-bit-per-pixel framebuffer, XOR-drawn exactly like the
// original CHIP-8 interpreter did for its sprites.
type Display struct {
	pixels [Width * Height]bool
}

// New returns a cleared display.
func New() *Display {
	return &Display{}
}

// Clear turns every pixel off (the CLS instruction).
func (d *Display) Clear() {
	d.pixels = [Width * Height]bool{}
}

// DrawSprite XORs an 8-pixel-wide, len(sprite)-tall sprite onto the screen
// at (x, y), wrapping at the edges, and reports whether any pixel that was
// on got turned off (used to set the VF collision flag).
func (d *Display) DrawSprite(x, y int, sprite []byte) bool {
	collision := false
	for row, bits := range sprite {
		py := (y + row) % Height
		for col := 0; col < 8; col++ {
			if bits&(0x80>>col) == 0 {
				continue
			}
			px := (x + col) % Width
			idx := py*Width + px
			if d.pixels[idx] {
				collision = true
			}
			d.pixels[idx] = !d.pixels[idx]
		}
	}
	return collision
}

// Pixels exposes the raw on/off grid, row-major, for rendering or save states.
func (d *Display) Pixels() [Width * Height]bool {
	return d.pixels
}

// Restore overwrites the grid with a previously captured snapshot.
func (d *Display) Restore(snapshot [Width * Height]bool) {
	d.pixels = snapshot
}
