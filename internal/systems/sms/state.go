package sms

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/Duc-inc/espaze/internal/systems/sms/cpu"
	"github.com/Duc-inc/espaze/internal/systems/sms/memory"
	"github.com/Duc-inc/espaze/internal/systems/sms/psg"
	"github.com/Duc-inc/espaze/internal/systems/sms/vdp"
)

// snapshot combines every component's own snapshot into one save
// state. The cartridge ROM itself is never included - LoadState
// assumes the same ROM is already loaded, exactly like every other
// core here.
type snapshot struct {
	CPU       cpu.Snapshot
	Bus       memory.Snapshot
	VDP       vdp.Snapshot
	PSG       psg.Snapshot
	PausePrev bool
}

// SaveState implements core.Core.
func (s *SMS) SaveState() ([]byte, error) {
	if !s.loaded {
		return nil, fmt.Errorf("sms: no rom loaded")
	}

	snap := snapshot{
		CPU:       s.proc.Snapshot(),
		Bus:       s.bus.Snapshot(),
		VDP:       s.video.Snapshot(),
		PSG:       s.sound.Snapshot(),
		PausePrev: s.pausePrev,
	}

	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(snap); err != nil {
		return nil, fmt.Errorf("sms: save state: %w", err)
	}
	return buf.Bytes(), nil
}

// LoadState implements core.Core.
func (s *SMS) LoadState(data []byte) error {
	if !s.loaded {
		return fmt.Errorf("sms: no rom loaded")
	}

	var snap snapshot
	if err := gob.NewDecoder(bytes.NewReader(data)).Decode(&snap); err != nil {
		return fmt.Errorf("sms: load state: %w", err)
	}

	s.proc.Restore(snap.CPU)
	s.bus.Restore(snap.Bus)
	s.video.Restore(snap.VDP)
	s.sound.Restore(snap.PSG)
	s.pausePrev = snap.PausePrev
	return nil
}
