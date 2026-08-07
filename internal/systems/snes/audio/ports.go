// Package audio wires the SNES's audio coprocessor: the from-scratch
// SPC700 (see the spc700 package) with its own 64KB ARAM, a
// simplified DSP (see the dsp package), and the 4 shared
// communication ports the main CPU and SPC700 use to hand data back
// and forth - the only channel between the two entirely separate
// memory spaces on real hardware. This project simplifies each port
// to a single shared byte both sides can read and write, rather than
// reproducing the real chip's more nuanced per-direction latching.
package audio

// Ports is the 4-byte mailbox shared between the main CPU and the
// SPC700.
type Ports struct {
	data [4]byte
}

func (p *Ports) Read(i int) byte     { return p.data[i&0x03] }
func (p *Ports) Write(i int, v byte) { p.data[i&0x03] = v }

// PortsSnapshot captures the shared mailbox's state.
type PortsSnapshot struct {
	Data [4]byte
}

// Snapshot captures the ports' current state.
func (p *Ports) Snapshot() PortsSnapshot { return PortsSnapshot{Data: p.data} }

// Restore reinstates a previously captured PortsSnapshot.
func (p *Ports) Restore(s PortsSnapshot) { p.data = s.Data }
