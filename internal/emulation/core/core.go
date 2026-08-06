package core

import (
	"github.com/Duc-inc/espaze/internal/emulation/input"
	"github.com/Duc-inc/espaze/internal/emulation/video"
)

// Core is the contract every emulated system must satisfy. The engine,
// the app bindings and the frontend only ever talk to this interface,
// never to a concrete system package directly - that's what lets new
// systems (Game Boy, NES, ...) plug in without touching existing code.
type Core interface {
	// Metadata describes this core (id, name, extensions, timing, screen size).
	Metadata() Metadata

	// LoadROM parses and installs the given ROM image, ready to run from reset.
	LoadROM(data []byte) error

	// Reset returns the core to its post-boot state, keeping the loaded ROM.
	Reset()

	// StepFrame advances emulation by exactly one output frame.
	StepFrame() error

	// FrameBuffer returns the picture produced by the most recent StepFrame.
	FrameBuffer() *video.FrameBuffer

	// DrainAudio returns and clears the audio samples produced since the
	// last call, alongside the sample rate they were generated at.
	DrainAudio() ([]int16, int)

	// SetInput applies the latest generic input state before the next frame.
	SetInput(state input.State)

	// SaveState serializes everything needed to resume emulation later.
	SaveState() ([]byte, error)

	// LoadState restores a state previously produced by SaveState.
	LoadState(data []byte) error
}
