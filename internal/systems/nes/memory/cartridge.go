package memory

import (
	"errors"
	"fmt"

	"github.com/Duc-inc/espaze/internal/systems/nes/ppu"
)

const (
	prgBankSize = 16 * 1024
	chrBankSize = 8 * 1024
	headerSize  = 16
	trainerSize = 512
)

// Cartridge holds a parsed iNES ROM image: PRG/CHR data plus the header
// facts a Mapper needs to interpret them (mapper number, fixed
// mirroring, whether CHR is actually RAM since the cart shipped none).
type Cartridge struct {
	PRG      []byte
	CHR      []byte
	ChrIsRAM bool
	Mapper   byte
	Mirror   ppu.MirrorMode
	Battery  bool
}

// ParseCartridge reads an iNES-format ROM image (the header iNES 2.0
// also starts with, so this covers both for the fields used here).
func ParseCartridge(data []byte) (*Cartridge, error) {
	if len(data) < headerSize || string(data[0:4]) != "NES\x1A" {
		return nil, errors.New("nes: not a valid iNES ROM (missing NES\\x1A header)")
	}

	prgUnits := int(data[4])
	chrUnits := int(data[5])
	flags6 := data[6]
	flags7 := data[7]

	mapper := (flags7 & 0xF0) | (flags6 >> 4)
	mirror := ppu.MirrorHorizontal
	if flags6&0x08 != 0 {
		mirror = ppu.MirrorFourScreen
	} else if flags6&0x01 != 0 {
		mirror = ppu.MirrorVertical
	}
	battery := flags6&0x02 != 0

	offset := headerSize
	if flags6&0x04 != 0 { // trainer present, skip it
		offset += trainerSize
	}

	prgSize := prgUnits * prgBankSize
	if offset+prgSize > len(data) {
		return nil, fmt.Errorf("nes: ROM too short for declared PRG size (%d bytes)", prgSize)
	}
	prg := data[offset : offset+prgSize]
	offset += prgSize

	chrIsRAM := chrUnits == 0
	var chr []byte
	if chrIsRAM {
		chr = make([]byte, chrBankSize)
	} else {
		chrSize := chrUnits * chrBankSize
		if offset+chrSize > len(data) {
			return nil, fmt.Errorf("nes: ROM too short for declared CHR size (%d bytes)", chrSize)
		}
		chr = data[offset : offset+chrSize]
	}

	return &Cartridge{
		PRG: prg, CHR: chr, ChrIsRAM: chrIsRAM,
		Mapper: mapper, Mirror: mirror, Battery: battery,
	}, nil
}
