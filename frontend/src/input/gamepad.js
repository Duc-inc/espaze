/**
 * Reads whichever gamepads are connected and turns their held buttons
 * into the same generic bitmask keyboard input produces, using a
 * button-index -> bit mapping. The Gamepad API has no press/release
 * events - callers must poll bitmask() on every frame.
 */
export class GamepadState {
    constructor(mapping) {
        this.mapping = mapping;
    }

    bitmask() {
        let mask = 0;
        const pads = navigator.getGamepads ? navigator.getGamepads() : [];
        for (const pad of pads) {
            if (!pad || !pad.connected) continue;
            for (const [indexStr, bit] of Object.entries(this.mapping)) {
                const button = pad.buttons[Number(indexStr)];
                if (button && button.pressed) {
                    mask |= 1 << bit;
                }
            }
        }
        return mask;
    }
}

/** True if at least one gamepad is currently connected. */
export function hasGamepad() {
    if (!navigator.getGamepads) return false;
    return [...navigator.getGamepads()].some((pad) => pad && pad.connected);
}
