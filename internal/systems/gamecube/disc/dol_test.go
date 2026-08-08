package disc

import "testing"

type testMem struct {
	written map[uint32]byte
}

func (m *testMem) Write8(addr uint32, v byte) {
	if m.written == nil {
		m.written = make(map[uint32]byte)
	}
	m.written[addr] = v
}

func newTestDOL() []byte {
	img := make([]byte, 0x200)
	// One text section: file offset 0x100, load address 0x80003000, size 4.
	putU32(img, dolOffsetsBase, 0x100)
	putU32(img, dolAddrsBase, 0x80003000)
	putU32(img, dolSizesBase, 4)
	putU32(img, dolEntryOffset, 0x80003000)
	copy(img[0x100:], []byte{0xAA, 0xBB, 0xCC, 0xDD})
	return img
}

func putU32(b []byte, offset int, v uint32) {
	b[offset] = byte(v >> 24)
	b[offset+1] = byte(v >> 16)
	b[offset+2] = byte(v >> 8)
	b[offset+3] = byte(v)
}

func TestParseDOLReadsEntryAndSections(t *testing.T) {
	img := newTestDOL()
	d := ParseDOL(img)
	if d.Entry != 0x80003000 {
		t.Fatalf("Entry = %#08x, want 0x80003000", d.Entry)
	}
	if len(d.Sections) != 1 {
		t.Fatalf("got %d sections, want 1 (empty-size sections should be skipped)", len(d.Sections))
	}
	if d.Sections[0].LoadAddr != 0x80003000 || d.Sections[0].Size != 4 {
		t.Fatalf("section = %+v, want addr=0x80003000 size=4", d.Sections[0])
	}
}

func TestLoadIntoCopiesSectionBytes(t *testing.T) {
	img := newTestDOL()
	d := ParseDOL(img)
	mem := &testMem{}
	d.LoadInto(img, mem)

	want := []byte{0xAA, 0xBB, 0xCC, 0xDD}
	for i, w := range want {
		if mem.written[0x80003000+uint32(i)] != w {
			t.Fatalf("mem[%#08x] = %#02x, want %#02x", 0x80003000+i, mem.written[0x80003000+uint32(i)], w)
		}
	}
}
