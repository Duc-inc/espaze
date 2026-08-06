import './player.css';
import {createCanvasRenderer} from './canvas.js';
import {onFrame} from '../../api/events.js';
import {launchGame, pauseGame, resumeGame, sendInput, stopGame} from '../../api/emulation.js';
import {CHIP8_KEYMAP, KeyState} from '../../input/keymap.js';

const CANVAS_WIDTH = 640;
const CANVAS_HEIGHT = 320;

/**
 * Mounts the in-app player: renders emulator frames to a canvas, forwards
 * keyboard input to the running core, and exposes pause/stop controls.
 * @param {HTMLElement} container
 * @param {string} gameId
 * @param {() => void} onExit called after the game has fully stopped
 */
export async function mountPlayer(container, gameId, onExit) {
    const root = buildLayout();
    container.appendChild(root.el);

    const renderer = createCanvasRenderer(root.canvas);
    const unsubscribeFrame = onFrame((payload) => renderer.drawFrame(payload));
    const keyState = new KeyState(CHIP8_KEYMAP);

    let paused = false;

    const handleKeyDown = (e) => forwardKey(e, keyState, keyState.handleKeyDown.bind(keyState));
    const handleKeyUp = (e) => forwardKey(e, keyState, keyState.handleKeyUp.bind(keyState));

    window.addEventListener('keydown', handleKeyDown);
    window.addEventListener('keyup', handleKeyUp);

    root.pauseBtn.addEventListener('click', async () => {
        paused = !paused;
        root.pauseBtn.textContent = paused ? 'Reprendre' : 'Pause';
        await (paused ? pauseGame() : resumeGame());
    });

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

    await launchGame(gameId);

    return {cleanup};
}

function forwardKey(event, keyState, apply) {
    if (!apply(event.code)) return;
    event.preventDefault();
    sendInput(keyState.bitmask());
}

function buildLayout() {
    const el = document.createElement('div');
    el.className = 'player';

    const toolbar = document.createElement('div');
    toolbar.className = 'player__toolbar';

    const backBtn = document.createElement('button');
    backBtn.textContent = '← Bibliothèque';

    const pauseBtn = document.createElement('button');
    pauseBtn.textContent = 'Pause';

    toolbar.append(backBtn, pauseBtn);

    const stage = document.createElement('div');
    stage.className = 'player__stage';

    const canvas = document.createElement('canvas');
    canvas.className = 'player__canvas';
    canvas.width = CANVAS_WIDTH;
    canvas.height = CANVAS_HEIGHT;

    stage.appendChild(canvas);
    el.append(toolbar, stage);

    return {el, canvas, backBtn, pauseBtn};
}
