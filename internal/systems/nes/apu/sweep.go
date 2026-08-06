package apu

// sweep periodically adjusts a pulse channel's own timer period up or
// down, producing the pitch-bend effect used for slides. The two pulse
// channels compute their negated target period slightly differently -
// pulse 1 uses one's complement, pulse 2 two's complement - a real
// hardware quirk (ones bit is tracked via onesComplement below), not a
// bug to "fix".
type sweep struct {
	enabled        bool
	period         byte
	negate         bool
	shift          byte
	onesComplement bool // true for pulse 1, false for pulse 2

	reload  bool
	divider byte
}

func newSweep(onesComplement bool) sweep {
	return sweep{onesComplement: onesComplement}
}

// write handles $4001/$4005.
func (s *sweep) write(v byte) {
	s.enabled = v&0x80 != 0
	s.period = (v >> 4) & 0x07
	s.negate = v&0x08 != 0
	s.shift = v & 0x07
	s.reload = true
}

// target computes the new timer period the sweep unit wants, and
// whether that would mute the channel (period out of the valid range).
func (s *sweep) target(timerPeriod uint16) (uint16, bool) {
	change := timerPeriod >> s.shift
	var next int32
	if s.negate {
		if s.onesComplement {
			next = int32(timerPeriod) - int32(change) - 1
		} else {
			next = int32(timerPeriod) - int32(change)
		}
	} else {
		next = int32(timerPeriod) + int32(change)
	}
	if next < 0 {
		next = 0
	}
	muted := timerPeriod < 8 || next > 0x7FF
	return uint16(next), muted
}

// tick runs once per half-frame, applying the target period into
// timerPeriod when due (the divider reached zero, sweeping is enabled,
// and the shift is non-zero) and returns the possibly-updated period.
func (s *sweep) tick(timerPeriod uint16) uint16 {
	_, muted := s.target(timerPeriod)
	next := timerPeriod

	if s.divider == 0 && s.enabled && s.shift != 0 && !muted {
		newPeriod, _ := s.target(timerPeriod)
		next = newPeriod
	}
	if s.divider == 0 || s.reload {
		s.divider = s.period
		s.reload = false
	} else {
		s.divider--
	}
	return next
}
