package events

import (
	"context"
	"encoding/base64"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/Duc-inc/espaze/internal/emulation/video"
)

// FramePayload is the JSON shape delivered to the frontend for each frame.
// Pixels travel as base64 RGBA8888 so the frontend can paint them straight
// onto a canvas via ImageData without any further decoding.
type FramePayload struct {
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	PixelsBase64 string `json:"pixelsBase64"`
}

// FrameSink implements video.Sink by emitting a Wails runtime event.
type FrameSink struct {
	ctx context.Context
}

// NewFrameSink builds a sink bound to the app's Wails context.
func NewFrameSink(ctx context.Context) *FrameSink {
	return &FrameSink{ctx: ctx}
}

// PublishFrame implements video.Sink.
func (s *FrameSink) PublishFrame(fb *video.FrameBuffer) {
	wailsruntime.EventsEmit(s.ctx, FrameEvent, FramePayload{
		Width:        fb.Width,
		Height:       fb.Height,
		PixelsBase64: base64.StdEncoding.EncodeToString(fb.Pixels),
	})
}
