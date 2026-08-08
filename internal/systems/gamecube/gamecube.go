// Package gamecube wires the from-scratch PowerPC CPU
// (internal/systems/powerpc) to this platform's own physical memory
// map (internal/systems/gamecube/memory), the real VI/SI/DI/AI/EXI/PI
// hardware register peripherals, and (via wgp, the real Write Gather
// Pipe mechanism) the gpu package's Command Processor and software
// rasterizer. A game's real GX command stream - LOAD_XF_REG/LOAD_CP_
// REG/LOAD_BP_REG writes and draw commands - reaches gpu.
// CommandProcessor exactly the way it would on real hardware: PowerPC
// store instructions to 0xCC008000, gathered into 32-byte bursts,
// flushed into the pipeline by FlushGP. This is real, working
// groundwork - see the ipl/hle packages for how a game gets loaded and
// run - but it still isn't registered with the emulation registry or
// core.Core: there's no encrypted-disc/authentication handling, no
// real VAT-driven vertex formats, and TEV is a simplified single-op-
// per-stage stand-in (see gpu's own doc comments), so a real
// commercial game is not expected to render correctly end to end yet.
package gamecube

import (
	"github.com/Duc-inc/espaze/internal/systems/gamecube/ai"
	"github.com/Duc-inc/espaze/internal/systems/gamecube/audio"
	"github.com/Duc-inc/espaze/internal/systems/gamecube/di"
	"github.com/Duc-inc/espaze/internal/systems/gamecube/disc"
	"github.com/Duc-inc/espaze/internal/systems/gamecube/exi"
	"github.com/Duc-inc/espaze/internal/systems/gamecube/gpu"
	"github.com/Duc-inc/espaze/internal/systems/gamecube/hle"
	"github.com/Duc-inc/espaze/internal/systems/gamecube/ipl"
	"github.com/Duc-inc/espaze/internal/systems/gamecube/memory"
	"github.com/Duc-inc/espaze/internal/systems/gamecube/pi"
	"github.com/Duc-inc/espaze/internal/systems/gamecube/si"
	"github.com/Duc-inc/espaze/internal/systems/gamecube/vi"
	"github.com/Duc-inc/espaze/internal/systems/gamecube/wgp"
	"github.com/Duc-inc/espaze/internal/systems/powerpc"
)

// defaultFBWidth/Height match real NTSC GameCube output's common
// resolution - this project's own choice of framebuffer size, not a
// hardware register any game configures here.
const (
	defaultFBWidth  = 640
	defaultFBHeight = 480
)

// GameCube wires the CPU to its memory bus and hardware peripherals.
type GameCube struct {
	bus  *memory.Bus
	proc *powerpc.CPU

	VI  *vi.VI
	SI  *si.SI
	DI  *di.DI
	AI  *ai.AI
	EXI *exi.EXI

	// PI is the Processor Interface: real hardware's own interrupt
	// router, ORing every peripheral's cause into the CPU's single
	// external interrupt line (see Step). A game reads/masks interrupt
	// causes through PI's real INTSR/INTMR registers, not through each
	// peripheral directly.
	PI *pi.PI

	// Audio is the sample mixer AI's PSTAT bit gates (Step below) - AI
	// itself only tracks streaming on/off and volume, it has no sample
	// data of its own; a caller feeds real channel data into Audio via
	// SetChannel/SetADPCMChannel same as before this field existed.
	// AI.Volume()'s L/R scaling isn't applied to Audio's output - Mixer
	// only produces one mixed mono sample, not a stereo pair.
	Audio *audio.Mixer

	// WGP is the Write Gather Pipe (see package doc) - PowerPC stores to
	// wgp.Base land here, not directly on CP. gpPending accumulates
	// completed 32-byte bursts every Step, mirroring the real GP FIFO's
	// own memory continuously filling; FlushGP is what actually feeds
	// the accumulated bytes to CP and rasterizes whatever triangles that
	// produces, matching real GX_Flush/GX_DrawDone's role. Step calls it
	// automatically at vblank (this project's own scheduling choice, not
	// a hardware fact - see Step's doc comment); call FlushGP directly
	// for finer control.
	WGP *wgp.WGP
	CP  *gpu.CommandProcessor
	FB  *gpu.Framebuffer

	hle *hle.Table

	gpPending []byte
}

// busMemoryReader adapts *memory.Bus to gpu.MemoryReader, so CP can
// resolve indexed vertex attributes (vertexformat.go's ARRAY_BASE/
// STRIDE) and CALL_DISPLAY_LIST targets against real MEM1 contents.
type busMemoryReader struct{ bus *memory.Bus }

