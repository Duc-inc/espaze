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

// NES bit layout matches internal/systems/nes/memory/controller.go's
// constants exactly (A=0, B=1, Select=2, Start=3, Up=4, Down=5, Left=6,
// Right=7), and reuses the same X/Z-for-A/B convention as Game Boy above.
const NES_BUTTONS = [
    {bit: 4, label: 'Haut'}, {bit: 5, label: 'Bas'}, {bit: 6, label: 'Gauche'}, {bit: 7, label: 'Droite'},
    {bit: 0, label: 'A'}, {bit: 1, label: 'B'}, {bit: 2, label: 'Select'}, {bit: 3, label: 'Start'},
];

const NES_DEFAULT_KEYMAP = {
    ArrowUp: 4, ArrowDown: 5, ArrowLeft: 6, ArrowRight: 7,
    KeyX: 0, KeyZ: 1, ShiftLeft: 2, Enter: 3,
};

// SMS bit layout matches internal/systems/sms/memory/joypad.go's
// constants exactly (Up=0, Down=1, Left=2, Right=3, Button1=4,
// Button2=5, Pause=6 - Pause is wired straight to NMI on real hardware,
// not either I/O port, but it's still just another button here).
const SMS_BUTTONS = [
    {bit: 0, label: 'Haut'}, {bit: 1, label: 'Bas'}, {bit: 2, label: 'Gauche'}, {bit: 3, label: 'Droite'},
    {bit: 4, label: '1'}, {bit: 5, label: '2'}, {bit: 6, label: 'Pause'},
];

const SMS_DEFAULT_KEYMAP = {
    ArrowUp: 0, ArrowDown: 1, ArrowLeft: 2, ArrowRight: 3,
    KeyX: 4, KeyZ: 5, Enter: 6,
};

// Genesis bit layout matches internal/systems/genesis/memory/controller.go's
// constants exactly (Up=0, Down=1, Left=2, Right=3, B=4, C=5, A=6, Start=7).
const GENESIS_BUTTONS = [
    {bit: 0, label: 'Haut'}, {bit: 1, label: 'Bas'}, {bit: 2, label: 'Gauche'}, {bit: 3, label: 'Droite'},
    {bit: 6, label: 'A'}, {bit: 4, label: 'B'}, {bit: 5, label: 'C'}, {bit: 7, label: 'Start'},
];

const GENESIS_DEFAULT_KEYMAP = {
    ArrowUp: 0, ArrowDown: 1, ArrowLeft: 2, ArrowRight: 3,
    KeyZ: 6, KeyX: 4, KeyC: 5, Enter: 7,
};

// Game Gear bit layout matches internal/systems/gamegear's constants
// exactly (Up/Down/Left/Right/Button1/Button2 reuse sms/memory's bits
// 0-5 directly, Start is the Game Gear's own addition at bit 6).
const GAMEGEAR_BUTTONS = [
    {bit: 0, label: 'Haut'}, {bit: 1, label: 'Bas'}, {bit: 2, label: 'Gauche'}, {bit: 3, label: 'Droite'},
    {bit: 4, label: '1'}, {bit: 5, label: '2'}, {bit: 6, label: 'Start'},
];

const GAMEGEAR_DEFAULT_KEYMAP = {
    ArrowUp: 0, ArrowDown: 1, ArrowLeft: 2, ArrowRight: 3,
    KeyX: 4, KeyZ: 5, Enter: 6,
};

// Atari 2600 bit layout matches internal/systems/atari2600/riot's
// constants exactly (Up=0, Down=1, Left=2, Right=3, Fire=4).
const ATARI2600_BUTTONS = [
    {bit: 0, label: 'Haut'}, {bit: 1, label: 'Bas'}, {bit: 2, label: 'Gauche'}, {bit: 3, label: 'Droite'},
    {bit: 4, label: 'Feu'},
];

const ATARI2600_DEFAULT_KEYMAP = {
    ArrowUp: 0, ArrowDown: 1, ArrowLeft: 2, ArrowRight: 3,
    KeyX: 4,
};

export const SYSTEM_BUTTONS = {
    chip8: CHIP8_BUTTONS,
    schip: CHIP8_BUTTONS,
    gameboy: GAMEBOY_BUTTONS,
    gbc: GAMEBOY_BUTTONS, // same 8 buttons, same bit layout as DMG - see gbc/memory joypad reuse
    nes: NES_BUTTONS,
    sms: SMS_BUTTONS,
    genesis: GENESIS_BUTTONS,
    gamegear: GAMEGEAR_BUTTONS,
    atari2600: ATARI2600_BUTTONS,
};

