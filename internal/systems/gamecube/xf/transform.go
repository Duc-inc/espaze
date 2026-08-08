// transform.go holds the actual per-vertex transform pipeline: model-
// space vertex -> multiply by the selected position matrix -> multiply
// by the projection matrix -> perspective divide -> viewport mapping
// -> screen-space vertex, the form gpu/render.go's rasterizer already
// consumes.
package xf

// ViewSpacePosition applies only the active position matrix (selected
// via Registers.PositionMatrixAddr, matrix_select.go) - the stage
// before projection. Lighting (Illuminate, lighting.go) operates in
// this same camera/view space, not the final screen space
// TransformPosition produces, so it's exposed on its own here.
func ViewSpacePosition(pos Vec3, mem *Memory, regs Registers) Vec3 {
	return mem.ReadPosMatrix(regs.PositionMatrixAddr()).MulVec3(pos)
}

// TransformPosition carries a vertex from model space to screen space
// through every stage real XF hardware applies in order: the active
// position matrix, the projection matrix, the perspective divide, and
// finally the viewport mapping.
func TransformPosition(pos Vec3, mem *Memory, regs Registers) Vec3 {
	viewSpace := ViewSpacePosition(pos, mem, regs)
	clip := regs.Projection.Matrix().MulVec3(viewSpace)
	ndc := PerspectiveDivide(clip)
	return regs.Viewport.Apply(ndc)
}

// PerspectiveDivide collapses a clip-space Vec4 into normalized
// device coordinates by dividing X/Y/Z by W - the step that actually
// makes distant objects appear smaller. A zero W (the degenerate case
// a malformed or not-yet-projected matrix would produce) leaves
// X/Y/Z unchanged rather than dividing by zero.
func PerspectiveDivide(clip Vec4) Vec3 {
	if clip.W == 0 {
		return Vec3{X: clip.X, Y: clip.Y, Z: clip.Z}
	}
	invW := 1 / clip.W
	return Vec3{X: clip.X * invW, Y: clip.Y * invW, Z: clip.Z * invW}
}

// Apply maps a normalized-device-coordinate position onto screen
// pixel coordinates using this viewport's scale and offset. Z passes
// through unchanged - it still feeds the rasterizer's own depth test
// (gpu/render.go), not a screen axis.
func (vp Viewport) Apply(ndc Vec3) Vec3 {
	return Vec3{
		X: ndc.X*vp.ScaleX + vp.OffsetX,
		Y: ndc.Y*vp.ScaleY + vp.OffsetY,
		Z: ndc.Z,
	}
}

// GenerateTexCoord computes texture generator n's output (S,T) by
// multiplying a source vector through n's selected texture matrix
// (TextureMatrixAddr, matrix_select.go) - real XF TexGen's "regular"
// mode. Only SourceGeom (the vertex's own model-space position) and
// SourceNormal (its model-space normal) are modeled as source rows;
// any other source, or a non-regular TexGenType (emboss, color-as-
// texcoord), reports ok=false since this pipeline doesn't carry the
// other attribute rows (colors, binormals, tex0-7) TexGen can also
// read from. The projective divide real hardware applies to STQ-mode
// outputs isn't modeled either - Q is discarded, only S/T returned.
func GenerateTexCoord(n int, position, normal Vec3, mem *Memory, regs Registers) (u, v float32, ok bool) {
	if n < 0 || n >= len(regs.TexGen) {
		return 0, 0, false
	}
	tg := regs.TexGen[n]
	if tg.Type() != TexGenRegular {
		return 0, 0, false
	}
	var source Vec3
	switch tg.SourceRow() {
	case SourceGeom:
		source = position
	case SourceNormal:
		source = normal
	default:
		return 0, 0, false
	}
	out := mem.ReadPosMatrix(regs.TextureMatrixAddr(n)).MulVec3(source)
	return out.X, out.Y, true
}

// TransformNormal carries a normal vector from model space to camera
// space using the active normal matrix (selected via
// Registers.NormalMatrixAddr, matrix_select.go - the same GeometryIndex
// the position matrix uses), renormalizing afterward since a non-
// uniform scale in the matrix can change the vector's length even
// when the direction it should represent is still correct.
func TransformNormal(normal Vec3, mem *Memory, regs Registers) Vec3 {
	return mem.ReadNormalMatrix(regs.NormalMatrixAddr()).MulVec3(normal).Normalize()
}
