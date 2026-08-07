// Maps a core's system ID to the display name shown in the UI (section
// headers grouping games by console).
const SYSTEM_NAMES = {
    chip8: 'CHIP-8',
    schip: 'Super-CHIP',
    gameboy: 'Game Boy',
    gbc: 'Game Boy Color',
    nes: 'NES',
    sms: 'Master System',
    genesis: 'Genesis / Mega Drive',
};

export function systemName(systemId) {
    return SYSTEM_NAMES[systemId] ?? systemId;
}
