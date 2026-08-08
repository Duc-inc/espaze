package gamecube

import (
	"testing"

	"github.com/Duc-inc/espaze/internal/systems/gamecube/ai"
	"github.com/Duc-inc/espaze/internal/systems/gamecube/hle"
	"github.com/Duc-inc/espaze/internal/systems/gamecube/pi"
	"github.com/Duc-inc/espaze/internal/systems/powerpc"
)

func TestLoadAtAndStepExecutesInstruction(t *testing.T) {
	g := New(nil)
	// addi r3, r0, 42 at address 0
	instr := uint32(14)<<26 | 3<<21 | 0<<16 | 42
	g.LoadAt(0, []byte{byte(instr >> 24), byte(instr >> 16), byte(instr >> 8), byte(instr)})

	g.Step()
	if g.proc.PC() != 4 {
		t.Fatalf("PC after one Step = %d, want 4", g.proc.PC())
	}
}

func TestPeripheralRegistersAreReachableFromRealPowerPCCode(t *testing.T) {
	g := New(nil)
	// addi r3,r0,0x0001 ; addi r4,r0,(AI.Base low bits offset 0) via lis/ori
	// simplest: use SetGPR to seed the address, then stw through it.
	g.proc.SetGPR(4, ai.Base)               // r4 = AI register block base (AICR)
	g.proc.SetGPR(3, 1)                     // r3 = PSTAT bit
	instr := uint32(36)<<26 | 3<<21 | 4<<16 // stw r3,0(r4)
	g.LoadAt(0, []byte{byte(instr >> 24), byte(instr >> 16), byte(instr >> 8), byte(instr)})

	g.Step()

	if !g.AI.Playing() {
		t.Fatal("expected a real stw through the CPU to reach AI's AICR register")
	}
}

func TestVIVBlankRaisesRealExternalInterrupt(t *testing.T) {
	g := New(nil)
	g.proc.SetMSR(powerpc.MSREE)
	g.PI.Write32(0x04, 1<<pi.BitVI) // INTMR: unmask VI
	g.VI.Write32(0x30, 1<<28|1)     // DI0: enabled, target line 1 (vpos after one Step)

	g.Step()

	if g.proc.PC() != powerpc.ExternalInterruptVector {
		t.Fatalf("PC after VBlank = %#x, want external interrupt vector %#x", g.proc.PC(), powerpc.ExternalInterruptVector)
	}
}

func TestMaskedInterruptDoesNotReachTheCPU(t *testing.T) {
	g := New(nil)
	g.proc.SetMSR(powerpc.MSREE)
	// INTMR left at 0 (everything masked).
	g.VI.Write32(0x30, 1<<28|1)

	g.Step()

	if g.proc.PC() == powerpc.ExternalInterruptVector {
		t.Fatal("expected a masked VI interrupt not to reach the CPU")
	}
}

func TestAIPlayingGatesAudioMixerStepping(t *testing.T) {
	g := New(nil)
	g.Audio.SetChannel(0, 1000, 255, true)

	g.Step() // AI not playing yet: Audio shouldn't advance
	if len(g.Audio.DrainSamples()) != 0 {
		t.Fatal("expected no samples while AI isn't playing")
	}

	g.proc.SetGPR(4, ai.Base)
	g.proc.SetGPR(3, 1) // PSTAT=1
	instr := uint32(36)<<26 | 3<<21 | 4<<16
	g.LoadAt(4, []byte{byte(instr >> 24), byte(instr >> 16), byte(instr >> 8), byte(instr)})
	g.proc.SetPC(4)
	g.Step() // stw sets PSTAT; AI is playing by the time this same Step checks it

	if len(g.Audio.DrainSamples()) != 1 {
		t.Fatal("expected exactly one sample once AI starts playing")
	}
}

func TestDITransferCompleteRaisesRealExternalInterrupt(t *testing.T) {
	g := New(make([]byte, 0x100))
	g.proc.SetMSR(powerpc.MSREE)
	g.PI.Write32(0x04, 1<<pi.BitDI) // INTMR: unmask DI
	g.DI.Write32(0x00, 1<<3)        // DISR: TCINTMASK=1 (DI's own local mask)

	g.DI.Write32(0x08, 0) // CMDBUF0: read-sector opcode 0
	g.DI.Write32(0x18, 4) // DILENGTH
	g.DI.Write32(0x1C, 1) // DICR: TSTART -> runs the command, sets TCINT

	g.Step()

	if g.proc.PC() != powerpc.ExternalInterruptVector {
		t.Fatalf("PC after DI transfer complete = %#x, want external interrupt vector %#x", g.proc.PC(), powerpc.ExternalInterruptVector)
	}
}

func TestResetClearsState(t *testing.T) {
	g := New(nil)
	g.LoadAt(0, []byte{0, 0, 0, 1})
	g.Step()
	g.Reset()
	if g.proc.PC() != 0 {
		t.Fatalf("PC after Reset = %d, want 0", g.proc.PC())
	}
	if v := g.bus.Read8(0); v != 0 {
		t.Fatalf("MEM1[0] after Reset = %#02x, want 0", v)
	}
}

