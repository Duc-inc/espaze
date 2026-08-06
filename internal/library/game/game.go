package game

import "time"

// Game is one entry in the user's library: a ROM file paired with the
// system core that can run it, plus everything the UI shows about it.
type Game struct {
	ID              string     `json:"id"`
	Title           string     `json:"title"`
	System          string     `json:"system"` // core.Metadata.ID, e.g. "chip8"
	Path            string     `json:"path"`
	ArtworkPath     string     `json:"artworkPath,omitempty"`
	AddedAt         time.Time  `json:"addedAt"`
	LastPlayedAt    *time.Time `json:"lastPlayedAt,omitempty"`
	PlayTimeSeconds int64      `json:"playTimeSeconds"`
}

// RecordSession updates play time and last-played bookkeeping after a run.
func (g *Game) RecordSession(duration time.Duration) {
	now := time.Now()
	g.LastPlayedAt = &now
	g.PlayTimeSeconds += int64(duration.Seconds())
}
