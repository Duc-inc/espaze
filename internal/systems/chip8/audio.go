package chip8

import "math"

// beepFrequency and beepAmplitude describe the square wave played back
// while the sound timer is counting down - CHIP-8 only ever needs one tone.
const (
	beepFrequency = 440.0
	beepAmplitude = 3000
)

// synthesizeAudio appends exactly one frame's worth of PCM samples to the
// audio buffer: a square wave while the sound timer is active, silence
// otherwise. Always emitting a full frame keeps the output stream steady.
func (c *Chip8) synthesizeAudio() {
	samplesPerFrame := sampleRate / int(Metadata().FramesPerSecond)
	samples := make([]int16, samplesPerFrame)

	if c.sound.Active() {
		for i := range samples {
			t := float64(c.beepPos+i) / float64(sampleRate)
			if math.Sin(2*math.Pi*beepFrequency*t) >= 0 {
				samples[i] = beepAmplitude
			} else {
				samples[i] = -beepAmplitude
			}
		}
		c.beepPos += samplesPerFrame
	} else {
		c.beepPos = 0
	}

	c.audio.Write(samples)
}
