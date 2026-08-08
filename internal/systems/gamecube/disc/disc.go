// Package disc parses raw GameCube disc images (.iso/.gcm dumps).
// Unlike Wii, Wii U, or 3DS media, GameCube discs are NOT encrypted -
// just a proprietary "mini-DVD" physical format that a raw image
// bypasses entirely, so parsing the header and boot executable is
// genuinely tractable from the binary layout alone (documented in
// public community references such as YAGCD, independent of any
// specific emulator's source). This project has no disc drive timing
// or physical-layer modeling - it just reads the image as a flat byte
// array, exactly like every real GameCube emulator's image-loading
// path does too.
package disc

import "fmt"

const (
	magicOffset    = 0x1C
	magicValue     = 0xC2339F3D
	gameNameOffset = 0x20
	gameNameLen    = 0x3E0
	dolOffsetAddr  = 0x0420
	fstOffsetAddr  = 0x0424
	fstSizeAddr    = 0x0428
)

// Header holds the fields this project reads from a disc image's
// fixed 0x440-byte header.
type Header struct {
	GameID    string // 6-byte disc ID + maker code
	GameName  string
	DOLOffset uint32
	FSTOffset uint32
	FSTSize   uint32
}

// ParseHeader reads a disc image's header and validates the magic
// word real GameCube discs (and dumps of them) always carry.
func ParseHeader(image []byte) (Header, error) {
	if len(image) < 0x440 {
		return Header{}, fmt.Errorf("disc: image too small to contain a header (%d bytes)", len(image))
	}
	magic := be32(image[magicOffset:])
	if magic != magicValue {
		return Header{}, fmt.Errorf("disc: bad magic word %#08x, want %#08x (not a GameCube image)", magic, magicValue)
	}

	name := image[gameNameOffset : gameNameOffset+gameNameLen]
	end := indexZero(name)

	return Header{
		GameID:    string(image[0:6]),
		GameName:  string(name[:end]),
		DOLOffset: be32(image[dolOffsetAddr:]),
		FSTOffset: be32(image[fstOffsetAddr:]),
		FSTSize:   be32(image[fstSizeAddr:]),
	}, nil
}

func indexZero(b []byte) int {
	for i, v := range b {
		if v == 0 {
			return i
		}
	}
	return len(b)
}

func be32(b []byte) uint32 {
	return uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
}
