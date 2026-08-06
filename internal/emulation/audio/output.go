package audio

// Sink receives PCM samples drained from a core once per frame.
// Mirrors video.Sink so the engine loop can treat both streams the same way.
type Sink interface {
	PublishSamples(samples []int16, sampleRate int)
}

// NullSink discards every sample. Used when no consumer is attached yet.
type NullSink struct{}

func (NullSink) PublishSamples([]int16, int) {}