export const DEFAULT_KEYMAPS = {
    chip8: CHIP8_DEFAULT_KEYMAP,
    schip: CHIP8_DEFAULT_KEYMAP,
    gameboy: GAMEBOY_DEFAULT_KEYMAP,
    gbc: GAMEBOY_DEFAULT_KEYMAP,
    nes: NES_DEFAULT_KEYMAP,
    sms: SMS_DEFAULT_KEYMAP,
    genesis: GENESIS_DEFAULT_KEYMAP,
    gamegear: GAMEGEAR_DEFAULT_KEYMAP,
    atari2600: ATARI2600_DEFAULT_KEYMAP,
};

export function buttonsForSystem(systemId) {
    return SYSTEM_BUTTONS[systemId] ?? CHIP8_BUTTONS;
}

export function defaultKeymapForSystem(systemId) {
    return DEFAULT_KEYMAPS[systemId] ?? CHIP8_DEFAULT_KEYMAP;
}

// Default gamepad binding per system: standard-mapping button index -> bit.
// Game Boy's 8 buttons map onto a controller's d-pad + face buttons +
// select/start almost exactly. CHIP-8/Super-CHIP have far more keys than
// any controller has buttons, so only the common "arcade" convention
// (2/8/4/6 for up/down/left/right, 5 as the main action key) gets a
// default; the rest stay keyboard-only.
const CHIP8_GAMEPAD_MAP = {
    12: 0x2, 13: 0x8, 14: 0x4, 15: 0x6, // d-pad -> up/down/left/right
    0: 0x5, 1: 0x0,                     // A -> 5 (action), B -> 0
};

const GAMEBOY_GAMEPAD_MAP = {
    12: 2, 13: 3, 14: 1, 15: 0, // d-pad -> Up/Down/Left/Right
    0: 4, 1: 5,                 // A -> A, B -> B
    8: 6, 9: 7,                 // Back -> Select, Start -> Start
};

const NES_GAMEPAD_MAP = {
    12: 4, 13: 5, 14: 6, 15: 7, // d-pad -> Up/Down/Left/Right
    0: 0, 1: 1,                 // A -> A, B -> B
    8: 2, 9: 3,                 // Back -> Select, Start -> Start
};

const SMS_GAMEPAD_MAP = {
    12: 0, 13: 1, 14: 2, 15: 3, // d-pad -> Up/Down/Left/Right
    0: 4, 1: 5,                 // A -> 1, B -> 2
    9: 6,                       // Start -> Pause
};

const GENESIS_GAMEPAD_MAP = {
    12: 0, 13: 1, 14: 2, 15: 3, // d-pad -> Up/Down/Left/Right
    0: 6, 1: 4, 2: 5,           // A -> A, B -> B, X -> C
    9: 7,                       // Start -> Start
};

const GAMEGEAR_GAMEPAD_MAP = {
    12: 0, 13: 1, 14: 2, 15: 3, // d-pad -> Up/Down/Left/Right
    0: 4, 1: 5,                 // A -> 1, B -> 2
    9: 6,                       // Start -> Start
};

const ATARI2600_GAMEPAD_MAP = {
    12: 0, 13: 1, 14: 2, 15: 3, // d-pad -> Up/Down/Left/Right
    0: 4,                       // A -> Fire
};

const DEFAULT_GAMEPAD_MAPS = {
    chip8: CHIP8_GAMEPAD_MAP,
    schip: CHIP8_GAMEPAD_MAP,
    gameboy: GAMEBOY_GAMEPAD_MAP,
    gbc: GAMEBOY_GAMEPAD_MAP,
    nes: NES_GAMEPAD_MAP,
    sms: SMS_GAMEPAD_MAP,
    genesis: GENESIS_GAMEPAD_MAP,
    gamegear: GAMEGEAR_GAMEPAD_MAP,
    atari2600: ATARI2600_GAMEPAD_MAP,
};

export function gamepadMapForSystem(systemId) {
    return DEFAULT_GAMEPAD_MAPS[systemId] ?? CHIP8_GAMEPAD_MAP;
}
