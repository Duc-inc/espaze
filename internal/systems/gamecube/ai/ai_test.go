package ai

import "testing"

func TestWriteAICRSetsPlayingAndClearsInterrupt(t *testing.T) {
	a := New()
	a.Write32(regAICR, bitPSTAT)
	if !a.Playing() {
		t.Fatal("expected playing")
	}

	a.interrupt = true
	a.Write32(regAICR, bitPSTAT|bitAIINT)
	if a.Read32(regAICR)&bitAIINT != 0 {
		t.Fatal("expected interrupt cleared")
	}
}

func TestVolumeRoundTrip(t *testing.T) {
	a := New()
	a.Write32(regAIVR, uint32(200)<<8|100)
	l, r := a.Volume()
	if l != 100 || r != 200 {
		t.Fatalf("volume = (%d,%d), want (100,200)", l, r)
	}
}

func TestStepFiresInterruptAtConfiguredCount(t *testing.T) {
	a := New()
	a.Write32(regAICR, bitPSTAT)
	a.Write32(regAIIT, 3)

	var fired bool
	for i := 0; i < 3; i++ {
		fired = a.Step()
	}
	if !fired {
		t.Fatal("expected interrupt to fire at sample count 3")
	}
}

func TestStepDoesNothingWhileStopped(t *testing.T) {
	a := New()
	if a.Step() {
		t.Fatal("expected no interrupt while not playing")
	}
	if a.Read32(regAISCNT) != 0 {
		t.Fatal("expected sample counter to stay 0 while stopped")
	}
}
