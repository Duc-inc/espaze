package apu

// envelope is the volume unit shared by both pulse channels and noise:
// either a fixed ("constant") volume, or a decaying one driven by a
// divider that ticks once per quarter-frame.
type envelope struct {
	period         byte // also doubles as the constant-volume level
	constantVolume bool
	loop           bool // also the channel's length-counter halt flag

	start      bool
	divider    byte
	decayLevel byte
}

// writeControl handles a write to $4000/$4004/$400C: bit4 selects
// constant volume, bit5 is loop/halt, bits0-3 are the period (or the
// constant volume level, sharing the same bits).
func (e *envelope) writeControl(v byte) {
	e.period = v & 0x0F
	e.constantVolume = v&0x10 != 0
	e.loop = v&0x20 != 0
}

// restart is triggered by writing the channel's length-load register
// ($4003/$4007/$400F).
func (e *envelope) restart() { e.start = true }

func (e *envelope) tick() {
	if e.start {
		e.start = false
		e.decayLevel = 15
		e.divider = e.period
		return
	}
	if e.divider > 0 {
		e.divider--
		return
	}
	e.divider = e.period
	if e.decayLevel > 0 {
		e.decayLevel--
	} else if e.loop {
		e.decayLevel = 15
	}
}

func (e *envelope) volume() byte {
	if e.constantVolume {
		return e.period
	}
	return e.decayLevel
}
