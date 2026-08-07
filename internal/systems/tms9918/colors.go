package tms9918

// palette is the TMS9918's fixed 16-color set (unlike every other
// video chip in this project, it has no programmable color RAM at
// all - the colors themselves are hardwired). The specific RGB values
// below are this project's best-effort reproduction of the
// commonly-published TMS9918 palette, not independently color-metered
// against real hardware.
var palette = [16][3]byte{
	{0, 0, 0},       // 0: transparent (treated as black when nothing behind it)
	{0, 0, 0},       // 1: black
	{33, 200, 66},   // 2: medium green
	{94, 220, 120},  // 3: light green
	{84, 85, 237},   // 4: dark blue
	{125, 118, 252}, // 5: light blue
	{212, 82, 77},   // 6: dark red
	{66, 235, 245},  // 7: cyan
	{252, 85, 84},   // 8: medium red
	{255, 121, 120}, // 9: light red
	{212, 193, 84},  // 10: dark yellow
	{230, 206, 128}, // 11: light yellow
	{33, 176, 59},   // 12: dark green
	{201, 91, 186},  // 13: magenta
	{204, 204, 204}, // 14: gray
	{255, 255, 255}, // 15: white
}

func colorRGB(index byte) (r, g, b byte) {
	c := palette[index&0x0F]
	return c[0], c[1], c[2]
}
