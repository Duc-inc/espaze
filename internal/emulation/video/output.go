package video

// Sink receives finished frames as they are produced by the engine loop.
// The app layer implements Sink to forward frames to the Wails frontend;
// tests can implement it to capture frames without any UI.
type Sink interface {
	PublishFrame(fb *FrameBuffer)
}

// NullSink discards every frame. Used when no consumer is attached yet.
type NullSink struct{}

func (NullSink) PublishFrame(*FrameBuffer) {}
