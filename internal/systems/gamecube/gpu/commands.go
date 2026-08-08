// Package gpu implements the Flipper GPU's Command Processor (CP): the
// decoder for the stream of GX commands real GameCube games write to
// drive rendering, plus (in render.go) a basic software rasterizer
// this project uses in place of Flipper's actual hardware pipeline.
// The command opcode values themselves are well-documented in public
// community references (e.g. YAGCD - Yet Another GameCube
// Documentation - independent of any specific emulator's source).
// Vertex attribute format decoding is simplified to one fixed layout
// (position + color, no texture coordinates or normals) rather than
// the real chip's fully configurable per-attribute format table.
//
// Every decoded vertex position is run through the xf package's
// transform pipeline (position matrix -> projection -> perspective
// divide -> viewport) before it reaches a Triangle, so LOAD_XF_REG
// commands that upload real matrix data now actually affect where a
// triangle ends up, not just get parsed and discarded.
package gpu

import "github.com/Duc-inc/espaze/internal/systems/gamecube/xf"

// GX command opcodes (the primary byte of each command in the stream).
const (
	cmdNop             = 0x00
	cmdLoadCPReg       = 0x08
	cmdLoadXFReg       = 0x10
	cmdCallDisplayList = 0x40
	cmdLoadBPReg       = 0x61
	cmdDrawQuads       = 0x80
	cmdDrawTriangles   = 0x90
	cmdDrawTriStrip    = 0x98
	cmdDrawTriFan      = 0xA0
	cmdDrawLines       = 0xA8
	cmdDrawLineStrip   = 0xB0
	cmdDrawPoints      = 0xB8
)

// Vertex is this project's own simplified fixed vertex layout: a
// position plus a flat RGBA color, read directly from the command
// stream (12 bytes: 3x int16 position, 4x byte color) - real
// hardware's vertex format is fully configurable (position can be 2D/
// 3D, indexed or direct, with independent texture-coordinate and
// normal attributes); this project always expects the one layout. The
// position starts out in model space as decoded, then gets replaced
// with its screen-space result once Execute runs it through the xf
// transform pipeline - by the time a Vertex reaches a Triangle, X/Y/Z
// are screen coordinates ready for the rasterizer (render.go).
type Vertex struct {
	X, Y, Z    int16
	R, G, B, A byte
}

const vertexBytes = 3*2 + 4

// CommandProcessor decodes a GX command stream into draw calls plus
// CP/XF/BP register writes.
type CommandProcessor struct {
	cpRegs [256]uint32
	bpRegs [256]uint32

	xfMemory    *xf.Memory
	xfRegisters xf.Registers

	pendingTriangles []Triangle
}

// Triangle is one flat-shaded triangle ready for rasterization.
type Triangle struct {
	V0, V1, V2 Vertex
}

// New returns a CommandProcessor with every register zeroed and the
// XF pipeline defaulted to an identity transform (identity position
// matrix at address 0, an orthographic no-op projection, and an
// unscaled viewport) - so a command stream that never uploads any
// matrices still renders vertices at their raw decoded position,
// exactly as this project did before the xf package existed. A real
// game overwrites this default via LOAD_XF_REG once it starts
// drawing.
func New() *CommandProcessor {
	mem := xf.NewMemory()
	mem.WritePosMatrix(0, xf.IdentityPos())

	regs := xf.NewRegisters()
	regs.Projection = xf.Projection{
		Coeffs: [6]float32{1, 0, 1, 0, 1, 0},
		Type:   xf.ProjectionOrthographic,
	}

	return &CommandProcessor{xfMemory: mem, xfRegisters: regs}
}

// DrainTriangles returns and clears every triangle decoded since the
// last call.
func (cp *CommandProcessor) DrainTriangles() []Triangle {
	out := cp.pendingTriangles
	cp.pendingTriangles = nil
	return out
}
