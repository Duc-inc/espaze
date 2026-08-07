// Package riot implements the Atari 2600's other custom chip: the
// 6532 RIOT (RAM-I/O-Timer) - 128 bytes of the system's only RAM, two
// 8-bit I/O ports (joysticks, console switches), and a countdown
// timer games poll for simple time-keeping.
package riot

import "github.com/Duc-inc/espaze/internal/emulation/input"

// Joystick bit positions within the generic input.State bitmask.
const (
	Up = iota
	Down
	Left
	Right
	Fire
)

// RIOT holds the chip's RAM and timer/port state.
type RIOT struct {
	ram [128]byte

	timerValue   byte
	timerShift   uint
	timerCounter int
	timerStarted bool

	state input.State
}

// New returns a RIOT with zeroed RAM and an inactive timer.
func New() *RIOT { return &RIOT{} }

// Reset clears RAM and the timer (real hardware doesn't clear RAM on
// reset, but every core here starts a fresh cartridge from a clean
// slate).
func (r *RIOT) Reset() { *r = RIOT{} }

// SetButtons applies the latest generic input state.
func (r *RIOT) SetButtons(state input.State) { r.state = state }

// ReadRAM/WriteRAM implement the 128-byte RAM window.
func (r *RIOT) ReadRAM(addr byte) byte     { return r.ram[addr&0x7F] }
func (r *RIOT) WriteRAM(addr byte, v byte) { r.ram[addr&0x7F] = v }

// Step advances the timer by cpuCycles CPU cycles.
func (r *RIOT) Step(cpuCycles int) {
	if !r.timerStarted {
		return
	}
	for i := 0; i < cpuCycles; i++ {
		r.timerCounter--
		if r.timerCounter <= 0 {
			r.timerValue--
			r.timerCounter = 1 << r.timerShift
		}
	}
}

// WriteTimer implements TIM1T/TIM8T/TIM64T/T1024T ($294/$295/$296/$297):
// sets the countdown value and the interval (1/8/64/1024 cycles per
// decrement, selected by the low 2 bits of the register offset).
func (r *RIOT) WriteTimer(offset byte, v byte) {
	r.timerValue = v
	switch offset & 0x03 {
	case 0:
		r.timerShift = 0
	case 1:
		r.timerShift = 3
	case 2:
		r.timerShift = 6
	default:
		r.timerShift = 10
	}
	r.timerCounter = 1 << r.timerShift
	r.timerStarted = true
}

// ReadINTIM implements $284: the current countdown value.
func (r *RIOT) ReadINTIM() byte { return r.timerValue }

// ReadSWCHA implements $280: the joystick port, active-low, upper
// nibble player 0 and lower nibble player 1 (player 1 isn't wired up,
// so it always reads released).
func (r *RIOT) ReadSWCHA() byte {
	var v byte = 0xFF
	if r.state.Pressed(Up) {
		v &^= 1 << 4
	}
	if r.state.Pressed(Down) {
		v &^= 1 << 5
	}
	if r.state.Pressed(Left) {
		v &^= 1 << 6
	}
	if r.state.Pressed(Right) {
		v &^= 1 << 7
	}
	return v
}

// ReadSWCHB implements $282: console switches (reset, select,
// difficulty). This project always reports the factory-default
// position: both difficulty switches in "Amateur" (B), color mode.
func (r *RIOT) ReadSWCHB() byte { return 0xFF }
