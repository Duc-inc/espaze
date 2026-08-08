package di

import "testing"

type fakeMem struct{ data [256]byte }

func (m *fakeMem) Write8(addr uint32, v byte) { m.data[addr] = v }

func TestReadSectorCommandDMAsFromImage(t *testing.T) {
	image := make([]byte, 0x1000)
	copy(image[0x40:], []byte{0xDE, 0xAD, 0xBE, 0xEF})

	mem := &fakeMem{}
	d := New(image, mem)

	d.Write32(regCMDBUF0, cmdReadSector<<24)
	d.Write32(regCMDBUF1, 0x40>>2) // offset in 32-bit words
	d.Write32(regCMDBUF2, 4)       // length
	d.Write32(regDILENGTH, 4)
	d.Write32(regDIMAR, 0x10)
	d.Write32(regDICR, 1) // TSTART

	if mem.data[0x10] != 0xDE || mem.data[0x13] != 0xEF {
		t.Fatalf("DMA'd bytes = %#x %#x, want 0xde 0xef", mem.data[0x10], mem.data[0x13])
	}
	if d.Read32(regDISR)&bitTCINT == 0 {
		t.Fatal("expected TCINT set after transfer")
	}
}

func TestWritingTCINTClearsIt(t *testing.T) {
	d := New(nil, &fakeMem{})
	d.tcint = true
	d.Write32(regDISR, bitTCINT)
	if d.Read32(regDISR)&bitTCINT != 0 {
		t.Fatal("expected TCINT cleared")
	}
}
