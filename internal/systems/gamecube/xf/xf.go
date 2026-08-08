// Package xf implements the Flipper GPU's Transform (XF) unit: the
// stage between the Command Processor's decoded vertices (the gpu
// package) and the software rasterizer (gpu/render.go) that
// multiplies each vertex by its selected position matrix, then by the
// projection matrix, performs the perspective divide, and maps the
// result into viewport (screen) coordinates.
//
// Real XF hardware exposes two separate address spaces reached via
// LOAD_XF_REG commands: "XF memory" (memory.go) holding uploaded
// matrix data, and "XF registers" (registers.go) controlling
// per-vertex behavior (active matrix index, texgen count, projection,
// viewport). Both are documented publicly via YAGCD (Yet Another
// GameCube Documentation), a community reference independent of any
// emulator's own source.
//
// Like the rest of internal/systems/gamecube, this package is
// deliberately NOT wired into the app as a playable system - see
// internal/systems/powerpc's package doc for why.
package xf
