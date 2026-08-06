// Maps a core's system ID (see internal/emulation/core Metadata.ID on the
// Go side) to the CSS variable carrying its accent color, kept in one
// place so every component that shows a system badge stays in sync.
const SYSTEM_COLORS = {
    chip8: 'var(--color-chip8)',
    schip: 'var(--color-schip)',
    gameboy: 'var(--color-gameboy)',
    nes: 'var(--color-nes)',
};

export function systemColor(systemId) {
    return SYSTEM_COLORS[systemId] ?? 'var(--color-text-muted)';
}
