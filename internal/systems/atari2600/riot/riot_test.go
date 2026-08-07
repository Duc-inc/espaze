package riot

import (
	"testing"

	"github.com/Duc-inc/espaze/internal/emulation/input"
)

func TestRAMReadWriteRoundTrip(t *testing.T) {
	r := New()
	r.WriteRAM(0x10, 0x42)
	if v := r.ReadRAM(0x10); v != 0x42 {
		t.Fatalf("ReadRAM = %#02x, want 0x42", v)
	}
}

func TestTimerCountsDownAtSelectedInterval(t *testing.T) {
	r := New()
	r.WriteTimer(0x00, 5) // TIM1T: decrements once per cycle

	r.Step(3)
	if v := r.ReadINTIM(); v != 2 {
		t.Fatalf("INTIM after 3 cycles at TIM1T = %d, want 2", v)
	}
}

func TestTimer1024TDecrementsSlowly(t *testing.T) {
	r := New()
	r.WriteTimer(0x03, 3) // T1024T: decrements once per 1024 cycles

	r.Step(100)
	if v := r.ReadINTIM(); v != 3 {
		t.Fatalf("INTIM after 100 cycles at T1024T = %d, want unchanged 3", v)
	}
	r.Step(1024)
	if v := r.ReadINTIM(); v != 2 {
		t.Fatalf("INTIM after a further 1024 cycles = %d, want 2", v)
	}
}

func TestSWCHAReflectsHeldDirections(t *testing.T) {
	r := New()
	r.SetButtons(input.State{}.With(Up, true).With(Right, true))

	v := r.ReadSWCHA()
	if v&0x10 != 0 {
		t.Fatal("Up should read active-low (bit4 clear) when held")
	}
	if v&0x80 != 0 {
		t.Fatal("Right should read active-low (bit7 clear) when held")
	}
	if v&0x20 == 0 {
		t.Fatal("Down should read released (bit5 set)")
	}
}
