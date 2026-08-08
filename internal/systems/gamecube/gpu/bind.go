package gpu

import (
	"math"

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

// XF pseudo-register addresses this project reserves within the same
// LOAD_XF_REG address space real matrix data (memory.go) occupies,
// set well above any address a matrix upload would realistically use
// in this project's own address scheme. Ambient/light colors are
// carried as IEEE-754 float32 bits, the same encoding every other
// LOAD_XF_REG float value uses; matrix index selectors are plain
// integers since they're array indices, not color/position data.
const (
	xfRegAmbientR = 3000
	xfRegAmbientG = 3001
	xfRegAmbientB = 3002

	xfRegPosMatrixIndex    = 3003
	xfRegNormalMatrixIndex = 3004

	// xfRegLightBase is the first of xf.MaxLights blocks of 8 words
	// each: pos.x, pos.y, pos.z, color.r, color.g, color.b, enabled,
	// and one reserved/unused word.
	xfRegLightBase  = 3100
	xfRegLightWords = 8
)

// applyXFRegisterWrite reacts to a LOAD_XF_REG write that already
// landed in cp.xfMemory, applying any binding-state side effect that
// address carries in addition to being stored as raw memory.
func (cp *CommandProcessor) applyXFRegisterWrite(addr uint16, val uint32) {
	switch {
	case addr == xfRegAmbientR:
		cp.ambient.Color.R = math.Float32frombits(val)
	case addr == xfRegAmbientG:
		cp.ambient.Color.G = math.Float32frombits(val)
	case addr == xfRegAmbientB:
		cp.ambient.Color.B = math.Float32frombits(val)
	case addr == xfRegPosMatrixIndex:
		cp.xfRegisters.PosMatrixIndex = uint16(val)
	case addr == xfRegNormalMatrixIndex:
		cp.xfRegisters.NormalMatrixIndex = uint16(val)
	case addr >= xfRegLightBase && addr < xfRegLightBase+xf.MaxLights*xfRegLightWords:
		cp.applyLightRegisterWrite(addr, val)
	}
}

func (cp *CommandProcessor) applyLightRegisterWrite(addr uint16, val uint32) {
	offset := addr - xfRegLightBase
	index := int(offset / xfRegLightWords)
	field := offset % xfRegLightWords
	l := &cp.lights[index]
	switch field {
	case 0:
		l.Position.X = math.Float32frombits(val)
	case 1:
		l.Position.Y = math.Float32frombits(val)
	case 2:
		l.Position.Z = math.Float32frombits(val)
	case 3:
		l.Color.R = math.Float32frombits(val)
	case 4:
		l.Color.G = math.Float32frombits(val)
	case 5:
		l.Color.B = math.Float32frombits(val)
	case 6:
		l.Enabled = val != 0
	}
}
