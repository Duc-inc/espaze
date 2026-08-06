import './game-row.css';
import {systemColor} from '../../systems/colors.js';
import {applyCoverArt} from '../cover-art.js';

/**
 * Builds one horizontal shelf ("Ajoutés récemment", "Jeux récents"): a
 * title and a scrollable row of cards. Each card is a color swatch (no
 * real box art to show) with the title overlaid at the bottom, plus a
 * small meta line below. Returns null when there's nothing to show, so
 * callers can skip an empty shelf instead of rendering a title over
 * nothing.
 * @param {string} title
 * @param {Array} games
 * @param {(id:string)=>void} onLaunch
 * @param {{variant: 'banner'|'cover'}} options
 */
export function createGameRow(title, games, onLaunch, options) {
    if (games.length === 0) {
        return null;
    }

    const section = document.createElement('div');
    section.className = 'game-row';

    const heading = document.createElement('div');
    heading.className = 'game-row__title';
    heading.textContent = title;

    const track = document.createElement('div');
    track.className = 'game-row__track';
    for (const game of games) {
        track.appendChild(createCard(game, onLaunch, options.variant));
    }

    section.append(heading, track);
    return section;
}

function createCard(game, onLaunch, variant) {
    const color = systemColor(game.system);

    const card = document.createElement('div');
    card.className = `game-row__card game-row__card--${variant}`;

    const cover = document.createElement('div');
    cover.className = 'game-row__cover';
    cover.style.background = `linear-gradient(135deg, ${color}, var(--color-bg-card) 75%)`;

    const covTitle = document.createElement('div');
    covTitle.className = 'game-row__cover-title';
    covTitle.textContent = game.title;
    cover.appendChild(covTitle);

    if (variant === 'cover' && game.playTimeSeconds > 0) {
        const covSub = document.createElement('div');
        covSub.className = 'game-row__cover-sub';
        covSub.textContent = formatPlaytime(game.playTimeSeconds);
        cover.appendChild(covSub);
    }

    const meta = document.createElement('div');
    meta.className = 'game-row__meta';
    const dot = document.createElement('span');
    dot.className = 'game-row__meta-dot';
    dot.style.background = color;
    const label = document.createElement('span');
    label.textContent = game.system;
    meta.append(dot, label);

    card.append(cover, meta);
    card.addEventListener('click', () => onLaunch(game));

    applyCoverArt(cover, game);

    return card;
}

function formatPlaytime(totalSeconds) {
    const hours = totalSeconds / 3600;
    if (hours < 0.1) {
        return "MOINS D'UNE HEURE";
    }
    return `${hours.toFixed(1)} HEURES`;
}
