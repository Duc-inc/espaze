package game

import (
	"os"
	"path/filepath"
	"strings"
)

// artworkExtensions are checked in order; the first match wins.
var artworkExtensions = []string{".png", ".jpg", ".jpeg"}

// FindAdjacentArtwork looks for a cover image sitting right next to a ROM
// file - same name, an image extension (e.g. "Tetris.gb" + "Tetris.png").
// Many hand-curated ROM collections ship exactly this layout. Returns ""
// if nothing matches, which just means no cover art, not an error.
func FindAdjacentArtwork(romPath string) string {
	base := strings.TrimSuffix(romPath, filepath.Ext(romPath))
	for _, ext := range artworkExtensions {
		candidate := base + ext
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate
		}
	}
	return ""
}
