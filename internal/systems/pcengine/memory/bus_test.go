package memory

import (
	"testing"

	"github.com/Duc-inc/espaze/internal/emulation/input"
)

type fakeVDC struct{ selected, lastLo, lastHi byte }

func (f *fakeVDC) SelectRegister(reg byte) { f.selected = reg }
func (f *fakeVDC) WriteDataLow(b byte)     { f.lastLo = b }
func (f *fakeVDC) WriteDataHigh(b byte)    { f.lastHi = b }
func (f *fakeVDC) ReadDataLow() byte       { return 0x11 }
func (f *fakeVDC) ReadDataHigh() byte      { return 0x22 }

type fakeVCE struct{ addrLo byte }

func (f *fakeVCE) WriteAddressLow(b byte)  { f.addrLo = b }
func (f *fakeVCE) WriteAddressHigh(b byte) {}
func (f *fakeVCE) WriteDataLow(b byte)     {}
func (f *fakeVCE) WriteDataHigh(b byte)    {}

type fakePSG struct{ selectedChannel byte }

func (f *fakePSG) SelectChannel(v byte)        { f.selectedChannel = v }
func (f *fakePSG) WriteMainVolumeLeft(v byte)  {}
func (f *fakePSG) WriteMainVolumeRight(v byte) {}
func (f *fakePSG) WriteFreqLow(v byte)         {}
func (f *fakePSG) WriteFreqHigh(v byte)        {}
func (f *fakePSG) WriteControl(v byte)         {}
func (f *fakePSG) WritePan(v byte)             {}
func (f *fakePSG) WriteWaveData(v byte)        {}
func (f *fakePSG) WriteNoiseControl(v byte)    {}

type fakeTimerIRQ struct{ mask byte }

func (f *fakeTimerIRQ) WriteTimerReload(v byte)  {}
func (f *fakeTimerIRQ) WriteTimerControl(v byte) {}
func (f *fakeTimerIRQ) WriteIRQMask(v byte)      { f.mask = v }
func (f *fakeTimerIRQ) ReadIRQStatus() byte      { return 0x05 }

func newTestBus() (*Bus, *fakeVDC, *fakeTimerIRQ) {
	vdc := &fakeVDC{}
	tirq := &fakeTimerIRQ{}
	rom := make([]byte, 0x1000)
	rom[0] = 0xAB
	b := New(rom, vdc, &fakeVCE{}, &fakePSG{}, tirq)
	return b, vdc, tirq
}

func TestCartridgeReadThroughBus(t *testing.T) {
	b, _, _ := newTestBus()
	if v := b.Read(0); v != 0xAB {
		t.Fatalf("Read(0) = %#02x, want 0xAB", v)
	}
}

func TestWorkRAMReadWriteRoundTrip(t *testing.T) {
	b, _, _ := newTestBus()
	b.Write(0x1F0010, 0x42)
	if v := b.Read(0x1F0010); v != 0x42 {
		t.Fatalf("Read(0x1F0010) = %#02x, want 0x42", v)
	}
}

func TestVDCRegisterWritesDelegate(t *testing.T) {
	b, vdc, _ := newTestBus()
	b.Write(0x1FE000, 0x05)
	b.Write(0x1FE002, 0x99)
	if vdc.selected != 0x05 || vdc.lastLo != 0x99 {
		t.Fatalf("VDC selected=%#02x lastLo=%#02x, want 0x05/0x99", vdc.selected, vdc.lastLo)
	}
}

func TestIRQMaskWriteDelegates(t *testing.T) {
	b, _, tirq := newTestBus()
	b.Write(0x1FF402, 0x07)
	if tirq.mask != 0x07 {
		t.Fatalf("IRQ mask = %#02x, want 0x07", tirq.mask)
	}
}

func TestControllerReadReflectsSELLine(t *testing.T) {
	b, _, _ := newTestBus()
	b.SetButtons(input.State{}.With(Up, true))

	b.Write(0x1FF000, 0x01) // SEL=1: d-pad
	v := b.Read(0x1FF000)
	if v&0x01 != 0 {
		t.Fatal("Up should read active-low when SEL selects the d-pad")
	}

	b.Write(0x1FF000, 0x00) // SEL=0: buttons
	v = b.Read(0x1FF000)
	if v&0x01 == 0 {
		t.Fatal("Button I bit should read released with only Up held")
	}
}
