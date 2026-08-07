package memory

// Interrupt Flag bits this project raises.
const (
	IFVBlank = 1 << 0
	IFTimer0 = 1 << 3
	IFTimer1 = 1 << 4
	IFTimer2 = 1 << 5
	IFTimer3 = 1 << 6
)

// interrupts is the IE/IF/IME trio: which sources are individually
// enabled, which are currently flagged, and the global master enable.
type interrupts struct {
	ie, iflags uint16
	ime        bool
}

func (i *interrupts) raise(bit uint16) { i.iflags |= bit }

// pending reports whether the CPU's IRQ line should currently be
// asserted.
func (i *interrupts) pending() bool {
	return i.ime && i.ie&i.iflags != 0
}

func (i *interrupts) writeIE(v uint16)       { i.ie = v }
func (i *interrupts) writeIME(v uint16)      { i.ime = v&1 != 0 }
func (i *interrupts) readIF() uint16         { return i.iflags }
func (i *interrupts) acknowledgeIF(v uint16) { i.iflags &^= v } // writing 1 bits clears them
