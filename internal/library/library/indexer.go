package library

import (
	"time"

	"github.com/Duc-inc/espaze/internal/library/game"
)

// buildGame turns a freshly scanned file into a brand new library entry.
func buildGame(f ScannedFile) game.Game {
	return game.Game{
		ID:          game.IDFromPath(f.Path),
		Title:       game.TitleFromFilename(f.Path),
		System:      f.System,
		Path:        f.Path,
		ArtworkPath: game.FindAdjacentArtwork(f.Path),
		AddedAt:     time.Now(),
	}
}

// mergeGames folds newly scanned files into the existing library: known
// games (matched by ID, derived from their path) keep all their metadata
// untouched (play time, artwork, ...); files seen for the first time are
// added. Games from folders not part of this scan are left alone.
func mergeGames(existing []game.Game, scanned []ScannedFile) []game.Game {
	byID := make(map[string]game.Game, len(existing))
	for _, g := range existing {
		byID[g.ID] = g
	}

	for _, f := range scanned {
		id := game.IDFromPath(f.Path)
		if _, known := byID[id]; !known {
			byID[id] = buildGame(f)
		}
	}

	merged := make([]game.Game, 0, len(byID))
	for _, g := range byID {
		merged = append(merged, g)
	}
	return merged
}
