/** Tracks which keys are currently held and derives the resulting bitmask
 * for whatever keymap (code -> bit) it's given. */
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
