package apu

// lengthTable maps a 5-bit length-load value to the actual number of
// APU frames (half-frame ticks) a channel plays for - a fixed lookup
// table baked into real hardware, not a formula.
var lengthTable = [32]byte{
	10, 254, 20, 2, 40, 4, 80, 6, 160, 8, 60, 10, 14, 12, 26, 14,
	12, 16, 24, 18, 48, 20, 96, 22, 192, 24, 72, 26, 16, 28, 32, 30,
}

// lengthCounter silences a channel once it counts down to zero, unless
// the channel's halt flag (shared with envelope loop, on pulse/noise;
// its own "control" flag on triangle) keeps it running forever.
type lengthCounter struct {
	value byte
	halt  bool
}

func (l *lengthCounter) load(index byte) { l.value = lengthTable[index&0x1F] }

func (l *lengthCounter) tick() {
	if l.value > 0 && !l.halt {
		l.value--
	}
}

func (l *lengthCounter) active() bool { return l.value > 0 }
