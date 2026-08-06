import './player.css';
import {createCanvasRenderer} from './canvas.js';
import {createSaveSlotsPanel} from './savestates.js';
import {onAudio, onFrame} from '../../api/events.js';
import {launchGame, pauseGame, resumeGame, sendInput, stopGame} from '../../api/emulation.js';
import {KeyState} from '../../input/keymap.js';
import {buttonsForSystem, gamepadMapForSystem} from '../../input/buttons.js';
import {loadKeymap} from '../../input/storage.js';
import {GamepadState, hasGamepad} from '../../input/gamepad.js';
import {createAudioPlayer} from '../../audio/player.js';
import {loadMuted, loadVolume, saveMuted, saveVolume} from '../../audio/storage.js';
import {t} from '../../i18n/i18n.js';
import {loadPlayerLocale} from '../../i18n/storage.js';

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
    const locale = loadPlayerLocale();
    const root = buildLayout(game.system, locale);
    container.appendChild(root.el);

    const renderer = createCanvasRenderer(root.canvas);
    const unsubscribeFrame = onFrame((payload) => renderer.drawFrame(payload));
    const audioPlayer = createAudioPlayer();
    const unsubscribeAudio = onAudio(({sampleRate, samples}) => audioPlayer.queueSamples(sampleRate, samples));

    // Unlike key bindings and other Settings changes, volume applies to
    // this already-running audioPlayer instance directly - no reload or
    // relaunch needed.
    function refreshVolumeUI() {
        const muted = loadMuted();
        const volume = loadVolume();
        root.volumeSlider.value = String(Math.round(volume * 100));
        root.volumeSlider.disabled = muted;
        root.muteBtn.classList.toggle('active', muted);
        root.muteBtn.innerHTML = muted
            ? '<i class="fa-solid fa-volume-xmark"></i>'
            : '<i class="fa-solid fa-volume-high"></i>';
    }
    refreshVolumeUI();
    root.volumeSlider.addEventListener('input', () => {
        const volume = Number(root.volumeSlider.value) / 100;
        saveVolume(volume);
        audioPlayer.setVolume(volume);
    });
    root.muteBtn.addEventListener('click', () => {
        const muted = !loadMuted();
        saveMuted(muted);
        audioPlayer.setMuted(muted);
        refreshVolumeUI();
    });

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
            ? `<i class="fa-solid fa-play"></i> ${t(locale, 'playerResume')}`
            : `<i class="fa-solid fa-pause"></i> ${t(locale, 'playerPause')}`;
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
        unsubscribeAudio();
        await audioPlayer.stop();
        await stopGame();
    }

    await launchGame(game.id);

    return {cleanup};
}

function buildLayout(system, locale) {
    const el = document.createElement('div');
    el.className = 'player';

    const toolbar = document.createElement('div');
    toolbar.className = 'player__toolbar';

    const backBtn = document.createElement('button');
    backBtn.innerHTML = `<i class="fa-solid fa-arrow-left"></i> ${t(locale, 'playerBack')}`;

    const pauseBtn = document.createElement('button');
    pauseBtn.innerHTML = `<i class="fa-solid fa-pause"></i> ${t(locale, 'playerPause')}`;

    const savesBtn = document.createElement('button');
    savesBtn.innerHTML = `<i class="fa-solid fa-floppy-disk"></i> ${t(locale, 'playerSaves')}`;

    const gamepadStatus = document.createElement('span');
    gamepadStatus.className = 'player__gamepad-status';
    gamepadStatus.title = t(locale, 'playerGamepadTitle');
    gamepadStatus.innerHTML = '<i class="fa-solid fa-gamepad"></i>';

    const volumeGroup = document.createElement('div');
    volumeGroup.className = 'player__volume';
    const muteBtn = document.createElement('button');
    muteBtn.className = 'player__mute-btn';
    const volumeSlider = document.createElement('input');
    volumeSlider.type = 'range';
    volumeSlider.min = '0';
    volumeSlider.max = '100';
    volumeSlider.className = 'player__volume-slider';
    volumeGroup.append(muteBtn, volumeSlider);

    const legend = buildLegend(system);

    toolbar.append(backBtn, pauseBtn, savesBtn, gamepadStatus, volumeGroup, legend);

    const savesPanel = createSaveSlotsPanel(locale);

    const stage = document.createElement('div');
    stage.className = 'player__stage';

    const canvas = document.createElement('canvas');
    canvas.className = 'player__canvas';
    canvas.width = CANVAS_WIDTH;
    canvas.height = CANVAS_HEIGHT;

    stage.appendChild(canvas);
    el.append(toolbar, savesPanel.el, stage);

    return {el, canvas, backBtn, pauseBtn, savesBtn, savesPanel, gamepadStatus, muteBtn, volumeSlider};
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
