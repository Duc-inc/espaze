package ppu

// Snapshot captures every piece of PPU state, flattened into exported
// fields by hand (rather than embedding the unexported internal types
// directly) so it round-trips correctly through gob, which silently
// drops unexported struct fields instead of erroring on them.
type Snapshot struct {
	Nametables [4][1024]byte
	Palette    [32]byte
	OAM        [256]byte

	ScrollV, ScrollT uint16
	FineX            byte
	WriteToggle      bool

	Ctrl, Mask, Status byte
	OAMAddr            byte
	DataBuffer         byte

	Dot, Scanline int
	OddFrame      bool

	FramePixels []byte
}

func (p *PPU) Snapshot() Snapshot {
	pixels := make([]byte, len(p.frame.Pixels))
	copy(pixels, p.frame.Pixels)

	return Snapshot{
		Nametables:  p.nametables,
		Palette:     p.palette.data,
		OAM:         p.oamMem.data,
		ScrollV:     p.scroll.v,
		ScrollT:     p.scroll.t,
		FineX:       p.scroll.fineX,
		WriteToggle: p.scroll.write,
		Ctrl:        p.ctrl,
		Mask:        p.mask,
		Status:      p.status,
		OAMAddr:     p.oamAddr,
		DataBuffer:  p.dataBuffer,
		Dot:         p.dot,
		Scanline:    p.scanline,
		OddFrame:    p.oddFrame,
		FramePixels: pixels,
	}
}

func (p *PPU) Restore(s Snapshot) {
	p.nametables = s.Nametables
	p.palette.data = s.Palette
	p.oamMem.data = s.OAM
	p.scroll.v, p.scroll.t, p.scroll.fineX, p.scroll.write = s.ScrollV, s.ScrollT, s.FineX, s.WriteToggle
	p.ctrl, p.mask, p.status = s.Ctrl, s.Mask, s.Status
	p.oamAddr, p.dataBuffer = s.OAMAddr, s.DataBuffer
	p.dot, p.scanline, p.oddFrame = s.Dot, s.Scanline, s.OddFrame
	if len(s.FramePixels) == len(p.frame.Pixels) {
		copy(p.frame.Pixels, s.FramePixels)
	}
}
