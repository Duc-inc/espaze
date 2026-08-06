import './player.css';
import {createCanvasRenderer} from './canvas.js';
import {createSaveSlotsPanel} from './savestates.js';
import {onFrame} from '../../api/events.js';
import {launchGame, pauseGame, resumeGame, sendInput, stopGame} from '../../api/emulation.js';
import {KeyState} from '../../input/keymap.js';
import {buttonsForSystem} from '../../input/buttons.js';
import {loadKeymap} from '../../input/storage.js';

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

    let paused = false;

    const handleKeyDown = (e) => forwardKey(e, keyState, keyState.handleKeyDown.bind(keyState));
    const handleKeyUp = (e) => forwardKey(e, keyState, keyState.handleKeyUp.bind(keyState));

    window.addEventListener('keydown', handleKeyDown);
    window.addEventListener('keyup', handleKeyUp);

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
        window.removeEventListener('keydown', handleKeyDown);
        window.removeEventListener('keyup', handleKeyUp);
        unsubscribeFrame();
        await stopGame();
    }

    await launchGame(game.id);

    return {cleanup};
}

function forwardKey(event, keyState, apply) {
    if (!apply(event.code)) return;
    event.preventDefault();
    sendInput(keyState.bitmask());
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

    const legend = buildLegend(system);

    toolbar.append(backBtn, pauseBtn, savesBtn, legend);

    const savesPanel = createSaveSlotsPanel();

    const stage = document.createElement('div');
    stage.className = 'player__stage';

    const canvas = document.createElement('canvas');
    canvas.className = 'player__canvas';
    canvas.width = CANVAS_WIDTH;
    canvas.height = CANVAS_HEIGHT;

    stage.appendChild(canvas);
    el.append(toolbar, savesPanel.el, stage);

    return {el, canvas, backBtn, pauseBtn, savesBtn, savesPanel};
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
