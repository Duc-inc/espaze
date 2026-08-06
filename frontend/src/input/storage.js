import {defaultKeymapForSystem} from './buttons.js';

const STORAGE_PREFIX = 'espaze:keymap:';

/** Loads the active keymap for a system: the user's custom one if they
 * ever remapped it in Settings, otherwise the factory default. */
export function loadKeymap(systemId) {
    try {
        const raw = localStorage.getItem(STORAGE_PREFIX + systemId);
        if (raw) {
            return JSON.parse(raw);
        }
    } catch {
        // corrupted storage entry: fall through to the default
    }
    return {...defaultKeymapForSystem(systemId)};
}

export function saveKeymap(systemId, keymap) {
    localStorage.setItem(STORAGE_PREFIX + systemId, JSON.stringify(keymap));
}

export function resetKeymap(systemId) {
    localStorage.removeItem(STORAGE_PREFIX + systemId);
}
