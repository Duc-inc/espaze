package cpu

// irqTimer models the interrupt controller and periodic timer that
// live on the HuC6280 die itself (not a separate chip): a 3-bit IRQ
// mask ($1402) covering IRQ2/VDC, IRQ1/external, and the built-in
// timer, an acknowledge/status port ($1403), and the timer itself
// ($0C00 reload value, $0C01 start/stop), which decrements at a fixed
// divisor of the CPU clock and fires periodically while running.
type irqTimer struct {
	maskIRQ2, maskIRQ1, maskTimer bool

	pendingIRQ2, pendingIRQ1, pendingTimer bool

	timerReload  byte
	timerCounter int
	timerRunning bool
	timerAccum   int
}

const timerDivisor = 1024 // cycles per timer decrement

func (t *irqTimer) writeMask(v byte) {
	t.maskIRQ2 = v&0x01 != 0
	t.maskIRQ1 = v&0x02 != 0
	t.maskTimer = v&0x04 != 0
}

func (t *irqTimer) readStatus() byte {
	var v byte
	if t.pendingIRQ2 {
		v |= 0x01
	}
	if t.pendingIRQ1 {
		v |= 0x02
	}
	if t.pendingTimer {
		v |= 0x04
	}
	return v
}

func (t *irqTimer) acknowledgeTimer() { t.pendingTimer = false }

func (t *irqTimer) writeTimerReload(v byte) { t.timerReload = v & 0x7F }

func (t *irqTimer) writeTimerControl(v byte) {
	t.timerRunning = v&0x01 != 0
	if t.timerRunning {
		t.timerCounter = int(t.timerReload)
		t.timerAccum = 0
	}
}

// TriggerIRQ2/TriggerIRQ1 latch the VDC's and any external device's
// interrupt lines.
func (t *irqTimer) TriggerIRQ2() { t.pendingIRQ2 = true }
func (t *irqTimer) TriggerIRQ1() { t.pendingIRQ1 = true }

func (t *irqTimer) step(cycles int) {
	if !t.timerRunning {
		return
	}
	t.timerAccum += cycles
	for t.timerAccum >= timerDivisor {
		t.timerAccum -= timerDivisor
		t.timerCounter--
		if t.timerCounter < 0 {
			t.timerCounter = int(t.timerReload)
			t.pendingTimer = true
		}
	}
}

// pendingVector returns the vector address of the highest-priority
// unmasked pending interrupt, and whether one exists. Real hardware's
// exact priority among IRQ1/IRQ2/timer is a minor detail this project
// hasn't confirmed against documentation; IRQ2 (the VDC, by far the
// most common source) is checked first.
func (t *irqTimer) pendingVector() (uint16, bool) {
	switch {
	case t.pendingIRQ2 && !t.maskIRQ2:
		return 0xFFF6, true
	case t.pendingIRQ1 && !t.maskIRQ1:
		return 0xFFF8, true
	case t.pendingTimer && !t.maskTimer:
		return 0xFFFA, true
	default:
		return 0, false
	}
}
