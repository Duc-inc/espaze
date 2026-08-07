import {buttonsForSystem} from '../../input/buttons.js';
import {loadKeymap, resetKeymap, saveKeymap} from '../../input/storage.js';
import {t} from '../../i18n/i18n.js';
import {loadAppLocale} from '../../i18n/storage.js';

const SYSTEMS = [
    {id: 'chip8', label: 'CHIP-8'},
    {id: 'schip', label: 'Super-CHIP'},
    {id: 'gameboy', label: 'Game Boy'},
    {id: 'gbc', label: 'Game Boy Color'},
    {id: 'nes', label: 'NES'},
    {id: 'sms', label: 'Master System'},
    {id: 'genesis', label: 'Genesis / Mega Drive'},
    {id: 'gamegear', label: 'Game Gear'},
    {id: 'atari2600', label: 'Atari 2600'},
    {id: 'pcengine', label: 'PC Engine / TurboGrafx-16'},
    {id: 'gba', label: 'Game Boy Advance'},
    {id: 'ngpc', label: 'Neo Geo Pocket Color'},
    {id: 'snes', label: 'Super Nintendo'},
    {id: 'colecovision', label: 'ColecoVision'},
];

/**
 * Renders the key-binding editor: pick a system, then click any button's
 * current key to reassign it - the next key you press replaces it,
 * stored per-system in localStorage.
 * @param {HTMLElement} container
 */
export function mountControlsSection(container) {
    const locale = loadAppLocale();

    const tabs = document.createElement('div');
    tabs.className = 'settings__tabs';
    const tabButtons = new Map();
    let activeSystem = SYSTEMS[0].id;

    const resetBtn = document.createElement('button');
    resetBtn.className = 'settings__reset';
    resetBtn.innerHTML = `<i class="fa-solid fa-rotate-left"></i> ${t(locale, 'controlsReset')}`;

    const list = document.createElement('div');
    list.className = 'settings__list';

    for (const system of SYSTEMS) {
        const btn = document.createElement('button');
        btn.textContent = system.label;
        btn.addEventListener('click', () => {
            activeSystem = system.id;
            updateTabs();
            renderList();
        });
        tabButtons.set(system.id, btn);
        tabs.appendChild(btn);
    }

    function updateTabs() {
        for (const [id, btn] of tabButtons) {
            btn.classList.toggle('active', id === activeSystem);
        }
    }

    function renderList() {
        const keymap = loadKeymap(activeSystem);
        list.innerHTML = '';

        for (const {bit, label} of buttonsForSystem(activeSystem)) {
            const row = document.createElement('div');
            row.className = 'settings__row';

            const labelEl = document.createElement('div');
            labelEl.className = 'settings__row-label';
            labelEl.textContent = label;

            const keyBtn = document.createElement('button');
            keyBtn.className = 'settings__row-key';
            keyBtn.textContent = describeBinding(keymap, bit, locale);
            keyBtn.addEventListener('click', () => captureKey(keyBtn, activeSystem, bit, renderList, locale));

            row.append(labelEl, keyBtn);
            list.appendChild(row);
        }
    }

    resetBtn.addEventListener('click', () => {
        resetKeymap(activeSystem);
        renderList();
    });

    updateTabs();
    renderList();

    container.append(tabs, resetBtn, list);
}

function describeBinding(keymap, bit, locale) {
    const codes = Object.entries(keymap)
        .filter(([, assignedBit]) => assignedBit === bit)
        .map(([code]) => formatCode(code));
    return codes.length > 0 ? codes.join(' / ') : t(locale, 'controlsUnassigned');
}

function captureKey(button, systemId, bit, onDone, locale) {
    const previousLabel = button.textContent;
    button.textContent = t(locale, 'controlsPressKey');
    button.classList.add('capturing');

    function handleKeyDown(e) {
        e.preventDefault();
        window.removeEventListener('keydown', handleKeyDown, true);
        button.classList.remove('capturing');

        if (e.code === 'Escape') {
            button.textContent = previousLabel;
            return;
        }

        const keymap = loadKeymap(systemId);
        delete keymap[e.code]; // that key can't mean two things at once
        for (const code of Object.keys(keymap)) {
            if (keymap[code] === bit) delete keymap[code]; // nor can this button have two keys
        }
        keymap[e.code] = bit;
        saveKeymap(systemId, keymap);
        onDone();
    }

    window.addEventListener('keydown', handleKeyDown, true);
}

const CODE_LABELS = {
    ArrowUp: '↑', ArrowDown: '↓', ArrowLeft: '←', ArrowRight: '→',
    Enter: 'Entrée', Escape: 'Échap', Space: 'Espace',
    ShiftLeft: 'Maj G', ShiftRight: 'Maj D',
    ControlLeft: 'Ctrl G', ControlRight: 'Ctrl D',
    AltLeft: 'Alt', AltRight: 'AltGr',
    Backspace: 'Retour arrière', Tab: 'Tab',
};

function formatCode(code) {
    if (CODE_LABELS[code]) return CODE_LABELS[code];
    if (code.startsWith('Key')) return code.slice(3);
    if (code.startsWith('Digit')) return code.slice(5);
    return code;
}
// (These are physical-key labels shown as English-style shorthand in the
// French UI too - "Maj"/"Ctrl" read fine either way - so they aren't
// pulled from the locale dictionary.)
