// Package gpu implements the Flipper GPU's Command Processor (CP): the
// decoder for the stream of GX commands real GameCube games write to
// drive rendering, plus (in render.go) a basic software rasterizer
// this project uses in place of Flipper's actual hardware pipeline.
// The command opcode values themselves are well-documented in public
// community references (e.g. YAGCD - Yet Another GameCube
// Documentation - independent of any specific emulator's source).
// Vertex attribute format decoding is simplified to one fixed layout
// (position, normal, UV texture coordinate, and color) rather than
// the real chip's fully configurable per-attribute format table.
//
// Every decoded vertex position is run through the xf package's
// transform pipeline (position matrix -> projection -> perspective
// divide -> viewport) before it reaches a Triangle, so LOAD_XF_REG
// commands that upload real matrix data now actually affect where a
// triangle ends up, not just get parsed and discarded. The vertex
// normal is transformed and lit (xf.Illuminate) using the ambient/
// light state, replacing the vertex's own color with the lit result
// before it reaches a Triangle - by default (white ambient, no
// lights) this leaves colors unchanged, matching this project's
// behavior before lighting existed. A texture bound (see bind.go) is
// sampled per pixel and combined with that lit color (render.go).
//
// bind.go decodes LOAD_BP_REG/LOAD_XF_REG writes at a small set of
// addresses this project reserves for binding state (texture setup,
// ambient, light objects, active matrix index) - a real command
// stream now drives what used to require calling SetTexture/
// SetAmbient/SetLight/etc. directly. Real hardware's exact BP/XF
// register numbers for this state aren't independently verified here
// (unlike the well-documented CP/XF/BP command opcodes themselves);
// this project reserves its own addresses for it rather than guessing
// real ones, the same way the vertex format is this project's own
// simplified layout rather than real hardware's configurable one.
// SetTexture/SetAmbient/SetLight remain available too, for callers
// that don't go through a command stream.
//
// indexed.go adds a second vertex path alongside the direct layout
// above: a reserved CP register (cpVertexMode) switches a draw call to
// reading 16-bit position/normal/UV array indices instead of full
// vertex data, resolved against array base/stride CP registers and
// fetched through the same MemoryReader texture binding uses - closer
// to how real games actually supply geometry (shared vertex arrays,
// referenced by index) than always re-sending full vertex data.
package gpu

import (
	"github.com/Duc-inc/espaze/internal/systems/gamecube/texture"
	"github.com/Duc-inc/espaze/internal/systems/gamecube/xf"
)

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
// position, a normal, a texture coordinate, and a flat RGBA color,
// read directly from the command stream (20 bytes: 3x int16 position,
// 3x int16 normal, 2x int16 UV, 4x byte color) - real hardware's
// vertex format is fully configurable (position can be 2D/3D, indexed
// or direct, with independent normal and up to 8 texture-coordinate
// attributes); this project always expects the one layout. The
// position starts out in model space as decoded, then gets replaced
// with its screen-space result once Execute runs it through the xf
// transform pipeline; the color similarly starts as the vertex's own
// flat color and gets replaced by the lit result of that color once
// Execute runs the normal through xf.Illuminate - by the time a
// Vertex reaches a Triangle, X/Y/Z are screen coordinates and R/G/B
// already include lighting, both ready for the rasterizer
// (render.go). U/V are left as raw decoded pixel-space coordinates
// into whatever texture is bound (texture.Texture.Sample), not real
// hardware's normalized/fixed-point texture coordinates.
type Vertex struct {
	X, Y, Z    int16
	NX, NY, NZ int16
	U, V       int16
	R, G, B, A byte
}

const vertexBytes = 3*2 + 3*2 + 2*2 + 4

// MaxTexStages is how many simultaneous texture stages this project
// supports - matching real hardware's own texture map count (GX_MAX_
// TEXMAP-family limit); each is independently bindable and combines
// into the output color in sequence (bind.go, render.go), the same
// chaining model real TEV stages use.
const MaxTexStages = 8

