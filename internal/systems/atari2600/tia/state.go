package tia

// Snapshot captures the TIA's full register and object state (not the
// pending sample buffer, transient like every other audio core here).
type Snapshot struct {
	VSync, VBlank bool

	PF0, PF1, PF2                  byte
	PFReflect, PFScore, PFPriority bool
	ColuPF, BG                     byte
	P0, P1                         playerSnapshot
	M0, M1, BL                     movableSnapshot
	A0, A1                         audioSnapshot
	HMP0, HMP1, HMM0, HMM1, HMBL   byte
	CXM0P, CXM1P, CXP0FB, CXP1FB   byte
	CXM0FB, CXM1FB, CXBLPF, CXPPMM byte
	InputLatches                   [6]byte
	Clock, Line                    int
	WSync, FrameDone               bool
	SampleCycles                   float64
}

type playerSnapshot struct {
	GRP, Color byte
	Pos        int
	Reflect    bool
	Scale      int
}

type movableSnapshot struct {
	Pos     int
	Width   int
	Enabled bool
}

type audioSnapshot struct {
	AUDC, AUDF, Volume byte
	DivCounter         int
	LFSR               uint16
	Square             bool
}

func snapshotPlayer(p *player) playerSnapshot {
	return playerSnapshot{GRP: p.grp, Color: p.color, Pos: p.pos, Reflect: p.reflect, Scale: p.scale}
}

func restorePlayer(p *player, s playerSnapshot) {
	p.grp, p.color, p.pos, p.reflect, p.scale = s.GRP, s.Color, s.Pos, s.Reflect, s.Scale
}

func snapshotMovable(m *movable) movableSnapshot {
	return movableSnapshot{Pos: m.pos, Width: m.width, Enabled: m.enabled}
}

func restoreMovable(m *movable, s movableSnapshot) {
	m.pos, m.width, m.enabled = s.Pos, s.Width, s.Enabled
}

func snapshotAudio(a *audioChannel) audioSnapshot {
	return audioSnapshot{AUDC: a.audc, AUDF: a.audf, Volume: a.volume, DivCounter: a.divCounter, LFSR: a.lfsr, Square: a.square}
}

func restoreAudio(a *audioChannel, s audioSnapshot) {
	a.audc, a.audf, a.volume = s.AUDC, s.AUDF, s.Volume
	a.divCounter, a.lfsr, a.square = s.DivCounter, s.LFSR, s.Square
}

// Snapshot captures the TIA's current state.
func (t *TIA) Snapshot() Snapshot {
	return Snapshot{
		VSync: t.vsync, VBlank: t.vblank,
		PF0: t.pf.pf0, PF1: t.pf.pf1, PF2: t.pf.pf2,
		PFReflect: t.pf.reflect, PFScore: t.pf.score, PFPriority: t.pf.priority,
		ColuPF: t.colupf, BG: t.bg,
		P0: snapshotPlayer(&t.p0), P1: snapshotPlayer(&t.p1),
		M0: snapshotMovable(&t.m0), M1: snapshotMovable(&t.m1), BL: snapshotMovable(&t.bl),
		A0: snapshotAudio(&t.a0), A1: snapshotAudio(&t.a1),
		HMP0: t.hmp0, HMP1: t.hmp1, HMM0: t.hmm0, HMM1: t.hmm1, HMBL: t.hmbl,
		CXM0P: t.cxm0p, CXM1P: t.cxm1p, CXP0FB: t.cxp0fb, CXP1FB: t.cxp1fb,
		CXM0FB: t.cxm0fb, CXM1FB: t.cxm1fb, CXBLPF: t.cxblpf, CXPPMM: t.cxppmm,
		InputLatches: t.inputLatches,
		Clock:        t.clock, Line: t.line,
		WSync: t.wsync, FrameDone: t.frameDone,
		SampleCycles: t.sampleCycles,
	}
}

// Restore reinstates a previously captured Snapshot.
func (t *TIA) Restore(s Snapshot) {
	t.vsync, t.vblank = s.VSync, s.VBlank
	t.pf.pf0, t.pf.pf1, t.pf.pf2 = s.PF0, s.PF1, s.PF2
	t.pf.reflect, t.pf.score, t.pf.priority = s.PFReflect, s.PFScore, s.PFPriority
	t.colupf, t.bg = s.ColuPF, s.BG
	restorePlayer(&t.p0, s.P0)
	restorePlayer(&t.p1, s.P1)
	restoreMovable(&t.m0, s.M0)
	restoreMovable(&t.m1, s.M1)
	restoreMovable(&t.bl, s.BL)
	restoreAudio(&t.a0, s.A0)
	restoreAudio(&t.a1, s.A1)
	t.hmp0, t.hmp1, t.hmm0, t.hmm1, t.hmbl = s.HMP0, s.HMP1, s.HMM0, s.HMM1, s.HMBL
	t.cxm0p, t.cxm1p, t.cxp0fb, t.cxp1fb = s.CXM0P, s.CXM1P, s.CXP0FB, s.CXP1FB
	t.cxm0fb, t.cxm1fb, t.cxblpf, t.cxppmm = s.CXM0FB, s.CXM1FB, s.CXBLPF, s.CXPPMM
	t.inputLatches = s.InputLatches
	t.clock, t.line = s.Clock, s.Line
	t.wsync, t.frameDone = s.WSync, s.FrameDone
	t.sampleCycles = s.SampleCycles
}
