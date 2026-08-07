package tia

// playfield holds the 20-bit pattern (PF0/PF1/PF2, reassembled into
// hardware bit order) spanning the left half of the screen, mirrored
// or repeated across the right half depending on CTRLPF.
type playfield struct {
	pf0, pf1, pf2 byte
	reflect       bool // CTRLPF bit0
	score         bool // CTRLPF bit1: left/right halves use COLUP0/COLUP1
	priority      bool // CTRLPF bit2: playfield/ball drawn above players/missiles
}

func (p *playfield) writeCTRLPF(v byte) {
	p.reflect = v&0x01 != 0
	p.score = v&0x02 != 0
	p.priority = v&0x04 != 0
}

// bit returns the playfield's 20-bit pattern bit at index 0-19, in the
// left-to-right screen order real hardware actually shifts it out in:
// PF0's 4 graphics bits are its upper nibble, read most-significant
// first; PF1 reads MSB-first; PF2 reads LSB-first.
func (p *playfield) bit(index int) bool {
	switch {
	case index < 4:
		return p.pf0&(0x10<<uint(index)) != 0
	case index < 12:
		i := index - 4
		return p.pf1&(0x80>>uint(i)) != 0
	default:
		i := index - 12
		return p.pf2&(0x01<<uint(i)) != 0
	}
}

// pixelAt reports whether the playfield is set at screen x (0-159);
// each of the 20 pattern bits covers 4 pixels.
func (p *playfield) pixelAt(x int) bool {
	if x < 80 {
		return p.bit(x / 4)
	}
	right := x - 80
	if p.reflect {
		return p.bit(19 - right/4)
	}
	return p.bit(right / 4)
}