// MemoryReader lets the Command Processor fetch texture data from
// main GameCube memory once BP registers point a texture at it - the
// same role real hardware's TMEM-loading hardware plays. Without one
// set (SetMemoryReader), BP-driven texture binding has no bytes to
// read and is a no-op; SetTexture still works either way.
type MemoryReader interface {
	ReadBytes(addr uint32, length int) []byte
}

// CommandProcessor decodes a GX command stream into draw calls plus
// CP/XF/BP register writes.
type CommandProcessor struct {
	cpRegs [256]uint32
	bpRegs [256]uint32

	xfMemory    *xf.Memory
	xfRegisters xf.Registers

	ambient xf.Ambient
	lights  [xf.MaxLights]xf.Light

	boundTextures [MaxTexStages]*texture.Texture
	memReader     MemoryReader
	pendingTex    [MaxTexStages]pendingTexture
	tevOps        [MaxTexStages]TEVOp
	activeTexSlot int

	pendingTriangles []Triangle
}

// SetMemoryReader wires up how bind.go's BP-driven texture binding
// fetches texel bytes from main memory.
func (cp *CommandProcessor) SetMemoryReader(m MemoryReader) {
	cp.memReader = m
}

// SetAmbient sets the constant ambient light term applied to every
// vertex transformed after this call - a stand-in for real hardware's
// XF-register-driven ambient color, which isn't decoded from the
// command stream yet.
func (cp *CommandProcessor) SetAmbient(a xf.Ambient) {
	cp.ambient = a
}

// SetLight sets one of up to xf.MaxLights point lights - a stand-in
// for real hardware's XF-register-driven light object uploads, which
// aren't decoded from the command stream yet. An out-of-range index
// is ignored.
func (cp *CommandProcessor) SetLight(index int, l xf.Light) {
	if index < 0 || index >= len(cp.lights) {
		return
	}
	cp.lights[index] = l
}

// Triangle is one shaded triangle ready for rasterization. Textures
// holds up to MaxTexStages bound textures (nil slots are unbound); the
// rasterizer (render.go) samples each bound slot in order and
// combines it into the running color via the matching TEVOps entry,
// starting from the lit vertex color - falling back to pure Gouraud
// shading when every slot is nil.
type Triangle struct {
	V0, V1, V2 Vertex
	Textures   [MaxTexStages]*texture.Texture
	TEVOps     [MaxTexStages]TEVOp
}

// SetTexture sets the texture bound to the given stage - a stand-in
// for real hardware's BP-register-driven TMEM binding, which isn't
// decoded from the command stream yet. An out-of-range slot is
// ignored.
func (cp *CommandProcessor) SetTexture(slot int, t *texture.Texture) {
	if slot < 0 || slot >= MaxTexStages {
		return
	}
	cp.boundTextures[slot] = t
}

// New returns a CommandProcessor with every register zeroed and the
// XF pipeline defaulted to an identity transform (identity position
// matrix at address 0, an orthographic no-op projection, an unscaled
// viewport, and a white ambient term with no lights enabled) - so a
// command stream that never uploads any matrices or lights still
// renders vertices at their raw decoded position and color, exactly
// as this project did before the xf package existed. A real game
// overwrites these defaults via LOAD_XF_REG/SetAmbient/SetLight once
// it starts drawing.
func New() *CommandProcessor {
	mem := xf.NewMemory()
	mem.WritePosMatrix(0, xf.IdentityPos())

	regs := xf.NewRegisters()
	regs.Projection = xf.Projection{
		Coeffs: [6]float32{1, 0, 1, 0, 1, 0},
		Type:   xf.ProjectionOrthographic,
	}

	cp := &CommandProcessor{
		xfMemory:    mem,
		xfRegisters: regs,
		ambient:     xf.Ambient{Color: xf.LightColor{R: 1, G: 1, B: 1}},
	}
	cp.tevOps[0] = TEVModulate // matches this project's pre-multi-texture default
	return cp
}

// DrainTriangles returns and clears every triangle decoded since the
// last call.
func (cp *CommandProcessor) DrainTriangles() []Triangle {
	out := cp.pendingTriangles
	cp.pendingTriangles = nil
	return out
}
