package memory

// timer is one of the GBA's 4 16-bit up-counters: it either free-runs
// off a prescaled CPU clock, or (timers 1-3 only) increments once each
// time the previous timer overflows ("count-up"/cascade mode).
type timer struct {
	counter, reload uint16
	prescaler       int // cycles per increment: 1, 64, 256, or 1024
	cascade         bool
	irqEnable       bool
	running         bool

	sub int
}

var prescalerCycles = [4]int{1, 64, 256, 1024}

func (t *timer) writeReload(v uint16) { t.reload = v }

func (t *timer) writeControl(v uint16) {
	t.prescaler = prescalerCycles[v&0x03]
	t.cascade = v&0x04 != 0
	t.irqEnable = v&0x40 != 0
	startingNow := v&0x80 != 0 && !t.running
	t.running = v&0x80 != 0
	if startingNow {
		t.counter = t.reload
	}
}

// tick advances by one cycle (ordinary mode) or one cascade pulse,
// returning whether it just overflowed.
func (t *timer) tick() bool {
	if !t.running {
		return false
	}
	t.sub++
	if t.sub < t.prescaler {
		return false
	}
	t.sub = 0
	return t.increment()
}

func (t *timer) increment() bool {
	t.counter++
	if t.counter != 0 {
		return false
	}
	t.counter = t.reload
	return true
}

// timers holds all 4 and the shared IRQ controller they raise flags on.
type timers struct {
	t   [4]timer
	irq *interrupts
}

var timerIRQBits = [4]uint16{IFTimer0, IFTimer1, IFTimer2, IFTimer3}

func (ts *timers) writeReload(index int, v uint16)  { ts.t[index].writeReload(v) }
func (ts *timers) writeControl(index int, v uint16) { ts.t[index].writeControl(v) }
func (ts *timers) readCounter(index int) uint16     { return ts.t[index].counter }

func (ts *timers) step(cpuCycles int) {
	for i := 0; i < cpuCycles; i++ {
		overflowed := false
		for n := 0; n < 4; n++ {
			t := &ts.t[n]
			var of bool
			if n > 0 && t.cascade {
				if t.running && overflowed {
					of = t.increment()
				}
			} else {
				of = t.tick()
			}
			if of && t.irqEnable {
				ts.irq.raise(timerIRQBits[n])
			}
			overflowed = of
		}
	}
}
