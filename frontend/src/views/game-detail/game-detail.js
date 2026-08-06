import './game-detail.css';
import {systemColor} from '../../systems/colors.js';

/**
 * Renders a game's detail page: a colored hero with its title and a
 * Play button, plus a few real stats. No store page, achievements or
 * trading cards - we don't have any of that data, so we don't fake it.
 * @param {HTMLElement} container
 * @param {object} game
 * @param {() => void} onPlay
 */
export function mountGameDetail(container, game, onPlay) {
    const root = document.createElement('div');
    root.className = 'detail';

    const hero = document.createElement('div');
    hero.className = 'detail__hero';
    hero.style.background = `linear-gradient(135deg, ${systemColor(game.system)}, var(--color-bg-card) 70%)`;

    const title = document.createElement('div');
    title.className = 'detail__title';
    title.textContent = game.title;

    const actions = document.createElement('div');
    actions.className = 'detail__actions';

    const playBtn = document.createElement('button');
    playBtn.className = 'detail__play-btn';
    playBtn.innerHTML = '<i class="fa-solid fa-play"></i> Jouer';
    playBtn.addEventListener('click', onPlay);

    const badge = document.createElement('div');
    badge.className = 'detail__badge';
    badge.textContent = game.system;

    actions.append(playBtn, badge);
    hero.append(title, actions);

    const stats = document.createElement('div');
    stats.className = 'detail__stats';
    stats.append(
        buildStat('Système', game.system),
        buildStat('Ajouté le', formatDate(game.addedAt)),
        buildStat('Dernière partie', game.lastPlayedAt ? formatDate(game.lastPlayedAt) : 'Jamais jouée'),
        buildStat('Temps de jeu total', formatPlaytime(game.playTimeSeconds)),
    );

    root.append(hero, stats);
    container.appendChild(root);
}

function buildStat(label, value) {
    const el = document.createElement('div');
    const l = document.createElement('div');
    l.className = 'detail__stat-label';
    l.textContent = label;
    const v = document.createElement('div');
    v.className = 'detail__stat-value';
    v.textContent = value;
    el.append(l, v);
    return el;
}

function formatDate(isoString) {
    return new Date(isoString).toLocaleDateString('fr-FR', {
        day: '2-digit',
        month: 'long',
        year: 'numeric',
    });
}

function formatPlaytime(totalSeconds) {
    const hours = totalSeconds / 3600;
    if (hours < 0.1) {
        return "Moins d'une heure";
    }
    return `${hours.toFixed(1)} heures`;
}
