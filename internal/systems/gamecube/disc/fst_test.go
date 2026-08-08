package disc

import "testing"

func putBE32(img []byte, off, v uint32) {
	img[off] = byte(v >> 24)
	img[off+1] = byte(v >> 16)
	img[off+2] = byte(v >> 8)
	img[off+3] = byte(v)
}

func putFSTEntry(img []byte, off uint32, typ byte, nameOffset, offsetOrParent, lengthOrNext uint32) {
	img[off] = typ
	img[off+1] = byte(nameOffset >> 16)
	img[off+2] = byte(nameOffset >> 8)
	img[off+3] = byte(nameOffset)
	putBE32(img, off+4, offsetOrParent)
	putBE32(img, off+8, lengthOrNext)
}

func newTestImageWithFST() (img []byte, fstBase uint32) {
	img = append(newTestImage(), make([]byte, 0x2000)...)
	fstBase = 0x1000
	putBE32(img, fstOffsetAddr, fstBase)
	putBE32(img, fstSizeAddr, 200)

	const entryCount = 3
	putFSTEntry(img, fstBase+0*entryBytes, byte(EntryDir), 0, 0, entryCount) // root
	putFSTEntry(img, fstBase+1*entryBytes, byte(EntryFile), 0, 0x3000, 42)   // readme.txt
	putFSTEntry(img, fstBase+2*entryBytes, byte(EntryFile), 11, 0x4000, 99) // data.bin

	stringTable := fstBase + entryCount*entryBytes
	copy(img[stringTable:], "readme.txt\x00data.bin\x00")
	return img, fstBase
}

func TestParseFSTReadsEntries(t *testing.T) {
	img, _ := newTestImageWithFST()
	h, err := ParseHeader(img)
	if err != nil {
		t.Fatalf("ParseHeader: %v", err)
	}

	fst, err := ParseFST(img, h)
	if err != nil {
		t.Fatalf("ParseFST: %v", err)
	}
	if len(fst.Entries) != 3 {
		t.Fatalf("got %d entries, want 3", len(fst.Entries))
	}

	root := fst.Entries[0]
	if root.Type != EntryDir {
		t.Fatalf("root type = %v, want EntryDir", root.Type)
	}

	readme := fst.Entries[1]
	if readme.Type != EntryFile || readme.Name != "readme.txt" || readme.FileOffset != 0x3000 || readme.FileLength != 42 {
		t.Fatalf("readme entry = %+v, want {File readme.txt 0x3000 42}", readme)
	}

	data := fst.Entries[2]
	if data.Name != "data.bin" || data.FileOffset != 0x4000 || data.FileLength != 99 {
		t.Fatalf("data entry = %+v, want {File data.bin 0x4000 99}", data)
	}
}

func TestFSTLookupFindsEntryByName(t *testing.T) {
	img, _ := newTestImageWithFST()
	h, _ := ParseHeader(img)
	fst, err := ParseFST(img, h)
	if err != nil {
		t.Fatalf("ParseFST: %v", err)
	}

	e, ok := fst.Lookup("data.bin")
	if !ok {
		t.Fatal("expected to find data.bin")
	}
	if e.FileOffset != 0x4000 {
		t.Fatalf("FileOffset = %#x, want 0x4000", e.FileOffset)
	}

	if _, ok := fst.Lookup("missing.txt"); ok {
		t.Fatal("expected not to find missing.txt")
	}
}

func TestParseFSTRejectsOutOfRangeOffset(t *testing.T) {
	img := newTestImage()
	putBE32(img, fstOffsetAddr, uint32(len(img)+100))
	h, _ := ParseHeader(img)
	if _, err := ParseFST(img, h); err == nil {
		t.Fatal("expected an error for an FST offset beyond the image")
	}
}
