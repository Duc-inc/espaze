import './game-card.css';

/**
 * Builds a single Steam-like tile for one library entry.
 * @param {{id:string,title:string,system:string}} game
 * @param {(id:string)=>void} onLaunch
 * @returns {HTMLElement}
 */
export function createGameCard(game, onLaunch) {
    const card = document.createElement('div');
    card.className = 'game-card';
    card.dataset.gameId = game.id;

    const artwork = document.createElement('div');
    artwork.className = 'game-card__artwork';
    artwork.textContent = initials(game.title);

    const body = document.createElement('div');
    body.className = 'game-card__body';

    const title = document.createElement('div');
    title.className = 'game-card__title';
    title.textContent = game.title;
    title.title = game.title;

    const system = document.createElement('div');
    system.className = 'game-card__system';
    system.textContent = game.system;

    body.append(title, system);
    card.append(artwork, body);

    card.addEventListener('click', () => onLaunch(game.id));

    return card;
}

function initials(title) {
    const words = title.trim().split(/\s+/).filter(Boolean);
    if (words.length === 0) return '?';
    if (words.length === 1) return words[0].slice(0, 2).toUpperCase();
    return (words[0][0] + words[1][0]).toUpperCase();
}
