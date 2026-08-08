// Package si implements the GameCube's Serial Interface: the real
// memory-mapped registers a game polls for controller data. Addresses
// and overall layout come from a public hardware register reference
// (YAGCD chapter 5, "SI - Serial Interface"). The exact bit meaning
// of the error/latch bits within InBufH isn't confidently resolved
// from that source, so only ERRSTAT (bit 31, confirmed) is modeled;
// this project also doesn't decode which specific bit means which
// controller button, since the source didn't give that mapping -
// SetChannel supplies raw controller bytes, matching what a real
// game's own controller library would parse from the same registers.
package si

const (
	Base = 0xCC006400
	Size = 0x100

	numChannels   = 4
	channelStride = 0x0C
	regInBufHigh  = 0x04 // input bytes 0-2 + ERRSTAT
	regInBufLow   = 0x08 // input bytes 4-7
)

// Channel is one controller's raw polled data - 8 input bytes, real
// hardware's own controller data-packet size.
type Channel struct {
	Connected bool
	Data      [8]byte
}

// SI holds the Serial Interface's per-channel controller data.
type SI struct {
	channels [numChannels]Channel
}

func New() *SI { return &SI{} }

// SetChannel supplies channel ch's current controller state. An
// out-of-range channel is ignored.
func (s *SI) SetChannel(ch int, c Channel) {
	if ch < 0 || ch >= numChannels {
		return
	}
	s.channels[ch] = c
}

// Read32 reads one SI register at a block-relative offset.
func (s *SI) Read32(offset uint32) uint32 {
	ch := int(offset / channelStride)
	if ch >= numChannels {
		return 0
	}
	c := &s.channels[ch]
	switch offset % channelStride {
	case regInBufHigh:
		var errStat uint32
		if !c.Connected {
			errStat = 1 << 31
		}
		return errStat | uint32(c.Data[0]&0x3F)<<16 | uint32(c.Data[1])<<8 | uint32(c.Data[2])
	case regInBufLow:
		return uint32(c.Data[4])<<24 | uint32(c.Data[5])<<16 | uint32(c.Data[6])<<8 | uint32(c.Data[7])
	default:
		return 0
	}
}

// Write32 accepts writes to the poll-control/command registers this
// project doesn't model yet (real hardware also has writable SI
// registers beyond the read-only input buffers above) - no effect,
// same "accepted but unmodeled" pattern this project uses elsewhere.
func (s *SI) Write32(offset uint32, val uint32) {}
