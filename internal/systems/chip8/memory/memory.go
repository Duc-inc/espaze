package memory

import "fmt"

// Size is the total addressable RAM of a CHIP-8 machine.
const Size = 4096

// FontStart is the conventional load address for the built-in hex font.
const FontStart = 0x50

// ProgramStart is where loaded ROMs are placed and where PC starts.
const ProgramStart = 0x200

// Memory models the flat 4KB address space shared by the CPU, font and ROM.
type Memory struct {
	data [Size]byte
}

// New allocates zeroed RAM with the built-in font pre-loaded.
func New() *Memory {
	m := &Memory{}
	m.LoadFont()
	return m
}

// Reset clears all RAM and reloads the font.
func (m *Memory) Reset() {
	m.data = [Size]byte{}
	m.LoadFont()
}

// LoadFont installs the built-in hex digit sprites at FontStart.
func (m *Memory) LoadFont() {
	copy(m.data[FontStart:], FontSet[:])
}

// LoadROM copies a program image into RAM starting at ProgramStart.
func (m *Memory) LoadROM(data []byte) error {
	if len(data) > Size-ProgramStart {
		return fmt.Errorf("memory: rom too large (%d bytes, max %d)", len(data), Size-ProgramStart)
	}
	copy(m.data[ProgramStart:], data)
	return nil
}

// Read returns the byte at addr, or 0 if out of range.
func (m *Memory) Read(addr uint16) byte {
	if int(addr) >= Size {
		return 0
	}
	return m.data[addr]
}

// Write stores a byte at addr, ignoring out-of-range addresses.
func (m *Memory) Write(addr uint16, value byte) {
	if int(addr) >= Size {
		return
	}
	m.data[addr] = value
}

// Snapshot returns a copy of the full RAM contents, for save states.
func (m *Memory) Snapshot() [Size]byte {
	return m.data
}

// Restore overwrites RAM with a previously captured snapshot.
func (m *Memory) Restore(snapshot [Size]byte) {
	m.data = snapshot
}
