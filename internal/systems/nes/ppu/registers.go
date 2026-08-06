package ppu

// PPUCTRL ($2000) bits.
const (
	ctrlNametableMask  = 0x03
	ctrlVRAMIncrement  = 1 << 2
	ctrlSpritePattern  = 1 << 3
	ctrlBGPattern      = 1 << 4
	ctrlSpriteSize8x16 = 1 << 5
	ctrlNMIEnable      = 1 << 7
)

// PPUMASK ($2001) bits.
const (
	maskGreyscale        = 1 << 0
	maskShowBGLeft8      = 1 << 1
	maskShowSpritesLeft8 = 1 << 2
	maskShowBG           = 1 << 3
	maskShowSprites      = 1 << 4
)

// PPUSTATUS ($2002) bits.
const (
	statusSpriteOverflow = 1 << 5
	statusSprite0Hit     = 1 << 6
	statusVBlank         = 1 << 7
)

// scrollRegs implements the PPU's "loopy" internal scroll state: two
// 15-bit registers (v = current VRAM address, t = temporary/latched
// address), a fine-X scroll (0-7), and the shared write-toggle that
// $2005/$2006 both use. Bit layout of v/t, from the PPU's own address
// bus format:
//
//	yyy NN YYYYY XXXXX
//	||| || ||||| +++++-- coarse X (tile column, 0-31)
//	||| || +++++-------- coarse Y (tile row, 0-31)
//	||| ++-------------- nametable select (X bit, Y bit)
//	+++----------------- fine Y (pixel row within a tile, 0-7)
type scrollRegs struct {
	v, t  uint16
	fineX byte
	write bool // the "w" latch: false = first write, true = second
}

func (s *scrollRegs) coarseX() byte    { return byte(s.v & 0x1F) }
func (s *scrollRegs) coarseY() byte    { return byte((s.v >> 5) & 0x1F) }
func (s *scrollRegs) nametableX() byte { return byte((s.v >> 10) & 1) }
func (s *scrollRegs) nametableY() byte { return byte((s.v >> 11) & 1) }
func (s *scrollRegs) fineY() byte      { return byte((s.v >> 12) & 0x7) }

// writeCtrlNametable updates only t's nametable-select bits, as PPUCTRL does.
func (s *scrollRegs) writeCtrlNametable(bits byte) {
	s.t = (s.t &^ (0x03 << 10)) | (uint16(bits&0x03) << 10)
}

// writeScroll implements one $2005 write - the first sets coarse/fine X,
// the second sets coarse/fine Y, flipping the write toggle each time.
func (s *scrollRegs) writeScroll(v byte) {
	if !s.write {
		s.t = (s.t &^ 0x1F) | uint16(v>>3)
		s.fineX = v & 0x07
	} else {
		s.t = (s.t &^ (0x1F << 5)) | (uint16(v>>3) << 5)
		s.t = (s.t &^ (0x7 << 12)) | (uint16(v&0x07) << 12)
	}
	s.write = !s.write
}

// writeAddr implements one $2006 write - high byte then low byte, with v
// only updated to match t after the second write (real hardware quirk).
func (s *scrollRegs) writeAddr(v byte) {
	if !s.write {
		s.t = (s.t &^ 0x7F00) | (uint16(v&0x3F) << 8)
	} else {
		s.t = (s.t &^ 0x00FF) | uint16(v)
		s.v = s.t
	}
	s.write = !s.write
}

func (s *scrollRegs) resetWriteToggle() { s.write = false }

// copyHorizontal copies t's coarse-X and nametable-X bits into v - real
// hardware does this every visible/pre-render scanline at dot 257, which
// is what lets a mid-frame $2005/$2006 write change scroll starting on
// the next scanline (the classic status-bar split trick).
func (s *scrollRegs) copyHorizontal() {
	s.v = (s.v &^ 0x041F) | (s.t & 0x041F)
}

// copyVertical copies t's coarse-Y, fine-Y and nametable-Y bits into v -
// real hardware does this repeatedly during the pre-render line.
func (s *scrollRegs) copyVertical() {
	s.v = (s.v &^ 0x7BE0) | (s.t & 0x7BE0)
}

// incrementY advances fine Y, carrying into coarse Y (and the vertical
// nametable bit) exactly like real hardware, including its quirk: coarse
// Y 31 is out-of-bounds (only reachable by directly poking $2006) and
// wraps to 0 without flipping the nametable bit, unlike the normal
// wrap at 30.
func (s *scrollRegs) incrementY() {
	if s.v&0x7000 != 0x7000 {
		s.v += 0x1000
		return
	}
	s.v &^= 0x7000
	y := (s.v >> 5) & 0x1F
	switch y {
	case 29:
		y = 0
		s.v ^= 0x0800
	case 31:
		y = 0
	default:
		y++
	}
	s.v = (s.v &^ 0x03E0) | (y << 5)
}
