package apu

// frequencySweep periodically shifts channel 1's frequency up or down,
// driven at 128Hz by the frame sequencer. Channels 2-4 don't have one.
type frequencySweep struct {
	period     byte
	negate     bool
	shift      byte
	timer      byte
	enabled    bool
	shadowFreq uint16
}

func (s *frequencySweep) writeRegister(v byte) {
	s.period = (v >> 4) & 0x07
	s.negate = v&0x08 != 0
	s.shift = v & 0x07
}

// trigger latches the current frequency and reports whether the very
// first sweep calculation already overflows (which disables the channel
// immediately, before it makes a sound).
func (s *frequencySweep) trigger(currentFreq uint16) (overflow bool) {
	s.shadowFreq = currentFreq
	s.timer = s.periodOrEight()
	s.enabled = s.period != 0 || s.shift != 0
	if s.shift == 0 {
		return false
	}
	_, overflow = s.calculate()
	return overflow
}

func (s *frequencySweep) periodOrEight() byte {
	if s.period == 0 {
		return 8
	}
	return s.period
}

func (s *frequencySweep) calculate() (newFreq uint16, overflow bool) {
	delta := s.shadowFreq >> s.shift
	if s.negate {
		newFreq = s.shadowFreq - delta
	} else {
		newFreq = s.shadowFreq + delta
	}
	return newFreq, newFreq > 2047
}

// tick returns the frequency to apply (if any) and whether the channel
// must be disabled (a sweep step overflowed past the 11-bit frequency).
func (s *frequencySweep) tick() (newFreq uint16, apply bool, disable bool) {
	if !s.enabled || s.period == 0 {
		return 0, false, false
	}
	if s.timer > 0 {
		s.timer--
	}
	if s.timer != 0 {
		return 0, false, false
	}
	s.timer = s.periodOrEight()

	freq, overflow := s.calculate()
	if overflow {
		return 0, false, true
	}
	if s.shift == 0 {
		return 0, false, false
	}

	s.shadowFreq = freq
	if _, overflow2 := s.calculate(); overflow2 {
		return freq, true, true
	}
	return freq, true, false
}
