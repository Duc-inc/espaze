package apu

// lengthCounter implements the "auto-mute after N ticks" feature every
// channel has: NRx1 loads how many 256Hz ticks the channel should play
// for, and when it reaches zero the channel silences itself.
type lengthCounter struct {
	counter int
	max     int
	enabled bool
}

func newLengthCounter(max int) lengthCounter {
	return lengthCounter{max: max}
}

// setFromRegister loads the counter from the bits written to NRx1.
func (l *lengthCounter) setFromRegister(bits int) {
	l.counter = l.max - bits
}

// tick advances the counter at 256Hz; returns true the instant it hits
// zero, telling the caller to silence the channel.
func (l *lengthCounter) tick() bool {
	if !l.enabled || l.counter <= 0 {
		return false
	}
	l.counter--
	return l.counter == 0
}
