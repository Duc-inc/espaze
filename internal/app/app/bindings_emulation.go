package app

import (
	"encoding/base64"
	"fmt"
	"os"
	"time"

	"github.com/Duc-inc/espaze/internal/app/events"
	"github.com/Duc-inc/espaze/internal/emulation/core"
	"github.com/Duc-inc/espaze/internal/emulation/engine"
	"github.com/Duc-inc/espaze/internal/emulation/input"
)

// LaunchGame stops whatever is currently running, then boots the given
// game's core, loads its ROM and starts the engine loop. Frames and audio
// stream to the frontend over Wails events as soon as it's running.
func (a *App) LaunchGame(id string) error {
	g, ok := a.lib.Get(id)
	if !ok {
		return fmt.Errorf("app: game %q not found", id)
	}

	a.stopEngineIfRunning()

	c, err := core.New(g.System)
	if err != nil {
		return fmt.Errorf("app: create core: %w", err)
	}

	data, err := os.ReadFile(g.Path)
	if err != nil {
		return fmt.Errorf("app: read rom %s: %w", g.Path, err)
	}

	eng := engine.New(c)
	eng.SetVideoSink(events.NewFrameSink(a.ctx))
	eng.SetAudioSink(events.NewAudioSink(a.ctx))

	if err := eng.LoadROM(data); err != nil {
		return fmt.Errorf("app: load rom: %w", err)
	}
	if err := eng.Start(); err != nil {
		return fmt.Errorf("app: start engine: %w", err)
	}

	a.engine = eng
	a.activeGameID = id
	a.sessionStart = time.Now()
	return nil
}

// StopGame halts the running engine and records the play session.
func (a *App) StopGame() error {
	a.stopEngineIfRunning()
	return nil
}

func (a *App) stopEngineIfRunning() {
	if a.engine == nil {
		return
	}
	a.engine.Stop()
	if a.activeGameID != "" {
		duration := time.Since(a.sessionStart)
		if err := a.lib.RecordSession(a.activeGameID, duration); err != nil {
			a.logError("record play session", err)
		}
	}
	a.engine = nil
	a.activeGameID = ""
}

// PauseGame freezes the running engine without stopping it.
func (a *App) PauseGame() error {
	if a.engine == nil {
		return fmt.Errorf("app: no game running")
	}
	a.engine.Pause()
	return nil
}

// ResumeGame continues a paused engine.
func (a *App) ResumeGame() error {
	if a.engine == nil {
		return fmt.Errorf("app: no game running")
	}
	a.engine.Resume()
	return nil
}

// SendInput applies a generic button bitmask to the running core.
func (a *App) SendInput(buttons uint32) error {
	if a.engine == nil {
		return fmt.Errorf("app: no game running")
	}
	a.engine.SetInput(input.State{Buttons: buttons})
	return nil
}

// SaveState captures the running game's state as a base64 string the
// frontend can store (e.g. in local storage or a save-slot picker).
func (a *App) SaveState() (string, error) {
	if a.engine == nil {
		return "", fmt.Errorf("app: no game running")
	}
	data, err := a.engine.SaveState()
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

// LoadState restores a base64 state previously produced by SaveState.
func (a *App) LoadState(encoded string) error {
	if a.engine == nil {
		return fmt.Errorf("app: no game running")
	}
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return fmt.Errorf("app: decode state: %w", err)
	}
	return a.engine.LoadState(data)
}
