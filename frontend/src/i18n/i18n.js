import {fr} from './translations/fr.js';
import {en} from './translations/en.js';

const DICTS = {fr, en};

// The app UI and the in-game player screen can be set to different
// languages independently (a real user preference, not just an app-wide
// toggle), so callers pass whichever locale applies to their scope.
export const SUPPORTED_LOCALES = [
    {id: 'fr', label: 'Français'},
    {id: 'en', label: 'English'},
];

export function t(locale, key, vars) {
    const dict = DICTS[locale] || DICTS.fr;
    let str = dict[key] ?? DICTS.fr[key] ?? key;
    if (vars) {
        for (const [name, value] of Object.entries(vars)) {
            str = str.replaceAll(`{${name}}`, String(value));
        }
    }
    return str;
}
