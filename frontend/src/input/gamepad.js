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

/**
 * Rumbles every connected gamepad that supports the standard haptics
 * actuator. No system currently plugged into Espaze drives this - CHIP-8,
 * Super-CHIP and the DMG Game Boy core all predate rumble hardware - it's
 * here so a future system that does (e.g. an MBC5+rumble GB cartridge,
 * or a later console) has a ready-made call.
 * @param {number} durationMs
 * @param {number} [weakMagnitude] 0-1
 * @param {number} [strongMagnitude] 0-1
 */
export function rumble(durationMs, weakMagnitude = 0.5, strongMagnitude = 1.0) {
    if (!navigator.getGamepads) return;
    for (const pad of navigator.getGamepads()) {
        if (!pad || !pad.connected || !pad.vibrationActuator) continue;
        pad.vibrationActuator.playEffect('dual-rumble', {
            duration: durationMs,
            weakMagnitude,
            strongMagnitude,
        });
    }
}
