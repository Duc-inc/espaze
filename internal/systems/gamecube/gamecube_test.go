package gamecube

import (
	"testing"

	"github.com/Duc-inc/espaze/internal/systems/gamecube/ai"
	"github.com/Duc-inc/espaze/internal/systems/gamecube/gpu"
	"github.com/Duc-inc/espaze/internal/systems/gamecube/pi"
	"github.com/Duc-inc/espaze/internal/systems/gamecube/wgp"
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

// triangleDrawStream builds a real GX DRAW_TRIANGLES command list for
// one white triangle in the default vertex format, padded to the
// WGP's 32-byte burst size the way real GX code (GX_Flush) always
// does before relying on the WGP having forwarded everything.
func triangleDrawStream() []byte {
	vertex := func(x, y, z int16) []byte {
		return []byte{
			byte(x >> 8), byte(x), byte(y >> 8), byte(y), byte(z >> 8), byte(z),
			0, 0, 0, 0, 0, 0, // normal (unused: default white ambient, no lights)
			255, 255, 255, 255, // color: opaque white
			0, 0, 0, 0, // uv
		}
	}
	const cmdDrawTriangles = 0x90
	stream := []byte{cmdDrawTriangles, 0x00, 0x03}
	stream = append(stream, vertex(50, 50, 0)...)
	stream = append(stream, vertex(590, 50, 0)...)
	stream = append(stream, vertex(320, 430, 0)...)
	for len(stream)%wgp.Size != 0 {
		stream = append(stream, 0x00) // cmdNop
	}
	return stream
}

// loadWGPCopyProgram places stream in MEM1 at srcAddr and a real
// hand-assembled PowerPC program at progAddr that copies it byte by
// byte to wgp.Base (lbz r6,i(r3) ; stb r6,0(r4) per byte - real load/
// store opcodes, no loop needed since the length is known at build
// time), then seeds r3/r4/PC to run it. Returns the instruction count
// so the caller knows how many Steps the copy itself takes.
func loadWGPCopyProgram(g *GameCube, stream []byte, srcAddr, progAddr uint32) int {
	g.LoadAt(srcAddr, stream)

	var progBytes []byte
	emit := func(instr uint32) {
		progBytes = append(progBytes, byte(instr>>24), byte(instr>>16), byte(instr>>8), byte(instr))
	}
	for i := range stream {
		emit(uint32(34)<<26 | 6<<21 | 3<<16 | uint32(uint16(i))) // lbz r6,i(r3)
		emit(uint32(38)<<26 | 6<<21 | 4<<16 | 0)                 // stb r6,0(r4)
	}
	g.LoadAt(progAddr, progBytes)

	g.proc.SetGPR(3, srcAddr)
	g.proc.SetGPR(4, wgp.Base)
	g.proc.SetPC(progAddr)
	return len(progBytes) / 4
}

func assertWhiteTriangleRendered(t *testing.T, g *GameCube) {
	t.Helper()
	fb := g.FB.FrameBuffer()
	// (320,200) sits inside the triangle (50,50)-(590,50)-(320,430).
	idx := (200*fb.Width + 320) * 4
	r, gg, b := fb.Pixels[idx], fb.Pixels[idx+1], fb.Pixels[idx+2]
	if r != 255 || gg != 255 || b != 255 {
		t.Fatalf("pixel at (320,200) = (%d,%d,%d), want (255,255,255): a real CPU-driven GX draw should have rasterized a white triangle there", r, gg, b)
	}
	// A point clearly outside the triangle should remain the cleared
	// background, confirming this isn't just a full-white framebuffer.
	outIdx := (10*fb.Width + 10) * 4
	if fb.Pixels[outIdx] != 0 {
		t.Fatalf("pixel at (10,10) = %d, want 0 (background, outside the triangle)", fb.Pixels[outIdx])
	}
}

// TestEndToEndCPUDrivenDrawRendersATriangle is this project's proof
// that the whole chain actually works, not just each piece in
// isolation: a real hand-assembled PowerPC program copies a real GX
// command stream byte by byte to wgp.Base - the real Write Gather Pipe
// address - exactly the way real game code would. No test helper
// pokes CP/XF state directly; everything happens through g.Step()
// executing real instructions and an explicit g.FlushGP() draining the
// real WGP -> CP -> Framebuffer pipeline.
func TestEndToEndCPUDrivenDrawRendersATriangle(t *testing.T) {
	g := New(nil)
	instrCount := loadWGPCopyProgram(g, triangleDrawStream(), 0x2000, 0x4000)

	for i := 0; i < instrCount; i++ {
		g.Step()
	}
	g.FlushGP()

	assertWhiteTriangleRendered(t, g)
}

// TestVBlankAutoFlushesWithoutAnExplicitCall proves Step's automatic
// FlushGP-at-vblank actually fires on its own: same real CPU-driven
// draw as above, but this test never calls FlushGP - it just keeps
// stepping (past the end of the loaded program, into zeroed/undefined-
// instruction memory that's a harmless 2-cycle no-op) until VI's
// raster position wraps into a new frame on its own.
func TestVBlankAutoFlushesWithoutAnExplicitCall(t *testing.T) {
	g := New(nil)
	instrCount := loadWGPCopyProgram(g, triangleDrawStream(), 0x2000, 0x4000)

	for i := 0; i < instrCount; i++ {
		g.Step()
	}
	for i := instrCount; i < linesPerFrameForTest; i++ {
		g.Step()
	}

	assertWhiteTriangleRendered(t, g)
}

// TestFlushGPWritesRealXFBBytesToVIsFramebufferAddress closes the last
// loop: a real CPU-driven draw, flushed, should leave the exact bytes
// a real VI video encoder would read at the exact address a real game
// configured via TFBL - not just an internal Framebuffer object no
// memory-mapped consumer could ever see.
func TestFlushGPWritesRealXFBBytesToVIsFramebufferAddress(t *testing.T) {
	g := New(nil)
	g.VI.Write32(0x02, 1)       // DCR: ENB=1
	const xfbAddr = 0x00180000  // arbitrary valid MEM1 address
	g.VI.Write32(0x1C, xfbAddr) // TFBL

	instrCount := loadWGPCopyProgram(g, triangleDrawStream(), 0x2000, 0x4000)
	for i := 0; i < instrCount; i++ {
		g.Step()
	}
	g.FlushGP()

	want := gpu.EncodeXFB(g.FB.FrameBuffer())
	for i := 0; i < 64; i++ { // spot-check the first 64 bytes, not the whole 614400-byte frame
		got := g.bus.Read8(xfbAddr + uint32(i))
		if got != want[i] {
			t.Fatalf("MEM1[xfbAddr+%d] = %#02x, want %#02x (EncodeXFB's own output)", i, got, want[i])
		}
	}
}

// linesPerFrameForTest mirrors vi's own linesPerFrame (unexported,
// package-local) - the raster line count a Step-per-line VI wraps at.
const linesPerFrameForTest = 525

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
