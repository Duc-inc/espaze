package rom

import (
	"fmt"

	"github.com/Duc-inc/espaze/internal/systems/chip8/memory"
)

// MaxSize is the largest ROM image that fits below the top of RAM once
// loaded at the standard program start address.
const MaxSize = memory.Size - memory.ProgramStart

// Load validates a ROM image and installs it into memory at ProgramStart.
func Load(mem *memory.Memory, data []byte) error {
	if len(data) == 0 {
		return fmt.Errorf("rom: empty image")
	}
	if len(data) > MaxSize {
		return fmt.Errorf("rom: image too large (%d bytes, max %d)", len(data), MaxSize)
	}
	return mem.LoadROM(data)
}
