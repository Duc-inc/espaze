// lightmemory.go bridges XF's real light memory block (memory.go's
// LightsStart..LightsEnd) into this project's simplified Illuminate
// model (lighting.go). Public hardware notes document light 0 at
// LightsStart with a 16-word stride per light, holding RGBA color,
// cosine/distance attenuation, and position-or-direction fields; this
// project reads only the two fields Illuminate currently uses
// (position and color), leaving attenuation and the position-vs-
// infinite-direction mode distinction for later, once the color-
// channel control registers that select between them are decoded
// (registers.go's RegColorOutputCtrl0/1, still raw-only).
package xf

const (
	// LightWords is the word stride between consecutive lights in XF
	// light memory.
	LightWords = 16
	// LightColorOffset/LightPositionOffset are word offsets from a
	// given light's base address (LightsStart + index*LightWords) to
	// its packed RGBA color and its position/direction (3 consecutive
	// words: x, y, z) - exported so callers can build real light-memory
	// writes without duplicating these offsets.
	LightColorOffset    = 3
	LightPositionOffset = 10
)

// LightIndexForAddr reports which light (0-MaxLights-1) the given XF
// memory address belongs to, if any - for callers (e.g. the gpu
// package's command decoder) that want to know when a LOAD_XF_REG
// write should trigger re-reading a light via ReadLight.
func LightIndexForAddr(addr uint16) (index int, ok bool) {
	if addr < LightsStart || addr >= LightsEnd {
		return 0, false
	}
	return int((addr - LightsStart) / LightWords), true
}

// ReadLight decodes light index (0-MaxLights-1) from XF memory into
// this project's simplified Light: position (assuming point-light
// mode, not the infinite-direction alternative real hardware also
// supports) and packed RGBA color.
//
// Enabled is unconditionally true and is the least faithful part of
// this bridge: real hardware's per-light enable state doesn't live in
// this memory block at all, it lives in the color-channel control
// registers (registers.go's RegColorOutputCtrl0/1, still raw-only), so
// this function can't actually tell whether a light with data present
// is switched on. Treat every light ReadLight returns as "configured,
// not necessarily lit" - a caller that needs the real answer must
// decode those control registers first; until then, a caller that
// wants a light disabled despite having memory data can still
// overwrite the result via SetLight (gpu package).
func (m *Memory) ReadLight(index int) Light {
	base := uint16(LightsStart + index*LightWords)
	return Light{
		Position: Vec3{
			X: m.ReadFloat32(base + LightPositionOffset),
			Y: m.ReadFloat32(base + LightPositionOffset + 1),
			Z: m.ReadFloat32(base + LightPositionOffset + 2),
		},
		Color:   rgba8ToLightColor(m.Read(base + LightColorOffset)),
		Enabled: true,
	}
}

// rgba8ToLightColor unpacks a byte-per-channel RGBA word, R in the
// most significant byte - this project's own consistent convention
// for packed RGBA words (matching the texture package's FormatRGBA8
// byte order), not independently re-confirmed for this specific
// register.
func rgba8ToLightColor(word uint32) LightColor {
	return LightColor{
		R: float32(byte(word>>24)) / 255,
		G: float32(byte(word>>16)) / 255,
		B: float32(byte(word>>8)) / 255,
	}
}
