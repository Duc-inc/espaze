import {availableSystems} from '../../api/games.js';
import {systemColor} from '../../systems/colors.js';
import {APP_VERSION} from '../../version.js';

/**
 * Renders app info and the list of supported emulation systems, read
 * straight from the Go core registry - never hand-maintained here.
 * @param {HTMLElement} container
 */
export async function mountAboutSection(container) {
    const heading = document.createElement('div');
    heading.className = 'settings__about-heading';
    heading.innerHTML = `<span class="settings__about-name">Espaze</span> <span class="settings__about-version">${APP_VERSION}</span>`;

    const tagline = document.createElement('div');
    tagline.className = 'settings__about-tagline';
    tagline.textContent = 'Un launcher unique pour plusieurs émulateurs maison, écrit en Go + Wails.';

    const list = document.createElement('div');
    list.className = 'settings__list';

    container.append(heading, tagline, list);

    const systems = await availableSystems();
    for (const system of systems) {
        const row = document.createElement('div');
        row.className = 'settings__row';

        const dot = document.createElement('span');
        dot.className = 'settings__system-dot';
        dot.style.background = systemColor(system.ID);

        const info = document.createElement('div');
        const name = document.createElement('div');
        name.className = 'settings__row-label';
        name.textContent = system.Name;
        const ext = document.createElement('div');
        ext.className = 'settings__system-ext';
        ext.textContent = system.Extensions.join(', ');
        info.append(name, ext);

        row.append(dot, info);
        list.appendChild(row);
    }
}
