package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"

	"github.com/Duc-inc/espaze/internal/app/app"

	// Blank-imported so its init() registers the core with the emulation
	// registry. Every future system gets the same one-line hookup here.
	_ "github.com/Duc-inc/espaze/internal/systems/chip8"
	_ "github.com/Duc-inc/espaze/internal/systems/schip"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	a := app.New()

	err := wails.Run(&options.App{
		Title:            "Espaze",
		Width:            1280,
		Height:           800,
		MinWidth:         960,
		MinHeight:        600,
		AssetServer:      &assetserver.Options{Assets: assets},
		BackgroundColour: &options.RGBA{R: 12, G: 14, B: 20, A: 1},
		OnStartup:        a.Startup,
		OnShutdown:       a.Shutdown,
		Bind: []interface{}{
			a,
		},
	})
	if err != nil {
		println("espaze: fatal:", err.Error())
	}
}
