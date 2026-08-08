package ipl

import (
	"testing"

	"github.com/Duc-inc/espaze/internal/systems/gamecube/memory"
	"github.com/Duc-inc/espaze/internal/systems/powerpc"
)

// newTestImage builds a minimal disc image: a valid header pointing
// at a DOL (placed right after the header) with one text section
// containing a single instruction, loaded at a MEM1 address.
func newTestImage() []byte {
	const dolOffset = 0x1000
	const loadAddr = 0x00003000
	const dolFileSize = 0x200

	img := make([]byte, dolOffset+dolFileSize)
	copy(img[0:6], []byte("GTEST"))
	putU32(img, 0x1C, 0xC2339F3D) // magic
	copy(img[0x20:], []byte("Boot Test\x00"))
	putU32(img, 0x420, dolOffset)

	dol := img[dolOffset:]
	putU32(dol, 0x00, 0x100)    // text0 file offset (DOL-relative)
	putU32(dol, 0x48, loadAddr) // text0 load address
	putU32(dol, 0x90, 4)        // text0 size
	putU32(dol, 0xE0, loadAddr) // entry point

	// addi r3,r0,7 at file offset 0x100 within the DOL.
	instr := uint32(14)<<26 | 3<<21 | 0<<16 | 7
	putU32(dol, 0x100, instr)

	return img
}

func putU32(b []byte, offset int, v uint32) {
	b[offset] = byte(v >> 24)
	b[offset+1] = byte(v >> 16)
	b[offset+2] = byte(v >> 8)
	b[offset+3] = byte(v)
}

func TestBootLoadsDOLAndSetsEntryPoint(t *testing.T) {
	mem := memory.New()
	cpu := powerpc.New(mem)

	header, err := Boot(newTestImage(), mem, cpu)
	if err != nil {
		t.Fatalf("Boot: %v", err)
	}
	if header.GameName != "Boot Test" {
		t.Fatalf("GameName = %q, want %q", header.GameName, "Boot Test")
	}
	if cpu.PC() != 0x00003000 {
		t.Fatalf("PC = %#08x, want 0x00003000", cpu.PC())
	}
	if cpu.GPR(1) != stackTop {
		t.Fatalf("r1 (SP) = %#08x, want %#08x", cpu.GPR(1), uint32(stackTop))
	}

	cpu.Step()
	if cpu.GPR(3) != 7 {
		t.Fatalf("r3 after executing the loaded instruction = %d, want 7", cpu.GPR(3))
	}
}

func TestBootRejectsInvalidImage(t *testing.T) {
	mem := memory.New()
	cpu := powerpc.New(mem)
	if _, err := Boot(make([]byte, 4), mem, cpu); err == nil {
		t.Fatal("expected an error for an invalid disc image")
	}
}

func TestSyscallHandlerIsInvoked(t *testing.T) {
	mem := memory.New()
	cpu := powerpc.New(mem)

	called := false
	cpu.SyscallHandler = func(c *powerpc.CPU) { called = true }

	scInstr := uint32(17) << 26
	mem.Write32(0, scInstr)
	cpu.Step()

	if !called {
		t.Fatal("expected the syscall handler to be invoked by `sc`")
	}
}
