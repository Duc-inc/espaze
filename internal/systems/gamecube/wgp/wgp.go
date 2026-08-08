// Package wgp implements the GameCube's Write Gather Pipe: the real
// mechanism the CPU uses to stream GX command bytes into the GP
// command FIFO without managing a write pointer itself. Confirmed
// facts (looked up directly for this addition, YAGCD): the CPU writes
// data of any size to a fixed address, 0xCC008000, and every 32 bytes
// gathered triggers one burst transaction into the GP FIFO - real GX
// software (e.g. GX_Flush) pads a command list to a 32-byte multiple
// for exactly this reason, since a partial tail stays buffered here
// until topped off.
package wgp

const (
	Base = 0xCC008000
	Size = 0x20 // the 32-byte gather buffer itself
)

// WGP gathers bytes written to Base into 32-byte bursts.
type WGP struct {
	buf    []byte
	bursts [][]byte
}

func New() *WGP { return &WGP{} }

// StreamByte appends one byte (real hardware accepts stores of any
// size; this project's memory.Bus routes every store through Write8,
// so byte-at-a-time is this package's only entry point) and flushes
// any complete 32-byte bursts.
func (w *WGP) StreamByte(v byte) {
	w.buf = append(w.buf, v)
	for len(w.buf) >= Size {
		burst := make([]byte, Size)
		copy(burst, w.buf[:Size])
		w.bursts = append(w.bursts, burst)
		w.buf = w.buf[Size:]
	}
}

// DrainBursts returns and clears every complete 32-byte burst
// gathered since the last call, in the order they completed.
func (w *WGP) DrainBursts() [][]byte {
	out := w.bursts
	w.bursts = nil
	return out
}

// Read32/Write32 satisfy memory.Peripheral for Attach; real reads and
// word-granularity writes aren't meaningful for a streaming write-only
// port, so Bus routes every byte through StreamByte instead (see
// memory.Bus's StreamPeripheral handling) - these exist only so WGP
// can be attached at all.
func (w *WGP) Read32(offset uint32) uint32 { return 0 }
func (w *WGP) Write32(offset uint32, val uint32) {
	w.StreamByte(byte(val >> 24))
	w.StreamByte(byte(val >> 16))
	w.StreamByte(byte(val >> 8))
	w.StreamByte(byte(val))
}
