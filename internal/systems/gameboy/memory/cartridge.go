package memory

import "fmt"

// Header offsets within the cartridge ROM (see Pan Docs "The Cartridge
// Header"). Only the fields needed to pick an MBC and size RAM are kept.
const (
	headerTitleStart = 0x0134
	headerTitleEnd   = 0x0144
	headerType       = 0x0147
	headerROMSize    = 0x0148
	headerRAMSize    = 0x0149
)

// Cartridge is the raw ROM image plus the header fields that determine
// which memory bank controller to use.
type Cartridge struct {
	Title   string
	Type    byte
	ROM     []byte
	RAMSize int
}

// ParseCartridge validates a ROM image is at least header-sized and reads
// out the fields needed to construct the right MBC.
func ParseCartridge(data []byte) (*Cartridge, error) {
	if len(data) < 0x150 {
		return nil, fmt.Errorf("memory: rom too small to contain a header (%d bytes)", len(data))
	}

	title := make([]byte, 0, headerTitleEnd-headerTitleStart)
	for _, b := range data[headerTitleStart:headerTitleEnd] {
		if b == 0 {
			break
		}
		title = append(title, b)
	}

	return &Cartridge{
		Title:   string(title),
		Type:    data[headerType],
		ROM:     data,
		RAMSize: ramSizeFromCode(data[headerRAMSize]),
	}, nil
}

func ramSizeFromCode(code byte) int {
	switch code {
	case 0x01:
		return 2 * 1024
	case 0x02:
		return 8 * 1024
	case 0x03:
		return 32 * 1024
	case 0x04:
		return 128 * 1024
	case 0x05:
		return 64 * 1024
	default:
		return 0
	}
}
