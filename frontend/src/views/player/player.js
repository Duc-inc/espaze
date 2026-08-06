import './player.css';
import {createCanvasRenderer} from './canvas.js';
import {createSaveSlotsPanel} from './savestates.js';
import {onFrame} from '../../api/events.js';
import {launchGame, pauseGame, resumeGame, sendInput, stopGame} from '../../api/emulation.js';
import {KeyState} from '../../input/keymap.js';
import {buttonsForSystem, gamepadMapForSystem} from '../../input/buttons.js';
import {loadKeymap} from '../../input/storage.js';
import {GamepadState, hasGamepad} from '../../input/gamepad.js';

const CANVAS_WIDTH = 640;
const CANVAS_HEIGHT = 320;

/**
 * Mounts the in-app player: renders emulator frames to a canvas, forwards
 * keyboard input to the running core (using the right keymap for this
 * game's system), and exposes pause/stop controls.
 * @param {HTMLElement} container
 * @param {{id:string, system:string}} game
 * @param {() => void} onExit called after the game has fully stopped
 */
export async function mountPlayer(container, game, onExit) {
    const root = buildLayout(game.system);
    container.appendChild(root.el);

    const renderer = createCanvasRenderer(root.canvas);
    const unsubscribeFrame = onFrame((payload) => renderer.drawFrame(payload));
    const keyState = new KeyState(loadKeymap(game.system));
    const gamepadState = new GamepadState(gamepadMapForSystem(game.system));

    let paused = false;

    const handleKeyDown = (e) => {
        if (!keyState.handleKeyDown(e.code)) return;
        e.preventDefault();
    };
    const handleKeyUp = (e) => {
        if (!keyState.handleKeyUp(e.code)) return;
        e.preventDefault();
    };

    window.addEventListener('keydown', handleKeyDown);
    window.addEventListener('keyup', handleKeyUp);

    const updateGamepadStatus = () => root.gamepadStatus.classList.toggle('connected', hasGamepad());
    updateGamepadStatus();
    window.addEventListener('gamepadconnected', updateGamepadStatus);
    window.addEventListener('gamepaddisconnected', updateGamepadStatus);

    // The Gamepad API has no press/release events, so input has to be
    // polled every frame - keyboard state rides along on the same loop
    // so both sources combine into one bitmask sent only when it changes.
    let lastMask = -1;
    let pollHandle = requestAnimationFrame(pollInput);
    function pollInput() {
        const mask = keyState.bitmask() | gamepadState.bitmask();
        if (mask !== lastMask) {
            lastMask = mask;
            sendInput(mask);
        }
        pollHandle = requestAnimationFrame(pollInput);
    }

    root.pauseBtn.addEventListener('click', async () => {
        paused = !paused;
        root.pauseBtn.innerHTML = paused
            ? '<i class="fa-solid fa-play"></i> Reprendre'
            : '<i class="fa-solid fa-pause"></i> Pause';
        await (paused ? pauseGame() : resumeGame());
    });

    root.savesBtn.addEventListener('click', () => root.savesPanel.toggle());

    root.backBtn.addEventListener('click', async () => {
        await cleanup();
        onExit();
    });

    async function cleanup() {
        cancelAnimationFrame(pollHandle);
        window.removeEventListener('keydown', handleKeyDown);
        window.removeEventListener('keyup', handleKeyUp);
        window.removeEventListener('gamepadconnected', updateGamepadStatus);
        window.removeEventListener('gamepaddisconnected', updateGamepadStatus);
        unsubscribeFrame();
        await stopGame();
    }

    await launchGame(game.id);

    return {cleanup};
}

function buildLayout(system) {
    const el = document.createElement('div');
    el.className = 'player';

    const toolbar = document.createElement('div');
    toolbar.className = 'player__toolbar';

    const backBtn = document.createElement('button');
    backBtn.innerHTML = '<i class="fa-solid fa-arrow-left"></i> Bibliothèque';

    const pauseBtn = document.createElement('button');
    pauseBtn.innerHTML = '<i class="fa-solid fa-pause"></i> Pause';

    const savesBtn = document.createElement('button');
    savesBtn.innerHTML = '<i class="fa-solid fa-floppy-disk"></i> Sauvegardes';

    const gamepadStatus = document.createElement('span');
    gamepadStatus.className = 'player__gamepad-status';
    gamepadStatus.title = 'Manette';
    gamepadStatus.innerHTML = '<i class="fa-solid fa-gamepad"></i>';

    const legend = buildLegend(system);

    toolbar.append(backBtn, pauseBtn, savesBtn, gamepadStatus, legend);

    const savesPanel = createSaveSlotsPanel();

    const stage = document.createElement('div');
    stage.className = 'player__stage';

    const canvas = document.createElement('canvas');
    canvas.className = 'player__canvas';
    canvas.width = CANVAS_WIDTH;
    canvas.height = CANVAS_HEIGHT;

    stage.appendChild(canvas);
    el.append(toolbar, savesPanel.el, stage);

    return {el, canvas, backBtn, pauseBtn, savesBtn, savesPanel, gamepadStatus};
}

function buildLegend(system) {
    const legend = document.createElement('div');
    legend.className = 'player__legend';

    const keymap = loadKeymap(system);
    for (const {bit, label} of buttonsForSystem(system)) {
        const codes = Object.entries(keymap)
            .filter(([, assignedBit]) => assignedBit === bit)
            .map(([code]) => formatCode(code));
        if (codes.length === 0) continue;

        const entry = document.createElement('span');
        entry.className = 'player__legend-entry';
        entry.innerHTML = `<kbd>${codes.join('/')}</kbd> ${label}`;
        legend.appendChild(entry);
    }
    return legend;
}

const CODE_LABELS = {
    ArrowUp: '↑', ArrowDown: '↓', ArrowLeft: '←', ArrowRight: '→',
    Enter: 'Entrée', ShiftLeft: 'Maj', ShiftRight: 'Maj',
};

function formatCode(code) {
    if (CODE_LABELS[code]) return CODE_LABELS[code];
    if (code.startsWith('Key')) return code.slice(3);
    if (code.startsWith('Digit')) return code.slice(5);
    return code;
}
