import {SUPPORTED_LOCALES, t} from '../../i18n/i18n.js';
import {loadAppLocale, loadPlayerLocale, saveAppLocale, savePlayerLocale} from '../../i18n/storage.js';

/**
 * Renders the general settings: the app's own language (the library,
 * sidebar, settings, etc. - applied on the next restart, since that
 * chrome is only ever mounted once) and the emulator's language (the
 * in-game player screen - applied the next time you launch a game,
 * since that view is remounted fresh every launch).
 * @param {HTMLElement} container
 */
export function mountGeneralSection(container) {
    const locale = loadAppLocale();

    container.append(
        buildLanguageRow(
            t(locale, 'generalAppLanguage'),
            t(locale, 'generalAppLanguageHint'),
            loadAppLocale(),
            (next) => {
                saveAppLocale(next);
                window.location.reload();
            },
        ),
        buildLanguageRow(
            t(locale, 'generalPlayerLanguage'),
            t(locale, 'generalPlayerLanguageHint'),
            loadPlayerLocale(),
            (next) => savePlayerLocale(next),
        ),
    );
}

function buildLanguageRow(label, hint, currentLocale, onChange) {
    const row = document.createElement('div');
    row.className = 'settings__row settings__row--language';

    const info = document.createElement('div');
    const labelEl = document.createElement('div');
    labelEl.className = 'settings__row-label';
    labelEl.textContent = label;
    const hintEl = document.createElement('div');
    hintEl.className = 'settings__row-hint';
    hintEl.textContent = hint;
    info.append(labelEl, hintEl);

    const select = document.createElement('select');
    select.className = 'settings__language-select';
    for (const {id, label: optionLabel} of SUPPORTED_LOCALES) {
        const opt = document.createElement('option');
        opt.value = id;
        opt.textContent = optionLabel;
        opt.selected = id === currentLocale;
        select.appendChild(opt);
    }
    select.addEventListener('change', () => onChange(select.value));

    row.append(info, select);
    return row;
}
