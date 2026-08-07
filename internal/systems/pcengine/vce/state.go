package vce

// Snapshot captures the VCE's palette memory and port state.
type Snapshot struct {
	Palette [512]uint16
	Addr    uint16
}

// Snapshot captures the VCE's current state.
func (v *VCE) Snapshot() Snapshot { return Snapshot{Palette: v.palette, Addr: v.addr} }

// Restore reinstates a previously captured Snapshot.
func (v *VCE) Restore(s Snapshot) {
	v.palette = s.Palette
	v.addr = s.Addr
}
