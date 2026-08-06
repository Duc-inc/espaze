import './settings.css';
import {mountLibrarySection} from './library-section.js';
import {mountControlsSection} from './controls.js';
import {mountAboutSection} from './about.js';

const SECTIONS = [
    {id: 'library', label: 'Bibliothèque', icon: 'fa-folder-open', mount: mountLibrarySection},
    {id: 'controls', label: 'Contrôles', icon: 'fa-gamepad', mount: mountControlsSection},
    {id: 'about', label: 'À propos', icon: 'fa-circle-info', mount: mountAboutSection},
];

/**
 * Renders the settings page: a category rail on the left (Bibliothèque,
 * Contrôles, À propos, ...) and the active section's panel on the right.
 * @param {HTMLElement} container
 * @param {() => void} onBack
 */
export function mountSettings(container, onBack) {
    const root = document.createElement('div');
    root.className = 'settings';

    const nav = document.createElement('div');
    nav.className = 'settings__nav';

    const backBtn = document.createElement('button');
    backBtn.className = 'settings__back';
    backBtn.innerHTML = '<i class="fa-solid fa-arrow-left"></i> Bibliothèque';
    backBtn.addEventListener('click', onBack);
    nav.appendChild(backBtn);

    const navButtons = new Map();
    for (const section of SECTIONS) {
        const btn = document.createElement('button');
        btn.className = 'settings__nav-item';
        btn.innerHTML = `<i class="fa-solid ${section.icon}"></i> ${section.label}`;
        btn.addEventListener('click', () => selectSection(section.id));
        navButtons.set(section.id, btn);
        nav.appendChild(btn);
    }

    const panel = document.createElement('div');
    panel.className = 'settings__panel';

    root.append(nav, panel);
    container.appendChild(root);

    function selectSection(sectionId) {
        for (const [id, btn] of navButtons) {
            btn.classList.toggle('active', id === sectionId);
        }
        panel.innerHTML = '';
        const section = SECTIONS.find((s) => s.id === sectionId);
        section.mount(panel);
    }

    selectSection(SECTIONS[0].id);
}
