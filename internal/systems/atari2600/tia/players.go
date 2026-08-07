package tia

// player is one of the TIA's two 8-pixel sprite objects. Real hardware
// can draw 2-3 evenly-spaced copies of a player per scanline (NUSIZ's
// "multiple copies" modes) - a trick many games use for multi-car
// racers - which isn't implemented here; every NUSIZ setting draws a
// single copy, only its size-doubling/quadrupling bits are honored.
type player struct {
	grp     byte
	color   byte
	pos     int
	reflect bool
	scale   int // pixels per graphics bit: 1, 2, or 4
}

func newPlayer() player { return player{scale: 1} }

func (p *player) writeNUSIZ(v byte) {
	switch v & 0x07 {
	case 5:
		p.scale = 2
	case 7:
		p.scale = 4
	default:
		p.scale = 1
	}
}

func (p *player) reset(clockX int) { p.pos = clockX }

func (p *player) applyMotion(hm byte) {
	motion := int(int8(hm)) >> 4
	p.pos = ((p.pos-motion)%160 + 160) % 160
}

func (p *player) pixelAt(x int) bool {
	width := 8 * p.scale
	rel := ((x - p.pos) + 160) % 160
	if rel >= width {
		return false
	}
	bitIndex := rel / p.scale
	if !p.reflect {
		bitIndex = 7 - bitIndex
	}
	return p.grp&(1<<uint(bitIndex)) != 0
}

// movable is a missile or the ball: a solid block of configurable
// pixel width, with no internal graphics pattern.
type movable struct {
	pos     int
	width   int
	enabled bool
}

func newMovable() movable { return movable{width: 1} }

func (m *movable) writeSize(bits byte) {
	switch bits & 0x03 {
	case 1:
		m.width = 2
	case 2:
		m.width = 4
	case 3:
		m.width = 8
	default:
		m.width = 1
	}
}

func (m *movable) reset(clockX int) { m.pos = clockX }

func (m *movable) applyMotion(hm byte) {
	motion := int(int8(hm)) >> 4
	m.pos = ((m.pos-motion)%160 + 160) % 160
}

func (m *movable) pixelAt(x int) bool {
	if !m.enabled {
		return false
	}
	rel := ((x - m.pos) + 160) % 160
	return rel < m.width
}
