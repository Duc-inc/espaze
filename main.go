package main

import (
	"context"
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/Duc-inc/espaze/internal/app/app"
	"github.com/Duc-inc/espaze/internal/config/config"
	"github.com/Duc-inc/espaze/internal/config/paths"

	// Blank-imported so its init() registers the core with the emulation
	// registry. Every future system gets the same one-line hookup here.
	_ "github.com/Duc-inc/espaze/internal/systems/atari2600"
	_ "github.com/Duc-inc/espaze/internal/systems/chip8"
	_ "github.com/Duc-inc/espaze/internal/systems/gameboy"
	_ "github.com/Duc-inc/espaze/internal/systems/gamegear"
	_ "github.com/Duc-inc/espaze/internal/systems/gba"
	_ "github.com/Duc-inc/espaze/internal/systems/gbc"
	_ "github.com/Duc-inc/espaze/internal/systems/genesis"
	_ "github.com/Duc-inc/espaze/internal/systems/nes"
	_ "github.com/Duc-inc/espaze/internal/systems/pcengine"
	_ "github.com/Duc-inc/espaze/internal/systems/schip"
	_ "github.com/Duc-inc/espaze/internal/systems/sms"
)

//go:embed all:frontend/dist
var assets embed.FS

const defaultWidth, defaultHeight = 1280, 800

func main() {
	a := app.New()
	width, height := loadWindowSize()

	err := wails.Run(&options.App{
		Title:            "Espaze",
		Width:            width,
		Height:           height,
		MinWidth:         960,
		MinHeight:        600,
		Frameless:        true,
		AssetServer:      &assetserver.Options{Assets: assets},
		BackgroundColour: &options.RGBA{R: 12, G: 14, B: 20, A: 1},
		OnStartup:        a.Startup,
		OnShutdown:       a.Shutdown,
		OnBeforeClose:    saveWindowSize,
		Bind: []interface{}{
			a,
		},
	})
	if err != nil {
		println("espaze: fatal:", err.Error())
	}
}

// loadWindowSize reads the last saved window size, before the window is
// even created - config.Load happens independently of app.Startup here
// because the initial size has to be known before wails.Run is called.
func loadWindowSize() (int, int) {
	cfgPath, err := paths.ConfigFile()
	if err != nil {
		return defaultWidth, defaultHeight
	}
	cfg, err := config.Load(cfgPath)
	if err != nil || cfg.WindowWidth == 0 || cfg.WindowHeight == 0 {
		return defaultWidth, defaultHeight
	}
	return cfg.WindowWidth, cfg.WindowHeight
}

// saveWindowSize persists the current window size right before the app
// closes, so the next launch can restore it.
func saveWindowSize(ctx context.Context) bool {
	width, height := wailsruntime.WindowGetSize(ctx)

	cfgPath, err := paths.ConfigFile()
	if err != nil {
		return false
	}
	cfg, err := config.Load(cfgPath)
	if err != nil {
		cfg = config.Default()
	}
	cfg.WindowWidth = width
	cfg.WindowHeight = height
	_ = cfg.Save(cfgPath)

	return false // never prevent closing
}
