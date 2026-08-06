package engine

import (
	"fmt"
	"sync"

	"github.com/Duc-inc/espaze/internal/emulation/audio"
	"github.com/Duc-inc/espaze/internal/emulation/core"
	"github.com/Duc-inc/espaze/internal/emulation/input"
	"github.com/Duc-inc/espaze/internal/emulation/savestate"
	"github.com/Duc-inc/espaze/internal/emulation/video"
)

// Engine drives a single loaded Core at its target frame rate and fans its
// output out to whatever video/audio sinks are attached. It knows nothing
// about which system it's running - that's entirely up to the Core.
type Engine struct {
	mu     sync.Mutex
	core   core.Core
	system string
	status Status

	videoSink video.Sink
	audioSink audio.Sink

	pendingInput input.State

	stopCh chan struct{}
	doneCh chan struct{}
}

// New wraps a freshly created, not-yet-loaded core.
func New(c core.Core) *Engine {
	return &Engine{
		core:      c,
		system:    c.Metadata().ID,
		status:    StatusStopped,
		videoSink: video.NullSink{},
		audioSink: audio.NullSink{},
	}
}

// SetVideoSink attaches the consumer that receives every rendered frame.
func (e *Engine) SetVideoSink(sink video.Sink) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.videoSink = sink
}

// SetAudioSink attaches the consumer that receives every audio chunk.
func (e *Engine) SetAudioSink(sink audio.Sink) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.audioSink = sink
}

// LoadROM installs a ROM image into the core. Must be called before Start.
func (e *Engine) LoadROM(data []byte) error {
	return e.core.LoadROM(data)
}

// SetInput records the latest input state to be applied before the next frame.
func (e *Engine) SetInput(state input.State) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.pendingInput = state
}

func (e *Engine) inputSnapshot() input.State {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.pendingInput
}

// Status reports whether the loop is stopped, running or paused.
func (e *Engine) Status() Status {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.status
}

// Start begins running the core loop in the background.
func (e *Engine) Start() error {
	e.mu.Lock()
	if e.status != StatusStopped {
		e.mu.Unlock()
		return fmt.Errorf("engine: cannot start, already %s", e.status)
	}
	e.status = StatusRunning
	e.stopCh = make(chan struct{})
	e.doneCh = make(chan struct{})
	e.mu.Unlock()

	go e.run()
	return nil
}

// Pause freezes the loop without tearing it down; Resume continues it.
func (e *Engine) Pause() {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.status == StatusRunning {
		e.status = StatusPaused
	}
}

// Resume continues a paused loop.
func (e *Engine) Resume() {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.status == StatusPaused {
		e.status = StatusRunning
	}
}

// Stop halts the loop and waits for it to fully exit.
func (e *Engine) Stop() {
	e.mu.Lock()
	if e.status == StatusStopped {
		e.mu.Unlock()
		return
	}
	stopCh := e.stopCh
	doneCh := e.doneCh
	e.mu.Unlock()

	close(stopCh)
	<-doneCh

	e.mu.Lock()
	e.status = StatusStopped
	e.mu.Unlock()
}

// SaveState captures the running core into a portable, self-describing blob.
func (e *Engine) SaveState() ([]byte, error) {
	raw, err := e.core.SaveState()
	if err != nil {
		return nil, fmt.Errorf("engine: save state: %w", err)
	}
	return savestate.Encode(savestate.Wrap(e.system, raw))
}

// LoadState restores a blob previously produced by SaveState, refusing
// anything that was captured from a different system.
func (e *Engine) LoadState(data []byte) error {
	env, err := savestate.Decode(data)
	if err != nil {
		return fmt.Errorf("engine: load state: %w", err)
	}
	if env.System != e.system {
		return fmt.Errorf("engine: state is for system %q, engine runs %q", env.System, e.system)
	}
	return e.core.LoadState(env.CoreData)
}
