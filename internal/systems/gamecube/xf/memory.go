// memory.go holds the XF unit's raw uploaded matrix data - the
// address space real hardware calls "XF memory", distinct from XF's
// control registers (registers.go). Games write into it via
// LOAD_XF_REG commands whose address falls in the memory range rather
// than the register range. Which addresses correspond to which
// matrix region (position/normal/texture) is left for implementation
// time, to be verified against YAGCD as each region is actually
// wired up rather than guessed upfront.
package xf

import "math"

// memSize is a working capacity for XF memory. Real hardware's exact
// total size is left to confirm against YAGCD once actual address
// ranges (position matrices, normal matrices, texture matrices) are
// wired up; this is sized generously in the meantime so nothing a
// real command stream writes gets silently dropped during
// development.
const memSize = 4096

// Memory holds the XF unit's raw uploaded matrix data, addressed the
// same way real LOAD_XF_REG commands address it: by 32-bit word
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

// ReadFloat32 reinterprets the word at addr as an IEEE-754 float32.
// An out-of-range address reads as 0.
func (m *Memory) ReadFloat32(addr uint16) float32 {
	if int(addr) >= len(m.words) {
		return 0
	}
	return math.Float32frombits(m.words[addr])
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
