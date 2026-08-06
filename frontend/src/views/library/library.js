import './library.css';
import {createGameCard} from '../../components/game-card/game-card.js';
import {createGameRow} from '../../components/game-row/game-row.js';
import {systemName} from '../../systems/names.js';

const SORTS = {
    alpha: {label: 'Ordre alphabétique', compare: (a, b) => a.title.localeCompare(b.title)},
    added: {label: 'Récemment ajoutés', compare: (a, b) => new Date(b.addedAt) - new Date(a.addedAt)},
    played: {label: 'Récemment joués', compare: (a, b) => new Date(b.lastPlayedAt ?? 0) - new Date(a.lastPlayedAt ?? 0)},
    playtime: {label: 'Temps de jeu', compare: (a, b) => b.playTimeSeconds - a.playTimeSeconds},
};

/**
 * Renders the library's main pane (shelves + sortable grid) into
 * container for an already-fetched list of games. Pure render: the
 * sidebar and the games list itself live one level up in app.js, shared
 * across the library and game-detail views.
 * @param {HTMLElement} container
 * @param {Array} games
 * @param {(game:object)=>void} onLaunch
 */
export function mountLibraryMain(container, games, onLaunch) {
    const main = buildMain();
    container.appendChild(main.el);

    renderShelves();
    main.gridTitle.textContent = `Tous les jeux (${games.length})`;
    renderGrid();
    main.sortSelect.addEventListener('change', renderGrid);

    function renderShelves() {
        main.shelves.innerHTML = '';
        const recentlyAdded = [...games].sort(SORTS.added.compare).slice(0, 10);
        const recentlyPlayed = games.filter((g) => g.lastPlayedAt).sort(SORTS.played.compare).slice(0, 10);
        const mostPlayed = games.filter((g) => g.playTimeSeconds > 0).sort(SORTS.playtime.compare).slice(0, 10);

        const addedRow = createGameRow('Ajoutés récemment', recentlyAdded, onLaunch, {variant: 'banner'});
        if (addedRow) main.shelves.appendChild(addedRow);

        const mostPlayedRow = createGameRow('Le plus joué', mostPlayed, onLaunch, {variant: 'cover'});
        if (mostPlayedRow) main.shelves.appendChild(mostPlayedRow);

        const playedRow = createGameRow('Jeux récents', recentlyPlayed, onLaunch, {variant: 'cover'});
        if (playedRow) main.shelves.appendChild(playedRow);
    }

    function renderGrid() {
        const sorted = [...games].sort(SORTS[main.sortSelect.value].compare);
        main.grid.innerHTML = '';

        if (sorted.length === 0) {
            const empty = document.createElement('div');
            empty.className = 'library__empty';
            empty.textContent = 'Aucun jeu pour le moment. Ajoute un dossier contenant des ROMs.';
            main.grid.appendChild(empty);
            return;
        }
        for (const game of sorted) {
            main.grid.appendChild(createGameCard(game, onLaunch));
        }
    }
}

/**
 * Renders every game as cards grouped into one section per console,
 * instead of the mixed shelves + flat grid mountLibraryMain shows.
 * @param {HTMLElement} container
 * @param {Array} games
 * @param {(game:object)=>void} onLaunch
 */
export function mountGroupedGrid(container, games, onLaunch) {
    const root = document.createElement('div');
    root.className = 'library__main';

    const groups = new Map();
    for (const game of games) {
        if (!groups.has(game.system)) {
            groups.set(game.system, []);
        }
        groups.get(game.system).push(game);
    }

    if (groups.size === 0) {
        const empty = document.createElement('div');
        empty.className = 'library__empty';
        empty.textContent = 'Aucun jeu pour le moment. Ajoute un dossier contenant des ROMs.';
        root.appendChild(empty);
    }

    const sortedSystems = [...groups.keys()].sort((a, b) => systemName(a).localeCompare(systemName(b)));
    for (const system of sortedSystems) {
        const systemGames = groups.get(system).sort((a, b) => a.title.localeCompare(b.title));

        const section = document.createElement('div');
        section.className = 'library__console-section';

        const heading = document.createElement('div');
        heading.className = 'library__grid-title';
        heading.textContent = `${systemName(system)} (${systemGames.length})`;

        const grid = document.createElement('div');
        grid.className = 'library__grid';
        for (const game of systemGames) {
            grid.appendChild(createGameCard(game, onLaunch));
        }

        section.append(heading, grid);
        root.appendChild(section);
    }

    container.appendChild(root);
}

function buildMain() {
    const el = document.createElement('div');
    el.className = 'library__main';

    const heroFade = document.createElement('div');
    heroFade.className = 'library__hero-fade';

    const shelves = document.createElement('div');
    shelves.className = 'library__shelves';

    const gridHeader = document.createElement('div');
    gridHeader.className = 'library__grid-header';
    const gridTitle = document.createElement('div');
    gridTitle.className = 'library__grid-title';
    const sortSelect = document.createElement('select');
    sortSelect.className = 'library__sort';
    for (const [key, {label}] of Object.entries(SORTS)) {
        const opt = document.createElement('option');
        opt.value = key;
        opt.textContent = label;
        sortSelect.appendChild(opt);
    }
    gridHeader.append(gridTitle, sortSelect);

    const grid = document.createElement('div');
    grid.className = 'library__grid';

    el.append(heroFade, shelves, gridHeader, grid);

    return {el, shelves, gridTitle, sortSelect, grid};
}
