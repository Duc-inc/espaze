import {getArtwork} from '../api/games.js';

/**
 * Swaps a card's color-gradient cover for a real cover image, if the
 * game has one (a user-provided image file sitting next to its ROM).
 * Does nothing when there's none, so the gradient placeholder stays.
 * @param {HTMLElement} coverEl
 * @param {{id:string, artworkPath?:string}} game
 */
export async function applyCoverArt(coverEl, game) {
    if (!game.artworkPath) {
        return;
    }
    const dataUri = await getArtwork(game.id);
    if (!dataUri) {
        return;
    }
    coverEl.style.backgroundImage = `url(${dataUri})`;
    coverEl.style.backgroundSize = 'cover';
    coverEl.style.backgroundPosition = 'center';
}