func (r busMemoryReader) ReadBytes(addr uint32, length int) []byte {
	out := make([]byte, length)
	for i := range out {
		out[i] = r.bus.Read8(addr + uint32(i))
	}
	return out
}

// New wires a fresh CPU, memory bus, and every peripheral (VI/SI/DI/
// AI/EXI/PI/WGP/CP/FB) together and resets the CPU/RAM. discImage may
// be nil if there's no disc to serve DI reads from yet.
func New(discImage []byte) *GameCube {
	g := &GameCube{bus: memory.New()}
	g.proc = powerpc.New(g.bus)

	g.VI = vi.New()
	g.SI = si.New()
	g.DI = di.New(discImage, g.bus)
	g.AI = ai.New()
	g.EXI = exi.New()
	g.Audio = audio.New()
	g.PI = pi.New()
	g.WGP = wgp.New()
	g.CP = gpu.New()
	g.FB = gpu.NewFramebuffer(defaultFBWidth, defaultFBHeight)
	g.CP.SetMemoryReader(busMemoryReader{g.bus})

	g.bus.Attach(vi.Base, vi.Size, g.VI)
	g.bus.Attach(si.Base, si.Size, g.SI)
	g.bus.Attach(di.Base, di.Size, g.DI)
	g.bus.Attach(ai.Base, ai.Size, g.AI)
	g.bus.Attach(exi.Base, exi.Size, g.EXI)
	g.bus.Attach(pi.Base, pi.Size, g.PI)
	g.bus.Attach(wgp.Base, wgp.Size, g.WGP)

	return g
}

// Reset clears RAM and every register. Peripheral state (SI channel
// data, AI volume, etc.) is left as-is - callers that want that reset
// too should reconstruct via New.
func (g *GameCube) Reset() {
	g.bus.Reset()
	g.proc.Reset()
}

// Step executes exactly one PowerPC instruction, ticks VI/AI, reports
// their current interrupt state into PI (real hardware's own interrupt
// router - PI.SetCause), and delivers a real external interrupt
// exception (powerpc's exceptions.go) if PI reports any unmasked cause
// still pending. This is level-triggered like real hardware: a game
// that returns from the exception without clearing the source (VI's
// DI0-3, AI's AICR AIINT bit) sees the interrupt fire again on the
// very next Step. Real hardware ticks VI/AI on their own pixel/sample
// clocks, independent of CPU instruction count; tying them to one Step
// call each is this project's own simplification, not a timing claim.
// Any Write Gather Pipe bursts completed this Step are appended to
// gpPending - see FlushGP for when they actually reach CP.
func (g *GameCube) Step() int {
	cycles := g.proc.Step()
	_, vblank := g.VI.Step()
	g.AI.Step()

	g.PI.SetCause(pi.BitVI, g.VI.AnyInterruptActive())
	g.PI.SetCause(pi.BitAI, g.AI.Interrupting())
	g.PI.SetCause(pi.BitDI, g.DI.Interrupting())
	if g.PI.Pending() {
		g.proc.RaiseExternalInterrupt()
	}

	if g.AI.Playing() {
		g.Audio.Step(1)
	}

	for _, burst := range g.WGP.DrainBursts() {
		g.gpPending = append(g.gpPending, burst...)
	}

	// Auto-flushing at vblank is this project's own choice, not a real
	// hardware behavior - real GX code calls GX_Flush/GX_DrawDone on its
	// own schedule, not because VI wrapped a frame. It's a reasonable
	// stand-in given most games do finish a frame's drawing and swap
	// around vblank anyway, and it means a command list left pending at
	// the end of a real Step loop still eventually reaches the screen. A
	// caller that wants exact control can still call FlushGP directly at
	// any other point; doing so here just drains whatever's pending
	// early, it's not harmful to call twice.
	if vblank {
		g.FlushGP()
	}
	return cycles
}

