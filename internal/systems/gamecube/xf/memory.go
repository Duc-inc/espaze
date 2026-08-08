// memory.go holds the XF unit's raw uploaded matrix and light data -
// the address space real hardware calls "XF memory", distinct from
// XF's control registers (registers.go). Public GameCube hardware
// notes (YAGCD chapter 5, "internal XF Memory") describe this
// occupying word addresses 0x0000-0x0fff, with control registers
// starting at 0x1000 (registers.go, state.go). Address sub-ranges
// YAGCD's own wording leaves ambiguous are documented as such below
// rather than asserted as verified fact.
package xf

import "math"

const (
	MemorySize = 0x1000

	// PosMatricesStart/End: 64 rows of 4 words each (256 words total).
	PosMatricesStart = 0x000
	PosMatricesEnd   = 0x100
	// NormalMatricesStart/End: 32 rows of 3 words each (96 words total).
	NormalMatricesStart = 0x400
	NormalMatricesEnd   = 0x460
	// DualTexMatricesStart/End: dual-texture transform matrices, the
	// same row shape as the position-matrix block.
	DualTexMatricesStart = 0x500
	DualTexMatricesEnd   = 0x600
	// LightsStart/End: light 0 begins at LightsStart; each of up to
	// MaxLights lights occupies 16 words (lighting.go covers the
	// simplified position+color model this project currently derives
	// from a light, not this raw block's full field layout).
	LightsStart = 0x600
	LightsEnd   = 0x680

	// 0x0100-0x03ff, 0x0460-0x04ff, and 0x0680-0x0fff are reserved/
	// unknown for this project - YAGCD's own wording around these
	// gaps is ambiguous, so no constant claims a meaning for them.
)

// memSize is a package-local alias kept for existing snapshot/test code.
const memSize = MemorySize

// Memory holds the XF unit's raw uploaded matrix/light data, addressed
// the same way real LOAD_XF_REG commands address it: by 32-bit word
// index, not byte offset. Values are stored as the raw bits the
// command stream carries and reinterpreted as float32 on read, since
// that's how real GX matrix uploads work - IEEE-754 floats carried as
// plain 32-bit words.
type Memory struct {
	words [memSize]uint32
}

// NewMemory returns an XF memory bank with every word zeroed.
func NewMemory() *Memory { return &Memory{} }

// Write stores one raw 32-bit word at addr, as decoded directly from
// a LOAD_XF_REG command. An out-of-range address is ignored rather
// than panicking, matching how this project's other GameCube register
// banks (e.g. gpu.CommandProcessor) tolerate unmapped addresses.
func (m *Memory) Write(addr uint16, word uint32) {
	if int(addr) >= len(m.words) {
		return
	}
	m.words[addr] = word
}

// Read returns the raw 32-bit word at addr. An out-of-range address
// reads as 0.
func (m *Memory) Read(addr uint16) uint32 {
	if int(addr) >= len(m.words) {
		return 0
	}
	return m.words[addr]
}

// ReadFloat32 reinterprets the word at addr as an IEEE-754 float32.
func (m *Memory) ReadFloat32(addr uint16) float32 {
	return math.Float32frombits(m.Read(addr))
}

// WriteFloat32 stores v at addr as its raw IEEE-754 bit pattern - the
// inverse of ReadFloat32, and how a real LOAD_XF_REG write of matrix
// data actually lands in memory.
func (m *Memory) WriteFloat32(addr uint16, v float32) {
	m.Write(addr, math.Float32bits(v))
}

// WritePosMatrix stores a 3x4 position matrix starting at addr: 12
// consecutive words, row-major - the inverse of ReadPosMatrix.
func (m *Memory) WritePosMatrix(addr uint16, mtx PosMatrix) {
	for row := 0; row < 3; row++ {
		for col := 0; col < 4; col++ {
			m.WriteFloat32(addr+uint16(row*4+col), mtx[row][col])
		}
	}
}

// WriteNormalMatrix stores a 3x3 normal matrix starting at addr: 9
// consecutive words, row-major - the inverse of ReadNormalMatrix.
func (m *Memory) WriteNormalMatrix(addr uint16, mtx NormalMatrix) {
	for row := 0; row < 3; row++ {
		for col := 0; col < 3; col++ {
			m.WriteFloat32(addr+uint16(row*3+col), mtx[row][col])
		}
	}
}

// ReadPosMatrix reads a 3x4 position matrix starting at addr: 12
// consecutive words, row-major, matching PosMatrix's own layout.
func (m *Memory) ReadPosMatrix(addr uint16) PosMatrix {
	var pm PosMatrix
	for row := 0; row < 3; row++ {
		for col := 0; col < 4; col++ {
			pm[row][col] = m.ReadFloat32(addr + uint16(row*4+col))
		}
	}
	return pm
}

// ReadNormalMatrix reads a 3x3 normal matrix starting at addr: 9
// consecutive words, row-major, matching NormalMatrix's own layout.
func (m *Memory) ReadNormalMatrix(addr uint16) NormalMatrix {
	var nm NormalMatrix
	for row := 0; row < 3; row++ {
		for col := 0; col < 3; col++ {
			nm[row][col] = m.ReadFloat32(addr + uint16(row*3+col))
		}
	}
	return nm
}
