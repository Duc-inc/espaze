package display

// scrollStep is how many columns 00FB/00FC shift, fixed by the SCHIP spec.
const scrollStep = 4

// ScrollDown shifts every pixel down by n rows, filling vacated rows at
// the top with black (the 00CN instruction).
func (d *Display) ScrollDown(n int) {
	next := make([]bool, len(d.pixels))
	for y := 0; y < d.height; y++ {
		srcY := y - n
		if srcY < 0 {
			continue
		}
		copy(next[y*d.width:(y+1)*d.width], d.pixels[srcY*d.width:(srcY+1)*d.width])
	}
	d.pixels = next
}

// ScrollLeft shifts every pixel left by scrollStep columns (00FC).
func (d *Display) ScrollLeft() {
	next := make([]bool, len(d.pixels))
	for y := 0; y < d.height; y++ {
		for x := 0; x < d.width; x++ {
			srcX := x + scrollStep
			if srcX >= d.width {
				continue
			}
			next[y*d.width+x] = d.pixels[y*d.width+srcX]
		}
	}
	d.pixels = next
}

// ScrollRight shifts every pixel right by scrollStep columns (00FB).
func (d *Display) ScrollRight() {
	next := make([]bool, len(d.pixels))
	for y := 0; y < d.height; y++ {
		for x := 0; x < d.width; x++ {
			srcX := x - scrollStep
			if srcX < 0 {
				continue
			}
			next[y*d.width+x] = d.pixels[y*d.width+srcX]
		}
	}
	d.pixels = next
}
