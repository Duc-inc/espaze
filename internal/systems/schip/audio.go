package schip

import "math"

// beepFrequency and beepAmplitude describe the square wave played back
// while the sound timer is counting down.
const (
	beepFrequency = 440.0
	beepAmplitude = 3000
)

// synthesizeAudio appends exactly one frame's worth of PCM samples to the
// audio buffer: a square wave while the sound timer is active, silence
// otherwise.
func (s *Schip) synthesizeAudio() {
	samplesPerFrame := sampleRate / int(Metadata().FramesPerSecond)
	samples := make([]int16, samplesPerFrame)

	if s.sound.Active() {
		for i := range samples {
			t := float64(s.beepPos+i) / float64(sampleRate)
			if math.Sin(2*math.Pi*beepFrequency*t) >= 0 {
				samples[i] = beepAmplitude
			} else {
				samples[i] = -beepAmplitude
			}
		}
		s.beepPos += samplesPerFrame
	} else {
		s.beepPos = 0
	}

	s.audio.Write(samples)
}
