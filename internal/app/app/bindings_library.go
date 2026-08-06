package app

import (
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/Duc-inc/espaze/internal/emulation/core"
	"github.com/Duc-inc/espaze/internal/library/game"
)

// ListGames returns every game currently in the library, for the frontend's
// Steam-like grid view.
func (a *App) ListGames() []game.Game {
	return a.lib.List()
}

// LibraryFolders returns the folders currently configured to be scanned.
func (a *App) LibraryFolders() []string {
	return a.cfg.LibraryFolders
}

// AvailableSystems returns metadata for every registered core, so the
// frontend can show what systems are supported (and their file extensions).
func (a *App) AvailableSystems() []core.Metadata {
	return core.List()
}

// BrowseForLibraryFolder opens a native folder picker and returns the
// chosen path, or "" if the user cancelled.
func (a *App) BrowseForLibraryFolder() (string, error) {
	return wailsruntime.OpenDirectoryDialog(a.ctx, wailsruntime.OpenDialogOptions{
		Title: "Choisir un dossier de jeux",
	})
}

// AddLibraryFolder registers a folder for scanning, persists the setting,
// scans it immediately and returns how many new games were found.
func (a *App) AddLibraryFolder(path string) (int, error) {
	a.cfg.AddLibraryFolder(path)
	if err := a.cfg.Save(a.cfgPath); err != nil {
		return 0, err
	}
	return a.lib.ScanFolder(path)
}

// RescanLibrary re-walks every configured folder and returns how many new
// games were found in total.
func (a *App) RescanLibrary() (int, error) {
	total := 0
	for _, folder := range a.cfg.LibraryFolders {
		added, err := a.lib.ScanFolder(folder)
		if err != nil {
			return total, err
		}
		total += added
	}
	return total, nil
}

// RemoveGame removes a game from the library (the ROM file is untouched).
func (a *App) RemoveGame(id string) error {
	return a.lib.Remove(id)
}