// FlushGP feeds every byte accumulated from the Write Gather Pipe
// since the last flush to CP.Execute, rasterizes every triangle that
// produced into FB, then - if VI is enabled and has a real
// framebuffer address configured (TFBL, VI.FramebufferAddr) - encodes
// FB into the real YUY2 byte layout (gpu.EncodeXFB) and writes it to
// MEM1 at that address, exactly where real VI's video encoder would
// read a real game's XFB from. A game that hasn't set up TFBL yet
// (FramebufferAddr() == 0, the power-on default) isn't written to,
// since address 0 is where boot code lives, not a real framebuffer.
// This is this project's stand-in for a game calling GX_Flush/GX_
// DrawDone (real code always pads a command list to a 32-byte
// multiple before relying on it having reached CP, exactly because of
// the WGP's own gathering behavior - see wgp's doc comment) and for
// the Pixel Engine that would otherwise consume finished primitives as
// CP produces them. Step already calls this automatically at vblank;
// call it directly for finer control (e.g. to see a partial draw
// before a frame ends).
func (g *GameCube) FlushGP() {
	g.CP.Execute(g.gpPending)
	g.gpPending = nil
	for _, tri := range g.CP.DrainTriangles() {
		g.FB.DrawTriangle(tri)
	}

	if g.VI.Enabled() {
		if addr := g.VI.FramebufferAddr(); addr != 0 {
			for i, b := range gpu.EncodeXFB(g.FB.FrameBuffer()) {
				g.bus.Write8(addr+uint32(i), b)
			}
		}
	}
}

// mem1CachedBase/mem1Size mirror ipl's own private constants (see
// apploader.go) - the same cached-view-of-MEM1 BAT mapping convention
// real GameCube DOL/apploader load addresses assume.
const (
	mem1CachedBase = 0x80000000
	mem1Size       = 24 * 1024 * 1024
)

// dolWriter translates a DOL section's effective (cached-view) load
// address to a physical MEM1 offset before writing, so LoadDOL's
// writes land at the same physical bytes the CPU's own BAT-translated
// instruction fetches will read from.
type dolWriter struct{ bus *memory.Bus }

func (w dolWriter) Write8(addr uint32, v byte) { w.bus.Write8(addr-mem1CachedBase, v) }

// LoadDOL parses and loads a real DOL executable (disc.ParseDOL/
// LoadInto - GameCube's standard, publicly-documented native
// executable format, not a disc image) into MEM1 at each section's
// declared load address, configures the same cached-view BAT mapping
// Boot does, and sets PC to the DOL's entry point. This is the same
// shortcut real development tools use to run homebrew directly,
// skipping the disc apploader boot path (Boot) entirely - there's no
// real apploader binary available to this project either way (that's
// Nintendo's copyrighted code, not something this project can ship or
// interpret), so for anything beyond BootViaApploader's own minimal
// header/entry handling, a caller needs an already-built DOL.
func (g *GameCube) LoadDOL(image []byte) {
	dol := disc.ParseDOL(image)
	g.proc.SetBAT(4, mem1CachedBase, mem1Size, 0)
	dol.LoadInto(image, dolWriter{g.bus})
	g.proc.SetPC(dol.Entry)
}

// LoadAt copies data directly into MEM1 at the given physical address
// - a stand-in for the disc-loading process real hardware's IPL (boot
// ROM) performs.
func (g *GameCube) LoadAt(addr uint32, data []byte) {
	for i, v := range data {
		g.bus.Write8(addr+uint32(i), v)
	}
}

// SetPC/SetGPR seed the CPU's initial execution state - this
// project's own stand-in for what real IPL firmware or a boot loader
// (ipl.BootViaApploader) would otherwise arrange before running code.
func (g *GameCube) SetPC(addr uint32)        { g.proc.SetPC(addr) }
func (g *GameCube) SetGPR(reg int, v uint32) { g.proc.SetGPR(reg, v) }

// PC exposes the current program counter, mainly for tests/tools.
func (g *GameCube) PC() uint32 { return g.proc.PC() }

// Boot runs the real apploader-based boot path (ipl.BootViaApploader)
// against this GameCube's own bus/CPU: parses image's disc header,
// loads its apploader code into MEM1, configures the BAT mapping real
// IPL firmware would (see ipl's own doc comments), and sets PC to the
// apploader's entry point - ready for Step to actually start running
// it, exactly like passing discImage to New wires up DI's reads but
// doesn't itself execute anything.
func (g *GameCube) Boot(image []byte) (disc.Header, ipl.ApploaderInfo, error) {
	return ipl.BootViaApploader(image, g.bus, g.proc)
}

// InstallHLE wires up a high-level-emulated function (see the hle
// package's doc comment for what that means and why) at addr: real
// code that reaches addr runs fn's Go implementation instead of
// interpreting real SDK code this project doesn't have. The
// underlying hle.Table is created on first use.
func (g *GameCube) InstallHLE(addr uint32, fn hle.Func) {
	if g.hle == nil {
		g.hle = hle.New(g.bus)
	}
	g.hle.Install(g.proc, addr, fn)
}
