import './footer.css';
import {APP_VERSION} from '../../version.js';
import {t} from '../../i18n/i18n.js';
import {loadAppLocale} from '../../i18n/storage.js';

/**
 * Renders the thin status bar pinned to the bottom of the window: a
 * credit line on the left, a centered label, and the app version on the
 * right.
 * @param {HTMLElement} container
 */
export function mountFooter(container) {
    const el = document.createElement('div');
    el.className = 'footer';

    const credits = document.createElement('div');
    credits.className = 'footer__credits';
    credits.textContent = 'Espaze';

    const downloads = document.createElement('div');
    downloads.className = 'footer__downloads';
    downloads.innerHTML = `<i class="fa-solid fa-arrow-down"></i> ${t(loadAppLocale(), 'footerDownloads')}`;

    const version = document.createElement('div');
    version.className = 'footer__version';
    version.textContent = APP_VERSION;

    el.append(credits, downloads, version);
    container.appendChild(el);
}
