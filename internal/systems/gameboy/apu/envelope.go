package apu

// volumeEnvelope fades a channel's volume up or down over time, driven
// at 64Hz by the frame sequencer. Shared by channels 1, 2 and 4.
type volumeEnvelope struct {
	raw           byte // last value written to NRx2, for DAC-enabled checks
	initialVolume byte
	increasing    bool
	period        byte
	volume        byte
	timer         byte
}

func (e *volumeEnvelope) writeRegister(v byte) {
	e.raw = v
	e.initialVolume = v >> 4
	e.increasing = v&0x08 != 0
	e.period = v & 0x07
}

// dacEnabled mirrors real hardware: the DAC (and so the whole channel)
// is off if both the initial volume and the envelope direction are zero.
func (e *volumeEnvelope) dacEnabled() bool {
	return e.raw&0xF8 != 0
}

func (e *volumeEnvelope) trigger() {
	e.volume = e.initialVolume
	e.timer = e.period
}

func (e *volumeEnvelope) tick() {
	if e.period == 0 {
		return
	}
	if e.timer > 0 {
		e.timer--
	}
	if e.timer != 0 {
		return
	}
	e.timer = e.period

	delta := 1
	if !e.increasing {
		delta = -1
	}
	if newVolume := int(e.volume) + delta; newVolume >= 0 && newVolume <= 15 {
		e.volume = byte(newVolume)
	}
}
