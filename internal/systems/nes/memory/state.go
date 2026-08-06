package memory

// Snapshot captures the bus's own state (work RAM, cartridge save RAM,
// both controllers' shift registers) plus the mapper's, since the
// mapper lives behind an interface the top-level core doesn't reach
// into directly.
type Snapshot struct {
	RAM    [0x0800]byte
	PRGRAM [0x2000]byte

	Controller1, Controller2 ControllerSnapshot

	Mapper []byte
}

// ControllerSnapshot captures one gamepad's shift-register state -
// held-button state itself isn't included, since that's re-supplied by
// the frontend every frame rather than being part of "the game's state".
type ControllerSnapshot struct {
	Buttons, Shift byte
	Strobe         bool
}

func (c *Controller) snapshot() ControllerSnapshot {
	return ControllerSnapshot{Buttons: c.buttons, Shift: c.shift, Strobe: c.strobe}
}

func (c *Controller) restore(s ControllerSnapshot) {
	c.buttons, c.shift, c.strobe = s.Buttons, s.Shift, s.Strobe
}

func (b *Bus) Snapshot() Snapshot {
	return Snapshot{
		RAM: b.ram, PRGRAM: b.prgRAM,
		Controller1: b.controller1.snapshot(), Controller2: b.controller2.snapshot(),
		Mapper: b.mapper.Snapshot(),
	}
}

func (b *Bus) Restore(s Snapshot) error {
	b.ram, b.prgRAM = s.RAM, s.PRGRAM
	b.controller1.restore(s.Controller1)
	b.controller2.restore(s.Controller2)
	return b.mapper.Restore(s.Mapper)
}
