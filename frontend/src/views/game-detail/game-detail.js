import './game-detail.css';
import {systemColor} from '../../systems/colors.js';
import {t} from '../../i18n/i18n.js';
import {loadAppLocale} from '../../i18n/storage.js';

/**
 * Renders a game's detail page: a colored hero with its title and a
 * Play button, plus a few real stats. No store page, achievements or
 * trading cards - we don't have any of that data, so we don't fake it.
 * @param {HTMLElement} container
 * @param {object} game
 * @param {() => void} onPlay
 */
export function mountGameDetail(container, game, onPlay) {
    const locale = loadAppLocale();

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
    playBtn.innerHTML = `<i class="fa-solid fa-play"></i> ${t(locale, 'detailPlay')}`;
    playBtn.addEventListener('click', onPlay);

    const badge = document.createElement('div');
    badge.className = 'detail__badge';
    badge.textContent = game.system;

    actions.append(playBtn, badge);
    hero.append(title, actions);

    const stats = document.createElement('div');
    stats.className = 'detail__stats';
    stats.append(
        buildStat(t(locale, 'detailStatSystem'), game.system),
        buildStat(t(locale, 'detailStatAdded'), formatDate(game.addedAt, locale)),
        buildStat(t(locale, 'detailStatLastPlayed'), game.lastPlayedAt ? formatDate(game.lastPlayedAt, locale) : t(locale, 'detailStatNeverPlayed')),
        buildStat(t(locale, 'detailStatPlaytime'), formatPlaytime(game.playTimeSeconds, locale)),
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

function formatDate(isoString, locale) {
    return new Date(isoString).toLocaleDateString(locale === 'en' ? 'en-US' : 'fr-FR', {
        day: '2-digit',
        month: 'long',
        year: 'numeric',
    });
}

function formatPlaytime(totalSeconds, locale) {
    const hours = totalSeconds / 3600;
    if (hours < 0.1) {
        return t(locale, 'detailPlaytimeLessThanHour');
    }
    return t(locale, 'detailPlaytimeHours', {hours: hours.toFixed(1)});
}
