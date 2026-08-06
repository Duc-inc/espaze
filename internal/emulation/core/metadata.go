package core

// Metadata describes a system core without needing to instantiate it:
// what it's called, which files it opens, and how fast it runs. The
// library scanner and the frontend both read this instead of poking
// at a live core instance.
type Metadata struct {
	ID              string   // stable identifier, e.g. "chip8"
	Name            string   // human readable, e.g. "CHIP-8"
	Extensions      []string // recognized ROM file extensions, lowercase with dot
	FramesPerSecond float64  // target emulation frame rate
	ScreenWidth     int      // native output width in pixels
	ScreenHeight    int      // native output height in pixels
}
