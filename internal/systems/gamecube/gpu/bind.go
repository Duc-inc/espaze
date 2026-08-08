// bind.go decodes LOAD_BP_REG writes at a small set of BP addresses
// this project reserves for texture binding and TEV configuration -
// see commands.go's package doc for why XF-side binding (matrix
// selection) goes through real registers instead, in xf/registers.go.
package gpu

import (
	"github.com/Duc-inc/espaze/internal/systems/gamecube/texture"
	"github.com/Duc-inc/espaze/internal/systems/gamecube/xf"
)

// BP register addresses this project reserves for texture binding.
// Real hardware spreads this across several TX_SETIMAGE-family
// registers per texture stage; this project uses one address per
// field instead, applied to whichever stage bpTexSlot last selected
// (defaulting to stage 0, so a command stream that never touches
// bpTexSlot behaves exactly as this project's original single-texture
// binding did), and rebuilds that stage's bound texture after every
// one of the four fields is set.
const (
	bpTexSlot   = 0x0F // selects which of MaxTexStages the fields below target
	bpTexAddr   = 0x10 // base byte address in main memory
	bpTexFormat = 0x11 // low byte: texture.Format value
	bpTexWidth  = 0x12
	bpTexHeight = 0x13

	// bpTevOp sets the TEV operation for the currently selected stage
	// (bpTexSlot) - each bound texture stage combines into the running
	// output color via its own op, chained in stage order (render.go),
	// the same chaining model real hardware's TEV stages use. Low
	// byte: a TEVOp value.
	bpTevOp = 0x14
)

// pendingTexture accumulates the four BP writes a texture bind needs
// before there's enough information to actually decode one.
type pendingTexture struct {
	addr, width, height uint32
	format              texture.Format
}

// applyBPRegisterWrite reacts to a BP register write that just landed
// in cp.bpRegs, updating any binding state that register address
// affects.
func (cp *CommandProcessor) applyBPRegisterWrite(reg byte) {
	slot := cp.activeTexSlot
	switch reg {
	case bpTexSlot:
		cp.activeTexSlot = int(cp.bpRegs[reg]) % MaxTexStages
		return
	case bpTexAddr:
		cp.pendingTex[slot].addr = cp.bpRegs[reg]
	case bpTexFormat:
		cp.pendingTex[slot].format = texture.Format(byte(cp.bpRegs[reg]))
	case bpTexWidth:
		cp.pendingTex[slot].width = cp.bpRegs[reg]
	case bpTexHeight:
		cp.pendingTex[slot].height = cp.bpRegs[reg]
	case bpTevOp:
		cp.tevOps[slot] = TEVOp(byte(cp.bpRegs[reg]))
		return
	default:
		return
	}
	cp.rebindTexture(slot)
}

// bridgeMatrixSelectionToXF forwards CP's MATIDX_REG_A straight into
// the real XF matrix-selection register. Both registers share the
// exact same 6-bit-field layout (vcd.go's MatIdxRegA doc comment), so
// a direct value copy decodes correctly on the XF side without
// reinterpreting the bits - but the copy itself is Espaze's own
// bridge, not a claim about real hardware's internal wiring between
// CP and XF, which this project hasn't verified against a public
// source or homebrew test. Espaze mirrors this value into XF state
// for now so vertices without a per-vertex matrix-index override
// (vertexformat.go) pick up whatever CP last set, matching the
// externally observable effect real games rely on.
func (cp *CommandProcessor) bridgeMatrixSelectionToXF(val uint32) {
	cp.xfState.Registers.Write(xf.RegMatrixSelection0, val)
}

// bridgeMatrixSelection1ToXF mirrors bridgeMatrixSelectionToXF for CP's
// MATIDX_REG_B / XF's RegMatrixSelection1 pair (texture matrix indices
// 4-7) - same reasoning, same caveat about this being Espaze's own
// bridge rather than a verified real wiring fact.
func (cp *CommandProcessor) bridgeMatrixSelection1ToXF(val uint32) {
	cp.xfState.Registers.Write(xf.RegMatrixSelection1, val)
}

// rebindTexture decodes and binds a texture for the given stage from
// its current pending state, if there's a memory reader to fetch
// bytes from and the texture has a real size.
func (cp *CommandProcessor) rebindTexture(slot int) {
	p := cp.pendingTex[slot]
	if cp.memReader == nil || p.width == 0 || p.height == 0 {
		return
	}
	length := int(p.width) * int(p.height) * texture.BytesPerTexel(p.format)
	raw := cp.memReader.ReadBytes(p.addr, length)
	cp.boundTextures[slot] = texture.New(p.format, raw, int(p.width), int(p.height))
}
