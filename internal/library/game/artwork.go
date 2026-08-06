package game

import "path/filepath"

// ArtworkPath returns where cached cover art for a game would live,
// regardless of whether it has actually been downloaded yet.
func ArtworkPath(artworkDir, gameID string) string {
	return filepath.Join(artworkDir, gameID+".png")
}
