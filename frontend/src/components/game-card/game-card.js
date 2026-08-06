import './game-card.css';
import {systemColor} from '../../systems/colors.js';

/**
 * Builds a single grid tile for one library entry. There's
 * no real cover art (no store, no scraping), so the cover is the
 * system's accent color with the title overlaid at the bottom, the same
 * treatment the horizontal shelves use.
 * @param {{id:string,title:string,system:string}} game
 * @param {(id:string)=>void} onLaunch
 * @returns {HTMLElement}
 */
export function createGameCard(game, onLaunch) {
    const card = document.createElement('div');
    card.className = 'game-card';
    card.dataset.gameId = game.id;

    const color = systemColor(game.system);

    const cover = document.createElement('div');
    cover.className = 'game-card__cover';
    cover.style.background = `linear-gradient(135deg, ${color}, var(--color-bg-card) 75%)`;

    const title = document.createElement('div');
    title.className = 'game-card__title';
    title.textContent = game.title;
    title.title = game.title;
    cover.appendChild(title);

    const meta = document.createElement('div');
    meta.className = 'game-card__meta';
    const dot = document.createElement('span');
    dot.className = 'game-card__dot';
    dot.style.background = color;
    const system = document.createElement('span');
    system.textContent = game.system;
    meta.append(dot, system);

    card.append(cover, meta);
    card.addEventListener('click', () => onLaunch(game));

    return card;
}
