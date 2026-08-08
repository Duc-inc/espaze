// state.go holds Snapshot/Restore for the XF unit's state, matching
// the save-state pattern used across this project's other hardware
// packages (see e.g. internal/systems/gamecube/audio/state.go).
package xf

// Snapshot captures the XF unit's full state: every uploaded matrix
// word (Memory) and the active control registers (Registers). This
// package keeps Memory and Registers as separate types since callers
// may only need one or the other, so Snapshot is the pairing that
// represents "the whole XF unit" together.
type Snapshot struct {
	Memory    [memSize]uint32
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
