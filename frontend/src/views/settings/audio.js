import {loadMuted, loadVolume, saveMuted, saveVolume} from '../../audio/storage.js';
import {t} from '../../i18n/i18n.js';
import {loadAppLocale} from '../../i18n/storage.js';

/**
 * Renders the audio settings: a volume slider and a mute toggle. Both
 * are picked up the next time a game is launched, the same as key
 * bindings in Contrôles - there's no running player to update live from
 * here.
 * @param {HTMLElement} container
 */
export function mountAudioSection(container) {
    const locale = loadAppLocale();

    const row = document.createElement('div');
    row.className = 'settings__row';

    const label = document.createElement('div');
    label.className = 'settings__row-label';
    label.textContent = t(locale, 'audioVolume');

    const controls = document.createElement('div');
    controls.className = 'settings__volume-controls';

    const muteBtn = document.createElement('button');
    muteBtn.className = 'settings__mute-btn';

    const slider = document.createElement('input');
    slider.type = 'range';
    slider.min = '0';
    slider.max = '100';
    slider.className = 'settings__volume-slider';

    const valueLabel = document.createElement('span');
    valueLabel.className = 'settings__volume-value';

    function refresh() {
        const muted = loadMuted();
        const volume = loadVolume();
        slider.value = String(Math.round(volume * 100));
        slider.disabled = muted;
        valueLabel.textContent = muted ? t(locale, 'audioMuted') : `${Math.round(volume * 100)}%`;
        muteBtn.classList.toggle('active', muted);
        muteBtn.title = muted ? t(locale, 'audioUnmute') : t(locale, 'audioMute');
        muteBtn.innerHTML = muted
            ? '<i class="fa-solid fa-volume-xmark"></i>'
            : '<i class="fa-solid fa-volume-high"></i>';
    }

    slider.addEventListener('input', () => {
        saveVolume(Number(slider.value) / 100);
        refresh();
    });

    muteBtn.addEventListener('click', () => {
        saveMuted(!loadMuted());
        refresh();
    });

    refresh();
    controls.append(muteBtn, slider, valueLabel);
    row.append(label, controls);
    container.appendChild(row);
}
