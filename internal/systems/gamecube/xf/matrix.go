// matrix.go holds the position (3x4, affine - no bottom row stored,
// matching real hardware's own space-saving layout) and normal (3x3)
// matrix types the XF memory holds, plus the vector and matrix math
// the transform pipeline (transform.go) needs.
package xf

import "math"

// Vec3 is a 3-component vector - vertex positions and normals in
// model or camera space.
type Vec3 struct {
	X, Y, Z float32
}

// Dot returns the dot product of v and other.
func (v Vec3) Dot(other Vec3) float32 {
	return v.X*other.X + v.Y*other.Y + v.Z*other.Z
}

// Length returns the Euclidean length of v.
func (v Vec3) Length() float32 {
	return float32(math.Sqrt(float64(v.Dot(v))))
}

// Normalize returns v scaled to unit length. A zero-length vector has
// no meaningful direction, so it's returned unchanged rather than
// dividing by zero.
func (v Vec3) Normalize() Vec3 {
	l := v.Length()
	if l == 0 {
		return v
	}
	return v.MulScalar(1 / l)
}

// Add returns the component-wise sum of v and other.
func (v Vec3) Add(other Vec3) Vec3 {
	return Vec3{X: v.X + other.X, Y: v.Y + other.Y, Z: v.Z + other.Z}
}

// Sub returns the component-wise difference of v and other.
func (v Vec3) Sub(other Vec3) Vec3 {
	return Vec3{X: v.X - other.X, Y: v.Y - other.Y, Z: v.Z - other.Z}
}

// MulScalar returns v scaled by s.
func (v Vec3) MulScalar(s float32) Vec3 {
	return Vec3{X: v.X * s, Y: v.Y * s, Z: v.Z * s}
}

// Vec4 is a 4-component homogeneous vector - the result of a
// projection multiply, before the perspective divide (transform.go)
// collapses W back out into normalized device coordinates.
type Vec4 struct {
	X, Y, Z, W float32
}

// PosMatrix is a real GameCube position/modelview matrix: 3 rows by 4
// columns. Real hardware never stores the implicit bottom row
// [0, 0, 0, 1], since every position matrix is affine (rotation,
// scale, translation only - no projection), so this type doesn't
// either.
type PosMatrix [3][4]float32

// NormalMatrix is a 3x3 matrix used to transform normal vectors
// separately from positions: correct handling of non-uniform scale
// needs the inverse-transpose of the position matrix's upper-left
// 3x3, not the position matrix itself. Deriving that inverse-
// transpose is left for whoever wires this to real matrix uploads;
// this type only carries the result and applies it.
type NormalMatrix [3][3]float32

// Mat4 is a full 4x4 matrix, used for the projection matrix - the one
// XF matrix that isn't affine, since perspective projection produces
// a non-trivial W component the position matrices never touch.
type Mat4 [4][4]float32

// IdentityPos returns a position matrix that leaves every vertex
// unchanged.
func IdentityPos() PosMatrix {
	return PosMatrix{
		{1, 0, 0, 0},
		{0, 1, 0, 0},
		{0, 0, 1, 0},
	}
}

// IdentityMat4 returns a 4x4 matrix that leaves every vertex
// unchanged (W comes out as 1).
func IdentityMat4() Mat4 {
	return Mat4{
		{1, 0, 0, 0},
		{0, 1, 0, 0},
		{0, 0, 1, 0},
		{0, 0, 0, 1},
	}
}

// MulVec3 transforms a position by this matrix, treating v as the
// homogeneous point (v.X, v.Y, v.Z, 1) implied by the matrix's
// missing bottom row.
func (m PosMatrix) MulVec3(v Vec3) Vec3 {
	return Vec3{
		X: m[0][0]*v.X + m[0][1]*v.Y + m[0][2]*v.Z + m[0][3],
		Y: m[1][0]*v.X + m[1][1]*v.Y + m[1][2]*v.Z + m[1][3],
		Z: m[2][0]*v.X + m[2][1]*v.Y + m[2][2]*v.Z + m[2][3],
	}
}

// MulVec3 transforms a normal by this matrix. Unlike
// PosMatrix.MulVec3, there is no translation column to add - normals
// are directions, not points.
func (m NormalMatrix) MulVec3(v Vec3) Vec3 {
	return Vec3{
		X: m[0][0]*v.X + m[0][1]*v.Y + m[0][2]*v.Z,
		Y: m[1][0]*v.X + m[1][1]*v.Y + m[1][2]*v.Z,
		Z: m[2][0]*v.X + m[2][1]*v.Y + m[2][2]*v.Z,
	}
}

// MulVec3 transforms a position through the full 4x4 matrix, treating
// v as the homogeneous point (v.X, v.Y, v.Z, 1). Used for the
// projection stage, where the resulting W is generally not 1 and
// still needs the perspective divide (transform.go) applied.
func (m Mat4) MulVec3(v Vec3) Vec4 {
	return Vec4{
		X: m[0][0]*v.X + m[0][1]*v.Y + m[0][2]*v.Z + m[0][3],
		Y: m[1][0]*v.X + m[1][1]*v.Y + m[1][2]*v.Z + m[1][3],
		Z: m[2][0]*v.X + m[2][1]*v.Y + m[2][2]*v.Z + m[2][3],
		W: m[3][0]*v.X + m[3][1]*v.Y + m[3][2]*v.Z + m[3][3],
	}
}
