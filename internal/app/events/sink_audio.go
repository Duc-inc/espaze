package events

import (
	"context"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// AudioPayload is the JSON shape delivered to the frontend for each chunk
// of PCM audio produced by the running core.
type AudioPayload struct {
	SampleRate int     `json:"sampleRate"`
	Samples    []int16 `json:"samples"`
}

// AudioSink implements audio.Sink by emitting a Wails runtime event.
type AudioSink struct {
	ctx context.Context
}

// NewAudioSink builds a sink bound to the app's Wails context.
func NewAudioSink(ctx context.Context) *AudioSink {
	return &AudioSink{ctx: ctx}
}

// PublishSamples implements audio.Sink.
func (s *AudioSink) PublishSamples(samples []int16, sampleRate int) {
	wailsruntime.EventsEmit(s.ctx, AudioEvent, AudioPayload{
		SampleRate: sampleRate,
		Samples:    samples,
	})
}
