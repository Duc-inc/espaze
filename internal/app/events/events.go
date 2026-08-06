package events

// Event names emitted to the frontend over the Wails runtime event bus.
// The frontend subscribes to these with runtime.EventsOn in JS.
const (
	FrameEvent  = "emulation:frame"
	AudioEvent  = "emulation:audio"
	StatusEvent = "emulation:status"
)
