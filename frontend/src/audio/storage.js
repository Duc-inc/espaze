const VOLUME_KEY = 'espaze:audio:volume';
const MUTED_KEY = 'espaze:audio:muted';

/** Loads the saved output volume (0-1), defaulting to 80%. */
export function loadVolume() {
    const raw = localStorage.getItem(VOLUME_KEY);
    const value = raw !== null ? Number(raw) : 0.8;
    return Number.isFinite(value) ? Math.min(1, Math.max(0, value)) : 0.8;
}

export function saveVolume(volume) {
    localStorage.setItem(VOLUME_KEY, String(Math.min(1, Math.max(0, volume))));
}

export function loadMuted() {
    return localStorage.getItem(MUTED_KEY) === 'true';
}

export function saveMuted(muted) {
    localStorage.setItem(MUTED_KEY, String(muted));
}
