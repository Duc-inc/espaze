package timer

// Timer models DIV/TIMA/TMA/TAC (0xFF04-0xFF07). DIV is the visible upper
// byte of an internal 16-bit counter that ticks every T-cycle; TIMA ticks
// at whichever rate TAC selects and requests an interrupt on overflow.
type Timer struct {
	div     uint16
	tima    byte
	tma     byte
	tac     byte
	timaAcc int
}

// New returns a timer in its post-boot state (all registers zero).
func New() *Timer {
	return &Timer{}
}

// Reset returns the timer to its post-boot state.
func (t *Timer) Reset() {
	*t = Timer{}
}

// Step advances the timer by tcycles T-cycles and reports whether TIMA
// overflowed (a Timer interrupt should be requested).
func (t *Timer) Step(tcycles int) bool {
	t.div += uint16(tcycles)

	if t.tac&0x04 == 0 {
		return false
	}

	interrupt := false
	threshold := t.threshold()
	t.timaAcc += tcycles
	for t.timaAcc >= threshold {
		t.timaAcc -= threshold
		t.tima++
		if t.tima == 0 {
			t.tima = t.tma
			interrupt = true
		}
	}
	return interrupt
}

func (t *Timer) threshold() int {
	switch t.tac & 0x03 {
	case 0:
		return 1024
	case 1:
		return 16
	case 2:
		return 64
	default:
		return 256
	}
}

// ReadRegister implements the CPU reading 0xFF04-0xFF07.
func (t *Timer) ReadRegister(addr uint16) byte {
	switch addr {
	case 0xFF04:
		return byte(t.div >> 8)
	case 0xFF05:
		return t.tima
	case 0xFF06:
		return t.tma
	default: // 0xFF07
		return t.tac | 0xF8
	}
}

// WriteRegister implements the CPU writing 0xFF04-0xFF07.
func (t *Timer) WriteRegister(addr uint16, v byte) {
	switch addr {
	case 0xFF04:
		t.div = 0
		t.timaAcc = 0
	case 0xFF05:
		t.tima = v
	case 0xFF06:
		t.tma = v
	default: // 0xFF07
		t.tac = v & 0x07
	}
}

// Snapshot/Restore persist timer state across save states.
type Snapshot struct {
	Div     uint16
	Tima    byte
	Tma     byte
	Tac     byte
	TimaAcc int
}

func (t *Timer) Snapshot() Snapshot {
	return Snapshot{Div: t.div, Tima: t.tima, Tma: t.tma, Tac: t.tac, TimaAcc: t.timaAcc}
}

func (t *Timer) Restore(s Snapshot) {
	t.div, t.tima, t.tma, t.tac, t.timaAcc = s.Div, s.Tima, s.Tma, s.Tac, s.TimaAcc
}
