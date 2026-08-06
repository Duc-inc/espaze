import './header.css';
import {Quit, WindowMinimise, WindowToggleMaximise} from '../../../wailsjs/runtime/runtime';

/**
 * Renders the custom title bar (the window is frameless, so this
 * replaces Windows' own): a draggable strip with the app name truly
 * centered, and minimize/maximize/close on the right.
 * @param {HTMLElement} container
 */
export function mountHeader(container) {
    const el = document.createElement('div');
    el.className = 'header';
    el.style.setProperty('--wails-draggable', 'drag');

    const spacer = document.createElement('div');
    spacer.className = 'header__spacer';

    const brand = document.createElement('div');
    brand.className = 'header__brand';
    brand.textContent = 'ESPAZE';

    const controls = document.createElement('div');
    controls.className = 'header__controls';
    controls.style.setProperty('--wails-draggable', 'no-drag');

    const minBtn = makeControlButton('fa-window-minimize', WindowMinimise);
    const maxBtn = makeControlButton('fa-window-maximize', WindowToggleMaximise);
    const closeBtn = makeControlButton('fa-xmark', Quit, 'header__control--close');

    controls.append(minBtn, maxBtn, closeBtn);
    el.append(spacer, brand, controls);
    container.appendChild(el);
}

function makeControlButton(icon, onClick, extraClass) {
    const btn = document.createElement('button');
    btn.className = 'header__control' + (extraClass ? ` ${extraClass}` : '');
    btn.innerHTML = `<i class="fa-solid ${icon}"></i>`;
    btn.addEventListener('click', onClick);
    return btn;
}
