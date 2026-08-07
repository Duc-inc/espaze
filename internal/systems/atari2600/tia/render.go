package tia

// renderPixel resolves one visible pixel's final color by real
// hardware's fixed priority order, and updates the collision latches
// for every pair of objects that overlap here - matching the actual
// TIA behavior of computing collisions from the analog video circuit
// every clock, not just when CXCLR happens to be read.
func (t *TIA) renderPixel(x, y int) {
	if t.vblank {
		t.frame.SetPixel(x, y, 0, 0, 0, 0xFF)
		return
	}

	pf := t.pf.pixelAt(x)
	ball := t.bl.pixelAt(x)
	p0 := t.p0.pixelAt(x)
	p1 := t.p1.pixelAt(x)
	m0 := t.m0.pixelAt(x)
	m1 := t.m1.pixelAt(x)
	t.updateCollisions(p0, p1, m0, m1, ball, pf)

	color := t.resolveColor(x, p0, p1, m0, m1, ball, pf)
	r, g, b := rgb(color)
	t.frame.SetPixel(x, y, r, g, b, 0xFF)
}

func (t *TIA) resolveColor(x int, p0, p1, m0, m1, ball, pf bool) byte {
	playfieldOrBall := pf || ball
	if t.pf.priority && playfieldOrBall {
		return t.playfieldColor(x)
	}
	switch {
	case p0 || m0:
		return t.p0.color
	case p1 || m1:
		return t.p1.color
	case playfieldOrBall:
		return t.playfieldColor(x)
	default:
		return t.bg
	}
}

func (t *TIA) playfieldColor(x int) byte {
	if t.pf.score {
		if x < 80 {
			return t.p0.color
		}
		return t.p1.color
	}
	return t.colupf
}

func (t *TIA) updateCollisions(p0, p1, m0, m1, ball, pf bool) {
	if m0 && p1 {
		t.cxm0p |= 0x80
	}
	if m0 && p0 {
		t.cxm0p |= 0x40
	}
	if m1 && p0 {
		t.cxm1p |= 0x80
	}
	if m1 && p1 {
		t.cxm1p |= 0x40
	}
	if p0 && pf {
		t.cxp0fb |= 0x80
	}
	if p0 && ball {
		t.cxp0fb |= 0x40
	}
	if p1 && pf {
		t.cxp1fb |= 0x80
	}
	if p1 && ball {
		t.cxp1fb |= 0x40
	}
	if m0 && pf {
		t.cxm0fb |= 0x80
	}
	if m0 && ball {
		t.cxm0fb |= 0x40
	}
	if m1 && pf {
		t.cxm1fb |= 0x80
	}
	if m1 && ball {
		t.cxm1fb |= 0x40
	}
	if ball && pf {
		t.cxblpf |= 0x80
	}
	if p0 && p1 {
		t.cxppmm |= 0x80
	}
	if m0 && m1 {
		t.cxppmm |= 0x40
	}
}
