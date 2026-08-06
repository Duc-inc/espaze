package audio

// Buffer is a simple growable queue of signed 16-bit PCM samples,
// mono, produced by a core between two frames and drained by the engine.
type Buffer struct {
	samples []int16
}

// NewBuffer creates an empty sample queue.
func NewBuffer() *Buffer {
	return &Buffer{samples: make([]int16, 0, 1024)}
}

// Write appends samples produced during emulation of the current frame.
func (b *Buffer) Write(samples []int16) {
	b.samples = append(b.samples, samples...)
}

// Drain returns and clears every buffered sample.
func (b *Buffer) Drain() []int16 {
	out := b.samples
	b.samples = make([]int16, 0, 1024)
	return out
}

// Len reports how many samples are currently queued.
func (b *Buffer) Len() int {
	return len(b.samples)
}
