// Maps a core's system ID (see internal/emulation/core Metadata.ID on the
// Go side) to the CSS variable carrying its accent color, kept in one
// place so every component that shows a system badge stays in sync.
const SYSTEM_COLORS = {
    chip8: 'var(--color-chip8)',
    schip: 'var(--color-schip)',
    gameboy: 'var(--color-gameboy)',
    gbc: 'var(--color-gbc)',
    nes: 'var(--color-nes)',
    sms: 'var(--color-sms)',
    genesis: 'var(--color-genesis)',
    gamegear: 'var(--color-gamegear)',
    atari2600: 'var(--color-atari2600)',
    pcengine: 'var(--color-pcengine)',
};

export function systemColor(systemId) {
    return SYSTEM_COLORS[systemId] ?? 'var(--color-text-muted)';
}
