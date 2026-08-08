// apploader.go loads a disc's Apploader: the small program every
// retail GameCube disc ships as data on the disc image itself (NOT
// Nintendo's IPL boot-ROM firmware, which this project doesn't have
// and can't include) that real hardware's IPL runs to actually decide
// what to load and where. Parsing/loading it is exactly as legitimate
// as this project's existing DOL parsing (disc.go) - both are just
// data read from a disc image the user already owns.
//
// Confidence note: the apploader header offset (0x2440) and the
// general shape (a date string, then entry/size/trailer-size fields)
// are commonly cited in public GameCube homebrew documentation. The
// exact byte layout within the header (how much padding follows the
// date string) and the fixed load address this project uses
// (apploaderLoadAddr) are this project's best-effort reconstruction,
// not independently re-verified against a byte-exact public spec the
// way disc.go's header/DOL/FST formats were - treat them as
// provisional.
package ipl

import (
	"fmt"

	"github.com/Duc-inc/espaze/internal/systems/gamecube/disc"
	"github.com/Duc-inc/espaze/internal/systems/powerpc"
)

const (
	apploaderHeaderOffset = 0x2440
	apploaderCodeOffset   = apploaderHeaderOffset + 0x20

	// apploaderLoadAddr is an effective (cached-view) address, real
	// hardware's own convention (see BootViaApploader) - not a raw
	// offset into this project's physical MEM1 backing array.
	apploaderLoadAddr = 0x81200000

	// mem1CachedBase/mem1Size describe the BAT mapping this project
	// configures so both the CPU's instruction fetch and this
	// function's own memory writes agree on where apploaderLoadAddr
	// physically lives: the same cached-view-of-MEM1 mapping real
	// GameCube software relies on (see internal/systems/powerpc/mmu.go).
	mem1CachedBase = 0x80000000
	mem1Size       = 24 * 1024 * 1024
)

// ApploaderInfo is what a disc's Apploader header tells IPL about the
// program that follows it.
type ApploaderInfo struct {
	Entry       uint32
	Size        uint32
	TrailerSize uint32
}

func parseApploaderHeader(image []byte) (ApploaderInfo, error) {
	if len(image) < apploaderCodeOffset {
		return ApploaderInfo{}, fmt.Errorf("ipl: image too small to contain an apploader header")
	}
	h := image[apploaderHeaderOffset:]
	return ApploaderInfo{
		Entry:       be32(h[0x10:]),
		Size:        be32(h[0x14:]),
		TrailerSize: be32(h[0x18:]),
	}, nil
}

// BootViaApploader loads a disc's Apploader into memory and points
// the CPU at its entry point - closer to real hardware's own boot
// sequence than Boot's direct DOL jump. What it doesn't do: the real
// apploader's own code expects to call back into a small set of IPL-
// provided OS services (report a status string, read more bytes from
// the disc) to actually do its job of locating and loading the game;
// this project doesn't service those calls yet (that's the OS/SDK HLE
// groundwork, a separate piece of work), so stepping the CPU much past
// the apploader's very first instructions will call into addresses
// this project has no code for. This function's contribution is
// getting a real entry point loaded and ready - not a working
// apploader run.
func BootViaApploader(image []byte, mem disc.Writer, cpu *powerpc.CPU) (disc.Header, ApploaderInfo, error) {
	header, err := disc.ParseHeader(image)
	if err != nil {
		return disc.Header{}, ApploaderInfo{}, err
	}
	info, err := parseApploaderHeader(image)
	if err != nil {
		return disc.Header{}, ApploaderInfo{}, err
	}
	if apploaderCodeOffset+int(info.Size) > len(image) {
		return disc.Header{}, ApploaderInfo{}, fmt.Errorf("ipl: apploader code (size %#x) runs past the end of the image", info.Size)
	}

	// Map the cached view of all of MEM1 (0x80000000+) onto physical
	// address 0, so the CPU's own instruction fetches - which go
	// through this same BAT translation, internal/systems/powerpc/
	// mmu.go - resolve apploaderLoadAddr to the same physical bytes
	// this loop writes.
	cpu.SetBAT(4, mem1CachedBase, mem1Size, 0)

	physicalLoadAddr := uint32(apploaderLoadAddr - mem1CachedBase)
	for i := uint32(0); i < info.Size; i++ {
		mem.Write8(physicalLoadAddr+i, image[apploaderCodeOffset+int(i)])
	}

	cpu.SetGPR(1, stackTop)
	cpu.SetPC(info.Entry)
	return header, info, nil
}

func be32(b []byte) uint32 {
	return uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
}
