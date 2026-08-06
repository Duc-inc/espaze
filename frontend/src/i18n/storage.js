const APP_LOCALE_KEY = 'espaze:locale:app';
const PLAYER_LOCALE_KEY = 'espaze:locale:player';

export function loadAppLocale() {
    return localStorage.getItem(APP_LOCALE_KEY) || 'fr';
}

export function saveAppLocale(locale) {
    localStorage.setItem(APP_LOCALE_KEY, locale);
}

export function loadPlayerLocale() {
    return localStorage.getItem(PLAYER_LOCALE_KEY) || 'fr';
}

export function savePlayerLocale(locale) {
    localStorage.setItem(PLAYER_LOCALE_KEY, locale);
}
