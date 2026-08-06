package storage

import "github.com/Duc-inc/espaze/internal/library/game"

// Store persists the game library between runs. JSONStore is the only
// implementation today; a future SQLite-backed store can be swapped in
// without touching any caller.
type Store interface {
	Load() ([]game.Game, error)
	Save(games []game.Game) error
}
