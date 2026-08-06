// Canonical per-system button list (bit position + human label) and the
// factory-default keyboard binding for each. Settings and the in-game
// legend both read from SYSTEM_BUTTONS so they can never drift apart.
const CHIP8_BUTTONS = [
    {bit: 0x1, label: '1'}, {bit: 0x2, label: '2'}, {bit: 0x3, label: '3'}, {bit: 0xC, label: 'C'},
    {bit: 0x4, label: '4'}, {bit: 0x5, label: '5'}, {bit: 0x6, label: '6'}, {bit: 0xD, label: 'D'},
    {bit: 0x7, label: '7'}, {bit: 0x8, label: '8'}, {bit: 0x9, label: '9'}, {bit: 0xE, label: 'E'},
    {bit: 0xA, label: 'A'}, {bit: 0x0, label: '0'}, {bit: 0xB, label: 'B'}, {bit: 0xF, label: 'F'},
];

const CHIP8_DEFAULT_KEYMAP = {
    Digit1: 0x1, Digit2: 0x2, Digit3: 0x3, Digit4: 0xC,
    KeyQ: 0x4, KeyW: 0x5, KeyE: 0x6, KeyR: 0xD,
    KeyA: 0x7, KeyS: 0x8, KeyD: 0x9, KeyF: 0xE,
    KeyZ: 0xA, KeyX: 0x0, KeyC: 0xB, KeyV: 0xF,
};

const GAMEBOY_BUTTONS = [
    {bit: 2, label: 'Haut'}, {bit: 3, label: 'Bas'}, {bit: 1, label: 'Gauche'}, {bit: 0, label: 'Droite'},
    {bit: 4, label: 'A'}, {bit: 5, label: 'B'}, {bit: 6, label: 'Select'}, {bit: 7, label: 'Start'},
];

const GAMEBOY_DEFAULT_KEYMAP = {
    ArrowUp: 2, ArrowDown: 3, ArrowLeft: 1, ArrowRight: 0,
    KeyX: 4, KeyZ: 5, ShiftLeft: 6, Enter: 7,
};

export const SYSTEM_BUTTONS = {
    chip8: CHIP8_BUTTONS,
    schip: CHIP8_BUTTONS,
    gameboy: GAMEBOY_BUTTONS,
};

export const DEFAULT_KEYMAPS = {
    chip8: CHIP8_DEFAULT_KEYMAP,
    schip: CHIP8_DEFAULT_KEYMAP,
    gameboy: GAMEBOY_DEFAULT_KEYMAP,
};

export function buttonsForSystem(systemId) {
    return SYSTEM_BUTTONS[systemId] ?? CHIP8_BUTTONS;
}

export function defaultKeymapForSystem(systemId) {
    return DEFAULT_KEYMAPS[systemId] ?? CHIP8_DEFAULT_KEYMAP;
}
