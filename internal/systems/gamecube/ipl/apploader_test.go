package ipl

import (
	"testing"

	"github.com/Duc-inc/espaze/internal/systems/gamecube/memory"
	"github.com/Duc-inc/espaze/internal/systems/powerpc"
)

// newTestImageWithApploader builds a minimal disc image with a valid
// header and an apploader header/code blob at the fixed offset.
func newTestImageWithApploader() []byte {
	const codeSize = 8
	img := make([]byte, apploaderCodeOffset+codeSize)
	copy(img[0:6], []byte("GTEST"))
	putU32(img, 0x1C, 0xC2339F3D) // magic
	copy(img[0x20:], []byte("Apploader Test\x00"))
	putU32(img, 0x420, 0x1000) // DOL offset, unused by BootViaApploader

	h := img[apploaderHeaderOffset:]
	putU32(h, 0x10, apploaderLoadAddr) // entry
	putU32(h, 0x14, codeSize)          // size
	putU32(h, 0x18, 0)                 // trailer size

	// A recognizable instruction as the apploader's "code": addi r5,r0,9
	instr := uint32(14)<<26 | 5<<21 | 0<<16 | 9
	putU32(img, apploaderCodeOffset, instr)

	return img
}

func TestBootViaApploaderLoadsCodeAndSetsEntry(t *testing.T) {
	mem := memory.New()
	cpu := powerpc.New(mem)

	header, info, err := BootViaApploader(newTestImageWithApploader(), mem, cpu)
	if err != nil {
		t.Fatalf("BootViaApploader: %v", err)
	}
	if header.GameName != "Apploader Test" {
		t.Fatalf("GameName = %q, want %q", header.GameName, "Apploader Test")
	}
	if info.Entry != apploaderLoadAddr {
		t.Fatalf("info.Entry = %#08x, want %#08x", info.Entry, uint32(apploaderLoadAddr))
	}
	if info.Size != 8 {
		t.Fatalf("info.Size = %d, want 8", info.Size)
	}
	if cpu.PC() != apploaderLoadAddr {
		t.Fatalf("PC = %#08x, want %#08x", cpu.PC(), uint32(apploaderLoadAddr))
	}
	if cpu.GPR(1) != stackTop {
		t.Fatalf("r1 (SP) = %#08x, want %#08x", cpu.GPR(1), uint32(stackTop))
	}

	// The loaded "apploader code" should be real, executable
	// instructions - not just bytes sitting in memory.
	cpu.Step()
	if cpu.GPR(5) != 9 {
		t.Fatalf("r5 after executing the loaded apploader code = %d, want 9", cpu.GPR(5))
	}
}

func TestBootViaApploaderRejectsTooSmallImage(t *testing.T) {
	mem := memory.New()
	cpu := powerpc.New(mem)
	if _, _, err := BootViaApploader(make([]byte, 16), mem, cpu); err == nil {
		t.Fatal("expected an error for an image too small to hold an apploader header")
	}
}

func TestBootViaApploaderRejectsCodeRunningPastImage(t *testing.T) {
	img := newTestImageWithApploader()
	putU32(img[apploaderHeaderOffset:], 0x14, uint32(len(img))) // size far larger than what's available

	mem := memory.New()
	cpu := powerpc.New(mem)
	if _, _, err := BootViaApploader(img, mem, cpu); err == nil {
		t.Fatal("expected an error when the apploader's declared size runs past the image")
	}
}
