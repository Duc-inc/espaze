package memory

// wram implements the CGB's banked work RAM: a fixed 4KB bank 0 plus 7
// switchable 4KB banks (1-7) selected through SVBK ($FF70) - versus
// DMG's flat, unbanked 8KB.
type wram struct {
	banks [8][0x1000]byte
	bank  byte
}

func newWRAM() *wram { return &wram{bank: 1} }

// selectedBank treats a raw value of 0 as bank 1, exactly like real
// hardware (there's no way to select the fixed bank as the switchable
// one too).
func (w *wram) selectedBank() byte {
	if w.bank == 0 {
		return 1
	}
	return w.bank
}

func (w *wram) writeSVBK(v byte) { w.bank = v & 0x07 }
func (w *wram) readSVBK() byte   { return w.bank | 0xF8 }

// readLow/writeLow implement $C000-$CFFF (and its $E000-$ECFF echo).
func (w *wram) readLow(offset uint16) byte     { return w.banks[0][offset] }
func (w *wram) writeLow(offset uint16, v byte) { w.banks[0][offset] = v }

// readHigh/writeHigh implement $D000-$DFFF (and its $ED00-$FDFF echo).
func (w *wram) readHigh(offset uint16) byte     { return w.banks[w.selectedBank()][offset] }
func (w *wram) writeHigh(offset uint16, v byte) { w.banks[w.selectedBank()][offset] = v }
