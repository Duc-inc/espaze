package timer

import "sync/atomic"

// Timer models one of CHIP-8's two 8-bit counters (delay and sound), both
// of which decrement at a fixed 60Hz whenever they hold a nonzero value.
type Timer struct {
	value atomic.Uint32
}

// New returns a timer at zero.
func New() *Timer {
	return &Timer{}
}

// Set loads the timer with a new value (the LD DT/ST instructions).
func (t *Timer) Set(v uint8) {
	t.value.Store(uint32(v))
}

// Get reads the current value (the LD Vx, DT instruction).
func (t *Timer) Get() uint8 {
	return uint8(t.value.Load())
}

// Tick decrements the timer by one if it is above zero. Call at 60Hz.
func (t *Timer) Tick() {
	for {
		cur := t.value.Load()
		if cur == 0 {
			return
		}
		if t.value.CompareAndSwap(cur, cur-1) {
			return
		}
	}
}

// Active reports whether the timer is still counting down.
func (t *Timer) Active() bool {
	return t.value.Load() > 0
}
