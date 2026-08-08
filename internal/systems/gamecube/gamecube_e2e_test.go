// gamecube_e2e_test.go holds this project's proof that the whole GX
// rendering chain actually works, not just each piece in isolation:
// real hand-assembled PowerPC programs, executed by the real CPU,
// streaming real GX command bytes through the real Write Gather Pipe
// (see wgp's doc comment) into the real CommandProcessor and
// Framebuffer. No test here pokes CP/XF/BP state directly.
package gamecube

import (
	"testing"

	"github.com/Duc-inc/espaze/internal/systems/gamecube/gpu"
	"github.com/Duc-inc/espaze/internal/systems/gamecube/wgp"
)

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

// texturedTriangleDrawStream builds a real BP-driven texture bind
// (address/format/width/height, YAGCD's TX_SETIMAGE-equivalent fields
// this project reserves its own BP addresses for - see bind.go's doc
// comment) followed by the same triangle as triangleDrawStream, but
// with non-zero UV so the bound texture actually shows, padded to the
// WGP's 32-byte burst size like every other real command list here.
func texturedTriangleDrawStream(texAddr uint32) []byte {
	const (
		cmdLoadBPReg     = 0x61
		cmdDrawTriangles = 0x90
		bpTexAddr        = 0x10
		bpTexFormat      = 0x11
		bpTexWidth       = 0x12
		bpTexHeight      = 0x13
		formatRGBA8      = 2
	)
	bpCmd := func(reg byte, data uint32) []byte {
		return []byte{cmdLoadBPReg, reg, byte(data >> 16), byte(data >> 8), byte(data)}
	}
	vertex := func(x, y, z, u, v int16) []byte {
		return []byte{
			byte(x >> 8), byte(x), byte(y >> 8), byte(y), byte(z >> 8), byte(z),
			0, 0, 0, 0, 0, 0,
			255, 255, 255, 255,
			byte(u >> 8), byte(u), byte(v >> 8), byte(v),
		}
	}

	var stream []byte
	stream = append(stream, bpCmd(bpTexAddr, texAddr)...)
	stream = append(stream, bpCmd(bpTexFormat, formatRGBA8)...)
	stream = append(stream, bpCmd(bpTexWidth, 2)...)
	stream = append(stream, bpCmd(bpTexHeight, 2)...) // last write completes the bind (rebindTexture)
	stream = append(stream, cmdDrawTriangles, 0x00, 0x03)
	stream = append(stream, vertex(50, 50, 0, 0, 0)...)
	stream = append(stream, vertex(590, 50, 0, 40, 0)...)
	stream = append(stream, vertex(320, 430, 0, 0, 40)...)
	for len(stream)%wgp.Size != 0 {
		stream = append(stream, 0x00)
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

// linesPerFrameForTest mirrors vi's own linesPerFrame (unexported,
// package-local) - the raster line count a Step-per-line VI wraps at.
const linesPerFrameForTest = 525

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

// TestEndToEndCPUDrivenTexturedDrawSamplesRealMemory proves the
// texture path works through the same real pipeline: a real BP-driven
// texture bind pointing at real texel bytes this test wrote into
// MEM1, sampled by the rasterizer while drawing a real CPU-driven
// triangle. Visually verified once during development (rendered to a
// PNG and inspected: a red/green/blue/white checkerboard tiled across
// the triangle) - this test checks the same fact programmatically: two
// well-separated sampled pixels must differ, and must not be flat
// white, which would only happen if texturing weren't actually
// reaching the rasterizer.
func TestEndToEndCPUDrivenTexturedDrawSamplesRealMemory(t *testing.T) {
	g := New(nil)

	const texAddr = 0x00003000
	// 2x2 RGBA8: red, green / blue, white (row-major).
	texData := []byte{
		255, 0, 0, 255, 0, 255, 0, 255,
		0, 0, 255, 255, 255, 255, 255, 255,
	}
	g.LoadAt(texAddr, texData)

	instrCount := loadWGPCopyProgram(g, texturedTriangleDrawStream(texAddr), 0x2000, 0x4000)
	for i := 0; i < instrCount; i++ {
		g.Step()
	}
	g.FlushGP()

	fb := g.FB.FrameBuffer()
	pixelAt := func(x, y int) (byte, byte, byte) {
		idx := (y*fb.Width + x) * 4
		return fb.Pixels[idx], fb.Pixels[idx+1], fb.Pixels[idx+2]
	}
	// The checkerboard texture tiles many times across the triangle
	// (UV spans 0-40 over a 2x2 texture), so two points this far apart
	// landing on the same color would be exceedingly unlikely if
	// texturing is actually happening - if it weren't wired up at all,
	// every pixel would instead be the flat lit vertex color (white).
	r1, g1, b1 := pixelAt(150, 150)
	r2, g2, b2 := pixelAt(400, 300)
	if r1 == r2 && g1 == g2 && b1 == b2 {
		t.Fatalf("pixel(150,150)=(%d,%d,%d) == pixel(400,300)=(%d,%d,%d): expected the tiled checkerboard texture to vary, not a flat color",
			r1, g1, b1, r2, g2, b2)
	}
	if r1 == 255 && g1 == 255 && b1 == 255 {
		t.Fatal("pixel(150,150) is pure white: expected a sampled checkerboard texel, not the flat unlit vertex color (texturing not applied)")
	}
}
