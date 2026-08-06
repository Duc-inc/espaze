// Default keyboard layout for the CHIP-8 core: the classic 4x4 hex keypad
// mapped onto the left-hand side of a QWERTY keyboard. Bit index === hex
// key value, matching how internal/systems/chip8 SetInput reads bits 0-15.
//
//   1 2 3 4        1 2 3 C
//   Q W E R   -->  4 5 6 D
//   A S D F        7 8 9 E
//   Z X C V        A 0 B F

export const CHIP8_KEYMAP = {
    Digit1: 0x1, Digit2: 0x2, Digit3: 0x3, Digit4: 0xC,
    KeyQ: 0x4, KeyW: 0x5, KeyE: 0x6, KeyR: 0xD,
    KeyA: 0x7, KeyS: 0x8, KeyD: 0x9, KeyF: 0xE,
    KeyZ: 0xA, KeyX: 0x0, KeyC: 0xB, KeyV: 0xF,
};

/** Tracks which keys are currently held and derives the resulting bitmask. */
export class KeyState {
    constructor(keymap) {
        this.keymap = keymap;
        this.pressed = new Set();
    }

    handleKeyDown(code) {
        if (!(code in this.keymap)) return false;
        this.pressed.add(this.keymap[code]);
        return true;
    }

    handleKeyUp(code) {
        if (!(code in this.keymap)) return false;
        this.pressed.delete(this.keymap[code]);
        return true;
    }

    bitmask() {
        let mask = 0;
        for (const bit of this.pressed) {
            mask |= (1 << bit);
        }
        return mask;
    }

    reset() {
        this.pressed.clear();
    }
}