// TestInstallHLERunsThroughARealSyscallTrap proves InstallHLE's whole
// mechanism works via this GameCube's own bus/CPU: OSReport reads a
// string this test wrote into real MEM1, and the CPU actually returns
// to the real caller address afterward (via LR) instead of falling
// through the planted trap instruction.
func TestInstallHLERunsThroughARealSyscallTrap(t *testing.T) {
	g := New(nil)
	log := &hle.ReportLog{}
	const funcAddr = 0x3000
	const returnAddr = 0x1000
	g.InstallHLE(funcAddr, hle.OSReport(log))

	const strAddr = 0x4000
	msg := "hello from a real GameCube instance"
	g.LoadAt(strAddr, append([]byte(msg), 0))
	g.SetGPR(3, strAddr)

	// mtlr r5 needs r5 preloaded with returnAddr first.
	addi := uint32(14)<<26 | 5<<21 | 0<<16 | returnAddr // addi r5,r0,returnAddr
	mtlr := uint32(31)<<26 | 5<<21 | 8<<16 | 0<<11 | 467<<1
	g.LoadAt(0, []byte{
		byte(addi >> 24), byte(addi >> 16), byte(addi >> 8), byte(addi),
		byte(mtlr >> 24), byte(mtlr >> 16), byte(mtlr >> 8), byte(mtlr),
	})
	g.SetPC(0)
	g.Step() // addi
	g.Step() // mtlr

	g.SetPC(funcAddr)
	g.Step() // executes the trap InstallHLE planted at funcAddr

	if len(log.Lines) != 1 || log.Lines[0] != msg {
		t.Fatalf("log.Lines = %v, want [%q]", log.Lines, msg)
	}
	if g.PC() != returnAddr {
		t.Fatalf("PC = %#x, want %#x (returned via LR)", g.PC(), uint32(returnAddr))
	}
}

// TestBootRunsRealApploaderCodeThroughGameCubesOwnCPU proves Boot
// wires ipl.BootViaApploader against this GameCube's actual bus/CPU,
// not just the ipl package's own isolated test setup: after Boot, PC
// sits at the apploader's real entry point and Step actually executes
// the loaded code.
func TestBootRunsRealApploaderCodeThroughGameCubesOwnCPU(t *testing.T) {
	// Mirrors ipl's own private layout constants (apploaderHeaderOffset,
	// apploaderCodeOffset, apploaderLoadAddr) - see ipl/apploader.go.
	const (
		apploaderHeaderOffset = 0x2440
		apploaderCodeOffset   = apploaderHeaderOffset + 0x20
		apploaderLoadAddr     = 0x81200000
		codeSize              = 4
	)
	img := make([]byte, apploaderCodeOffset+codeSize)
	copy(img[0:6], []byte("GTEST"))
	putU32 := func(b []byte, off int, v uint32) {
		b[off] = byte(v >> 24)
		b[off+1] = byte(v >> 16)
		b[off+2] = byte(v >> 8)
		b[off+3] = byte(v)
	}
	putU32(img, 0x1C, 0xC2339F3D) // magic
	copy(img[0x20:], []byte("Boot Test\x00"))
	h := img[apploaderHeaderOffset:]
	putU32(h, 0x10, apploaderLoadAddr)
	putU32(h, 0x14, codeSize)
	putU32(h, 0x18, 0)
	// addi r7,r0,77 - a recognizable instruction as the apploader's "code".
	putU32(img, apploaderCodeOffset, uint32(14)<<26|7<<21|0<<16|77)

	g := New(nil)
	header, info, err := g.Boot(img)
	if err != nil {
		t.Fatalf("Boot: %v", err)
	}
	if header.GameName != "Boot Test" {
		t.Fatalf("GameName = %q, want %q", header.GameName, "Boot Test")
	}
	if g.PC() != info.Entry {
		t.Fatalf("PC = %#x, want apploader entry %#x", g.PC(), info.Entry)
	}

	g.Step()
	if got := g.proc.GPR(7); got != 77 {
		t.Fatalf("r7 after executing the booted apploader code = %d, want 77", got)
	}
}

// TestLoadDOLExecutesRealCodeAtItsEntryPoint proves LoadDOL's BAT
// translation is actually consistent between where it writes and
// where the CPU fetches from: a real DOL section loaded at an
// effective (cached-view) address runs correctly once PC (also an
// effective address) starts fetching there.
func TestLoadDOLExecutesRealCodeAtItsEntryPoint(t *testing.T) {
	const loadAddr = 0x80003100 // an effective, cached-view MEM1 address
	img := make([]byte, 0x200)
	putU32 := func(b []byte, off int, v uint32) {
		b[off] = byte(v >> 24)
		b[off+1] = byte(v >> 16)
		b[off+2] = byte(v >> 8)
		b[off+3] = byte(v)
	}
	// One text section: file offset 0x100, load address loadAddr, size 4.
	putU32(img, 0x00, 0x100)    // dolOffsetsBase
	putU32(img, 0x48, loadAddr) // dolAddrsBase
	putU32(img, 0x90, 4)        // dolSizesBase
	putU32(img, 0xE0, loadAddr) // dolEntryOffset
	// addi r8,r0,55 - a recognizable instruction as the DOL's "code".
	putU32(img, 0x100, uint32(14)<<26|8<<21|0<<16|55)

	g := New(nil)
	g.LoadDOL(img)

	if g.PC() != loadAddr {
		t.Fatalf("PC = %#x, want DOL entry %#x", g.PC(), uint32(loadAddr))
	}
	g.Step()
	if got := g.proc.GPR(8); got != 55 {
		t.Fatalf("r8 after executing the loaded DOL code = %d, want 55", got)
	}
}
