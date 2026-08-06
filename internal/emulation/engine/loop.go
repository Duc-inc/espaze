package engine

import "time"

// run is the background goroutine started by Start. It ticks the core at
// its declared frame rate, applies buffered input, then publishes the
// resulting picture and sound to whichever sinks are attached.
func (e *Engine) run() {
	defer close(e.doneCh)

	fps := e.core.Metadata().FramesPerSecond
	if fps <= 0 {
		fps = 60
	}
	ticker := time.NewTicker(time.Duration(float64(time.Second) / fps))
	defer ticker.Stop()

	for {
		select {
		case <-e.stopCh:
			return
		case <-ticker.C:
			if e.Status() != StatusRunning {
				continue
			}
			e.tick()
		}
	}
}

func (e *Engine) tick() {
	e.core.SetInput(e.inputSnapshot())

	if err := e.core.StepFrame(); err != nil {
		e.Pause()
		return
	}

	e.mu.Lock()
	videoSink := e.videoSink
	audioSink := e.audioSink
	e.mu.Unlock()

	videoSink.PublishFrame(e.core.FrameBuffer())

	if samples, rate := e.core.DrainAudio(); len(samples) > 0 {
		audioSink.PublishSamples(samples, rate)
	}
}
