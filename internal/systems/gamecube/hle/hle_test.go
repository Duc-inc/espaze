package hle

import (
	"testing"

	"github.com/Duc-inc/espaze/internal/systems/gamecube/memory"
	"github.com/Duc-inc/espaze/internal/systems/powerpc"
)

func TestInstallTrapsAndReturnsToCaller(t *testing.T) {
	mem := memory.New()
	cpu := powerpc.New(mem)

	const funcAddr = 0x2000
	const callSite = 0x1000
	const returnAddr = 0x1004

	// bl funcAddr (from callSite) - real PowerPC would set LR to
	// callSite+4 and jump to funcAddr; this project's test builds
	// that state directly rather than encoding a real branch, since
	// what's under test is Install's trap/return mechanism, not
	// branch decoding (already covered elsewhere).
	cpu.SetPC(funcAddr)
	cpu.SetGPR(1, 0) // LR isn't GPR-addressable; set via the field below
	// There's no exported SetLR, so drive it through an actual mtlr:
	// addi r5,r0,returnAddr ; mtlr r5
	mem.Write32(0, uint32(14)<<26|5<<21|0<<16|returnAddr)
	mem.Write32(4, uint32(31)<<26|5<<21|8<<16|0<<11|467<<1)
	cpu.SetPC(0)
	cpu.Step()
	cpu.Step()
	cpu.SetPC(funcAddr)

	called := false
	table := New(mem)
	table.Install(cpu, funcAddr, func(cpu *powerpc.CPU, mem MemoryAccess) {
		called = true
		cpu.SetGPR(3, 42)
	})

	cpu.Step() // executes the trap Install planted at funcAddr

	if !called {
		t.Fatal("expected the installed function to run")
	}
	if cpu.GPR(3) != 42 {
		t.Fatalf("GPR3 = %d, want 42", cpu.GPR(3))
	}
	if cpu.PC() != returnAddr {
		t.Fatalf("PC = %#08x, want %#08x (returned via LR)", cpu.PC(), uint32(returnAddr))
	}
}

func TestOSReportCollectsReportedString(t *testing.T) {
	mem := memory.New()
	cpu := powerpc.New(mem)
	log := &ReportLog{}

	const strAddr = 0x3000
	writeCString(mem, strAddr, "hello from the game")
	cpu.SetGPR(3, strAddr)

	OSReport(log)(cpu, mem)

	if len(log.Lines) != 1 || log.Lines[0] != "hello from the game" {
		t.Fatalf("log.Lines = %v, want [\"hello from the game\"]", log.Lines)
	}
}

func TestDVDReadCopiesBytesFromImage(t *testing.T) {
	mem := memory.New()
	cpu := powerpc.New(mem)

	image := make([]byte, 0x2000)
	copy(image[0x1000:], []byte{0xDE, 0xAD, 0xBE, 0xEF})

	const dest = 0x4000
	cpu.SetGPR(3, dest)
	cpu.SetGPR(4, 4)      // length
	cpu.SetGPR(5, 0x1000) // disc offset

	DVDRead(image)(cpu, mem)

	if cpu.GPR(3) != 4 {
		t.Fatalf("GPR3 (bytes read) = %d, want 4", cpu.GPR(3))
	}
	if mem.Read8(dest) != 0xDE || mem.Read8(dest+3) != 0xEF {
		t.Fatalf("copied bytes = %#x %#x, want 0xde ... 0xef", mem.Read8(dest), mem.Read8(dest+3))
	}
}

func TestDVDReadStopsAtImageBoundary(t *testing.T) {
	mem := memory.New()
	cpu := powerpc.New(mem)
	image := make([]byte, 4)

	cpu.SetGPR(3, 0x4000)
	cpu.SetGPR(4, 100) // asks for far more than the image has
	cpu.SetGPR(5, 0)

	DVDRead(image)(cpu, mem)

	if cpu.GPR(3) != 4 {
		t.Fatalf("GPR3 (bytes read) = %d, want 4 (clamped to image size)", cpu.GPR(3))
	}
}

func writeCString(mem MemoryAccess, addr uint32, s string) {
	for i := 0; i < len(s); i++ {
		mem.Write8(addr+uint32(i), s[i])
	}
	mem.Write8(addr+uint32(len(s)), 0)
}
