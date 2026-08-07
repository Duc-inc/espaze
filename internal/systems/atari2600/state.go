package atari2600

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/Duc-inc/espaze/internal/systems/atari2600/riot"
	"github.com/Duc-inc/espaze/internal/systems/atari2600/tia"
	cpu "github.com/Duc-inc/espaze/internal/systems/nes/cpu"
)

// snapshot combines every component's own snapshot into one save
// state. The cartridge ROM itself is never included - LoadState
// assumes the same ROM is already loaded, exactly like every other
// core here.
type snapshot struct {
	CPU  cpu.Snapshot
	TIA  tia.Snapshot
	RIOT riot.Snapshot
}

// SaveState implements core.Core.
func (a *Atari2600) SaveState() ([]byte, error) {
	if !a.loaded {
		return nil, fmt.Errorf("atari2600: no rom loaded")
	}

	snap := snapshot{
		CPU:  a.proc.Snapshot(),
		TIA:  a.video.Snapshot(),
		RIOT: a.riotC.Snapshot(),
	}

	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(snap); err != nil {
		return nil, fmt.Errorf("atari2600: save state: %w", err)
	}
	return buf.Bytes(), nil
}

// LoadState implements core.Core.
func (a *Atari2600) LoadState(data []byte) error {
	if !a.loaded {
		return fmt.Errorf("atari2600: no rom loaded")
	}

	var snap snapshot
	if err := gob.NewDecoder(bytes.NewReader(data)).Decode(&snap); err != nil {
		return fmt.Errorf("atari2600: load state: %w", err)
	}

	a.proc.Restore(snap.CPU)
	a.video.Restore(snap.TIA)
	a.riotC.Restore(snap.RIOT)
	return nil
}
