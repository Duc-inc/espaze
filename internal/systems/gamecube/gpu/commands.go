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
package gpu

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

// Vertex is this project's own simplified fixed vertex layout:
// screen-space position plus a flat RGBA color, read directly from
// the command stream (12 bytes: 3x int16 position, 4x byte color) -
// real hardware's vertex format is fully configurable (position can
// be 2D/3D, indexed or direct, with independent texture-coordinate
// and normal attributes); this project always expects the one layout.
type Vertex struct {
	X, Y, Z    int16
	R, G, B, A byte
}

const vertexBytes = 3*2 + 4

// CommandProcessor decodes a GX command stream into draw calls plus
// CP/XF/BP register writes.
type CommandProcessor struct {
	cpRegs [256]uint32
	xfRegs [256]uint32
	bpRegs [256]uint32

	pendingTriangles []Triangle
}

// Triangle is one flat-shaded triangle ready for rasterization.
type Triangle struct {
	V0, V1, V2 Vertex
}

// New returns a CommandProcessor with every register zeroed.
func New() *CommandProcessor { return &CommandProcessor{} }

// DrainTriangles returns and clears every triangle decoded since the
// last call.
func (cp *CommandProcessor) DrainTriangles() []Triangle {
	out := cp.pendingTriangles
	cp.pendingTriangles = nil
	return out
}
