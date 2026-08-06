// Thin wrapper around the generated Wails bindings for controlling the
// currently running emulation core.
import {
    LaunchGame,
    ListSaveSlots,
    LoadState,
    LoadStateFromSlot,
    PauseGame,
    ResumeGame,
    SaveState,
    SaveStateToSlot,
    SendInput,
    StopGame,
} from '../../wailsjs/go/app/App';

export async function launchGame(id) {
    return await LaunchGame(id);
}

export async function stopGame() {
    return await StopGame();
}

export async function pauseGame() {
    return await PauseGame();
}

export async function resumeGame() {
    return await ResumeGame();
}

export async function sendInput(buttonsBitmask) {
    return await SendInput(buttonsBitmask);
}

export async function saveState() {
    return await SaveState();
}

export async function loadState(encoded) {
    return await LoadState(encoded);
}

export async function listSaveSlots() {
    return (await ListSaveSlots()) ?? [];
}

export async function saveStateToSlot(slot) {
    return await SaveStateToSlot(slot);
}

export async function loadStateFromSlot(slot) {
    return await LoadStateFromSlot(slot);
}
