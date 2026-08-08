package gpu

import "github.com/Duc-inc/espaze/internal/systems/gamecube/xf"

// Execute decodes and processes an entire GX command stream in one
// call - real hardware processes it continuously as the CPU feeds the
// FIFO; this project doesn't model that timing, just the resulting
// command semantics.
func (cp *CommandProcessor) Execute(stream []byte) {
	pos := 0
	for pos < len(stream) {
		opcode := stream[pos]
		pos++

		switch {
		case opcode == cmdNop:
			// no operand
		case opcode == cmdLoadCPReg:
			if pos+5 > len(stream) {
				return
			}
			reg := stream[pos]
			val := be32(stream[pos+1:])
			cp.cpRegs[reg] = val
			pos += 5
		case opcode == cmdLoadXFReg:
			if pos+6 > len(stream) {
				return
			}
			reg := be16(stream[pos:])
			val := be32(stream[pos+2:])
			cp.xfMemory.Write(reg, val)
			cp.applyXFRegisterWrite(reg, val)
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
			verts := make([]Vertex, 0, count)
			for i := 0; i < count; i++ {
				if pos+vertexBytes > len(stream) {
					return
				}
				verts = append(verts, cp.transformVertex(decodeVertex(stream[pos:])))
				pos += vertexBytes
			}
			cp.emitTriangles(opcode, verts)
		default:
			return // unrecognized opcode: stop rather than misparse the rest of the stream
		}
	}
}

func decodeVertex(b []byte) Vertex {
	return Vertex{
		X: int16(be16(b)), Y: int16(be16(b[2:])), Z: int16(be16(b[4:])),
		NX: int16(be16(b[6:])), NY: int16(be16(b[8:])), NZ: int16(be16(b[10:])),
		U: int16(be16(b[12:])), V: int16(be16(b[14:])),
		R: b[16], G: b[17], B: b[18], A: b[19],
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
	screen := xf.TransformPosition(model, cp.xfMemory, cp.xfRegisters)
	viewSpace := xf.ViewSpacePosition(model, cp.xfMemory, cp.xfRegisters)

	modelNormal := xf.Vec3{X: float32(v.NX), Y: float32(v.NY), Z: float32(v.NZ)}
	normal := xf.TransformNormal(modelNormal, cp.xfMemory, cp.xfRegisters)
	material := xf.LightColor{R: float32(v.R) / 255, G: float32(v.G) / 255, B: float32(v.B) / 255}
	lit := xf.Illuminate(viewSpace, normal, material, cp.ambient, cp.lights[:])

	v.X, v.Y, v.Z = int16(screen.X), int16(screen.Y), int16(screen.Z)
	v.R, v.G, v.B = byte(lit.R*255), byte(lit.G*255), byte(lit.B*255)
	return v
}

// emitTriangles converts a decoded vertex list into flat triangles
// according to the primitive type - QUADS split into 2 triangles
// each, strips/fans expanded the standard way; LINES/LINE_STRIP/
// POINTS are parsed (so the stream stays in sync) but produce no
// triangles, since this project's rasterizer only draws filled
// triangles.
func (cp *CommandProcessor) emitTriangles(opcode byte, verts []Vertex) {
	tex := cp.boundTexture
	switch opcode {
	case cmdDrawQuads:
		for i := 0; i+3 < len(verts); i += 4 {
			cp.pendingTriangles = append(cp.pendingTriangles,
				Triangle{verts[i], verts[i+1], verts[i+2], tex},
				Triangle{verts[i], verts[i+2], verts[i+3], tex},
			)
		}
	case cmdDrawTriangles:
		for i := 0; i+2 < len(verts); i += 3 {
			cp.pendingTriangles = append(cp.pendingTriangles, Triangle{verts[i], verts[i+1], verts[i+2], tex})
		}
	case cmdDrawTriStrip:
		for i := 0; i+2 < len(verts); i++ {
			cp.pendingTriangles = append(cp.pendingTriangles, Triangle{verts[i], verts[i+1], verts[i+2], tex})
		}
	case cmdDrawTriFan:
		for i := 1; i+1 < len(verts); i++ {
			cp.pendingTriangles = append(cp.pendingTriangles, Triangle{verts[0], verts[i], verts[i+1], tex})
		}
	}
}

func be16(b []byte) uint16 { return uint16(b[0])<<8 | uint16(b[1]) }
func be32(b []byte) uint32 {
	return uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
}
