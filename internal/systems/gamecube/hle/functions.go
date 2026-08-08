package hle

import "github.com/Duc-inc/espaze/internal/systems/powerpc"

// ReportLog collects OSReport's reported strings for a caller to
// inspect (e.g. surface in the app's own log) - the same "drain
// what's accumulated" shape this project already uses for
// audio.Mixer.DrainSamples and gpu.CommandProcessor.DrainTriangles.
type ReportLog struct {
	Lines []string
}

// OSReport returns an HLE stand-in for the SDK's OSReport(fmt, ...):
// reads the format string argument (GPR3, the real PowerPC/PowerOpen
// calling convention's first argument register) and appends it
// verbatim to log - no printf-style %d/%s substitution from the
// remaining arguments (GPR4+), a real simplification of the actual
// varargs behavior.
func OSReport(log *ReportLog) Func {
	return func(cpu *powerpc.CPU, mem MemoryAccess) {
		log.Lines = append(log.Lines, readCString(mem, cpu.GPR(3)))
		cpu.SetGPR(3, 0)
	}
}

// DVDRead returns an HLE stand-in for a DVD read call: copies length
// bytes from image starting at discOffset into memory at destAddr -
// GPR3/GPR4/GPR5 respectively - and returns the byte count actually
// read in GPR3, matching real DVDRead's own return value on success.
// Real DVD access calls take a file handle and run asynchronously
// with a completion callback; this project's simplification is a
// direct, synchronous, offset-based read, closer to how disc.Header/
// disc.FST already expose file locations as plain offsets.
func DVDRead(image []byte) Func {
	return func(cpu *powerpc.CPU, mem MemoryAccess) {
		dest, length, offset := cpu.GPR(3), cpu.GPR(4), cpu.GPR(5)
		var n uint32
		for n < length && int(offset+n) < len(image) {
			mem.Write8(dest+n, image[offset+n])
			n++
		}
		cpu.SetGPR(3, n)
	}
}
