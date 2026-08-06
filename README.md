<div align="center">

<p align="center">
    <img src=".github/assets/logo_res.png" width="600" />
</p>

*A retro game launcher, with its own multi-emulators.*

<a href="https://github.com/Duc-inc/espaze/actions/workflows/build.yml">
  <img src="https://github.com/Duc-inc/espaze/actions/workflows/build.yml/badge.svg" alt="Build status" />
</a>
<img src="https://img.shields.io/badge/Go-1.25-00ADD8?logo=go&logoColor=white" alt="Go 1.25" />
<img src="https://img.shields.io/badge/Wails-v2-DF0000?logo=go&logoColor=white" alt="Wails v2" />
<img src="https://img.shields.io/badge/platforms-Windows%20%7C%20macOS%20%7C%20Linux-555" alt="Platforms" />

<br>

**Espaze** is a multiplatform (Windows, macOS, Linux) retro game launcher.
<br>
Bundling several emulators into a single launcher.

</div>

<br>

## Supported systems

| System | Status | Extensions |
|---|---|---|
| CHIP-8 | ✅ Complete | `.ch8` |
| Super-CHIP | ✅ Complete | `.sc8` |
| Game Boy (DMG) | ✅ Complete  | `.gb` |
| NES | 🚧 In progress  | `.nes` |


## Tech stack

- **Backend**: [Go](https://go.dev/) + [Wails v2](https://wails.io/) (native window, Go ↔ JS bindings).
- **Frontend**: vanilla JavaScript + [Vite](https://vitejs.dev/), no framework - each view is a small module that mounts its own DOM.
- **Emulation**: homemade implementations under `internal/systems/<console>/`, all behind one common interface (`core.Core`) that lets any new console plug in without touching the rest of the app.


## Architecture


- `internal/emulation/core/core.go` - the `Core` contract every system implements.
- `internal/systems/<console>/` - one implementation per console (CPU, memory, video, sound...).
- `internal/app/app/` - the bindings exposed to the frontend via Wails.
- `frontend/src/views/` - one view per screen (library, detail, settings, player...).
