package memory

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/Duc-inc/espaze/internal/systems/nes/ppu"
)

// gobEncode/gobDecode are tiny shared helpers each mapper's Snapshot/
// Restore uses to (de)serialize its own small register struct, so every
// mapper doesn't need to hand-roll the same boilerplate.
func gobEncode(v any) []byte {
	var buf bytes.Buffer
	_ = gob.NewEncoder(&buf).Encode(v)
	return buf.Bytes()
}

func gobDecode(data []byte, v any) error {
	if len(data) == 0 {
		return nil
	}
	return gob.NewDecoder(bytes.NewReader(data)).Decode(v)
}

// Mapper is the cartridge-specific logic for translating CPU/PPU
// addresses into PRG/CHR data - bank switching, mirroring overrides,
// and so on all live behind this interface so the rest of the NES
// package never needs to know which mapper number a game uses.
type Mapper interface {
	ReadPRG(addr uint16) byte
	WritePRG(addr uint16, v byte)
	ReadCHR(addr uint16) byte
	WriteCHR(addr uint16, v byte)
	Mirroring() ppu.MirrorMode

	// Snapshot/Restore capture only the mapper's own mutable registers
	// (bank selection etc) - never PRG/CHR ROM data, which LoadState
	// callers are expected to already have from the loaded cartridge.
	Snapshot() []byte
	Restore(data []byte) error
}

// NewMapper builds the right Mapper for cart's declared mapper number.
// The four implemented here (NROM, MMC1, UxROM, CNROM) cover a large
// share of the NES library on their own; more can be added the same
// way without touching anything else.
func NewMapper(cart *Cartridge) (Mapper, error) {
	switch cart.Mapper {
	case 0:
		return newNROM(cart), nil
	case 1:
		return newMMC1(cart), nil
	case 2:
		return newUxROM(cart), nil
	case 3:
		return newCNROM(cart), nil
	default:
		return nil, fmt.Errorf("nes: unsupported mapper %d", cart.Mapper)
	}
}
