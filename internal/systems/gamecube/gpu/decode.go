package gpu

import "github.com/Duc-inc/espaze/internal/systems/gamecube/xf"

// maxDisplayListDepth caps how deeply a CALL_DISPLAY_LIST command may
// nest (a display list calling another, and so on) - a real command
// stream has no reason to nest more than a couple of levels deep;
// this exists purely to keep a corrupt or cyclic display list from
// recursing forever.
const maxDisplayListDepth = 8

// Execute decodes and processes an entire GX command stream in one
// call - real hardware processes it continuously as the CPU feeds the
// FIFO; this project doesn't model that timing, just the resulting
// command semantics.
func (cp *CommandProcessor) Execute(stream []byte) {
	cp.execute(stream, 0)
}

func (cp *CommandProcessor) execute(stream []byte, depth int) {
	pos := 0
	for pos < len(stream) {
		opcode := stream[pos]
		pos++

		switch {
		case opcode == cmdNop:
			// no operand
		case opcode == cmdCallDisplayList:
			if pos+8 > len(stream) {
				return
			}
			addr := be32(stream[pos:])
			size := be32(stream[pos+4:])
			pos += 8
			if cp.memReader != nil && depth < maxDisplayListDepth {
				cp.execute(cp.memReader.ReadBytes(addr, int(size)), depth+1)
			}
		case opcode == cmdLoadCPReg:
			if pos+5 > len(stream) {
				return
			}
			reg := stream[pos]
			val := be32(stream[pos+1:])
			cp.cpRegs[reg] = val
			cp.applyCPRegisterWrite(reg, val)
			pos += 5
		case opcode == cmdLoadXFReg:
			if pos+6 > len(stream) {
				return
			}
			reg := be16(stream[pos:])
			val := be32(stream[pos+2:])
			cp.xfState.Load(reg, []uint32{val})
			if idx, ok := xf.LightIndexForAddr(reg); ok {
				cp.lights[idx] = cp.xfState.Memory.ReadLight(idx)
			}
			pos += 6
		case opcode == cmdLoadBPReg:
			if pos+4 > len(stream) {
				return
			}
			word := be32(stream[pos:])
			reg := byte(word >> 24)
			cp.bpRegs[reg] = word & 0x00FFFFFF
			cp.applyBPRegisterWrite(reg)
			pos += 4
		case opcode >= cmdDrawQuads:
			if pos+2 > len(stream) {
				return
			}
			count := int(be16(stream[pos:]))
			pos += 2
			vcd := cp.currentVCD()
			vb := vcd.byteWidth()
			verts := make([]Vertex, 0, count)
			for i := 0; i < count; i++ {
				if pos+vb > len(stream) {
					return
				}
				v := cp.decodeDynamicVertex(vcd, stream[pos:])
				verts = append(verts, cp.transformVertex(v))
				pos += vb
			}
			cp.emitTriangles(opcode, verts)
		default:
			return // unrecognized opcode: stop rather than misparse the rest of the stream
		}
	}
}

// transformVertex runs a decoded vertex's position through the xf
// package's transform pipeline (position matrix -> projection ->
// perspective divide -> viewport), replacing X/Y/Z with the result,
// and runs its normal through the lighting pipeline (transformed
// normal + current ambient/lights, xf.Illuminate), replacing R/G/B
// with the lit color. With the default white ambient and no lights
// enabled (New), Illuminate returns the vertex's own color unchanged.
func (cp *CommandProcessor) transformVertex(v Vertex) Vertex {
	model := xf.Vec3{X: float32(v.X), Y: float32(v.Y), Z: float32(v.Z)}
	screen := xf.TransformPosition(model, &cp.xfState.Memory, cp.xfState.Registers)
	viewSpace := xf.ViewSpacePosition(model, &cp.xfState.Memory, cp.xfState.Registers)

	modelNormal := xf.Vec3{X: float32(v.NX), Y: float32(v.NY), Z: float32(v.NZ)}
	normal := xf.TransformNormal(modelNormal, &cp.xfState.Memory, cp.xfState.Registers)
	material := xf.LightColor{R: float32(v.R) / 255, G: float32(v.G) / 255, B: float32(v.B) / 255}
	lit := xf.Illuminate(viewSpace, normal, material, cp.ambient, cp.lights[:])

	v.X, v.Y, v.Z = int16(screen.X), int16(screen.Y), int16(screen.Z)
	v.R, v.G, v.B = byte(lit.R*255), byte(lit.G*255), byte(lit.B*255)

	if cp.xfState.Registers.NumTexGens > 0 {
		if u, uv, ok := xf.GenerateTexCoord(0, model, modelNormal, &cp.xfState.Memory, cp.xfState.Registers); ok {
			v.U, v.V = int16(u), int16(uv)
		}
	}
	return v
}

// emitTriangles converts a decoded vertex list into flat triangles
// according to the primitive type - QUADS split into 2 triangles
// each, strips/fans expanded the standard way; LINES/LINE_STRIP/
// POINTS are parsed (so the stream stays in sync) but produce no
// triangles, since this project's rasterizer only draws filled
// triangles.
func (cp *CommandProcessor) emitTriangles(opcode byte, verts []Vertex) {
	texs, ops := cp.boundTextures, cp.tevOps
	tri := func(v0, v1, v2 Vertex) Triangle {
		return Triangle{V0: v0, V1: v1, V2: v2, Textures: texs, TEVOps: ops}
	}
	switch opcode {
	case cmdDrawQuads:
		for i := 0; i+3 < len(verts); i += 4 {
			cp.pendingTriangles = append(cp.pendingTriangles,
				tri(verts[i], verts[i+1], verts[i+2]),
				tri(verts[i], verts[i+2], verts[i+3]),
			)
		}
	case cmdDrawTriangles:
		for i := 0; i+2 < len(verts); i += 3 {
			cp.pendingTriangles = append(cp.pendingTriangles, tri(verts[i], verts[i+1], verts[i+2]))
		}
	case cmdDrawTriStrip:
		for i := 0; i+2 < len(verts); i++ {
			cp.pendingTriangles = append(cp.pendingTriangles, tri(verts[i], verts[i+1], verts[i+2]))
		}
	case cmdDrawTriFan:
		for i := 1; i+1 < len(verts); i++ {
			cp.pendingTriangles = append(cp.pendingTriangles, tri(verts[0], verts[i], verts[i+1]))
		}
	}
}

func be16(b []byte) uint16 { return uint16(b[0])<<8 | uint16(b[1]) }
func be32(b []byte) uint32 {
	return uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
}
