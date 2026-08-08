// state.go holds Snapshot/Restore for the XF unit's state, matching
// the save-state pattern used across this project's other hardware
// packages (see e.g. internal/systems/gamecube/audio/state.go).
package xf

// State is the full XF unit state: word-addressed XF memory
// (memory.go) plus the decoded control registers (registers.go).
type State struct {
	Memory    Memory
	Registers Registers
}

// NewState returns a zeroed XF state with default decoded registers.
func NewState() *State {
	return &State{Registers: NewRegisters()}
}

// Load applies a LOAD_XF_REG-style transfer starting at addr: words
// below RegistersStart land in raw XF memory (memory.go); words from
// RegistersStart through RegistersEnd update decoded control
// registers (registers.go) - the same address split those two files
// document.
func (s *State) Load(addr uint16, words []uint32) {
	for i, word := range words {
		a := uint32(addr) + uint32(i)
		if a >= RegistersEnd {
			continue
		}
		if a < RegistersStart {
			s.Memory.Write(uint16(a), word)
			continue
		}
		s.Registers.Write(uint16(a), word)
	}
}

// Snapshot captures the XF unit's full state: every uploaded matrix
// word (Memory) and the active control registers (Registers). This
// package keeps Memory and Registers as separate types since callers
// may only need one or the other, so Snapshot is the pairing that
// represents "the whole XF unit" together.
type Snapshot struct {
	Memory    [MemorySize]uint32
	Registers Registers
}

// Snapshot captures m's current contents alongside the given
// registers.
func (m *Memory) Snapshot(regs Registers) Snapshot {
	return Snapshot{Memory: m.words, Registers: regs}
}

// Restore reinstates a previously captured Snapshot's memory contents
// into m and returns the registers half, for the caller to reinstate
// on its own Registers value.
func (m *Memory) Restore(s Snapshot) Registers {
	m.words = s.Memory
	return s.Registers
}
