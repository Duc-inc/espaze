package app

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/Duc-inc/espaze/internal/config/config"
	"github.com/Duc-inc/espaze/internal/config/paths"
	"github.com/Duc-inc/espaze/internal/emulation/engine"
	"github.com/Duc-inc/espaze/internal/library/library"
	"github.com/Duc-inc/espaze/internal/library/storage"
)

// App is the single object bound into the Wails frontend. It owns the
// long-lived pieces (config, library, active engine) and exposes plain
// methods the JS side calls directly - Wails handles the RPC plumbing.
type App struct {
	ctx context.Context

	cfg     *config.Config
	cfgPath string

	lib *library.Library

	engine       *engine.Engine
	activeGameID string
	sessionStart time.Time
}

// New constructs an App with nothing wired up yet; call Startup first.
func New() *App {
	return &App{}
}

// Startup is called by Wails once the frontend is ready. It loads config
// and the persisted library, creating both on first run.
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx

	if err := paths.EnsureDataDirs(); err != nil {
		a.logError("startup: ensure data dirs", err)
		return
	}

	cfgPath, err := paths.ConfigFile()
	if err != nil {
		a.logError("startup: resolve config path", err)
		return
	}
	a.cfgPath = cfgPath

	cfg, err := config.Load(cfgPath)
	if err != nil {
		a.logError("startup: load config", err)
		cfg = config.Default()
	}
	a.cfg = cfg

	libPath, err := paths.LibraryStoreFile()
	if err != nil {
		a.logError("startup: resolve library path", err)
		return
	}
	a.lib = library.New(storage.NewJSONStore(libPath))
	if err := a.lib.Load(); err != nil {
		a.logError("startup: load library", err)
	}
}

// Shutdown is called by Wails as the app closes.
func (a *App) Shutdown(context.Context) {
	if a.engine != nil {
		a.engine.Stop()
	}
}

func (a *App) logError(context string, err error) {
	fmt.Fprintf(os.Stderr, "espaze: %s: %v\n", context, err)
}
