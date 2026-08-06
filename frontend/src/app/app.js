import {mountSidebar} from '../views/sidebar/sidebar.js';
import {mountLibrary} from '../views/library/library.js';
import {mountPlayer} from '../views/player/player.js';

/**
 * Boots the whole UI into root: a fixed sidebar plus a content area that
 * swaps between the library grid and the in-app player.
 * @param {HTMLElement} root
 */
export async function createApp(root) {
    root.innerHTML = '';

    const sidebarContainer = document.createElement('div');
    const content = document.createElement('div');
    content.className = 'app__content';
    root.append(sidebarContainer, content);

    const sidebar = mountSidebar(sidebarContainer, () => showLibrary());

    let activePlayer = null;

    async function showLibrary() {
        if (activePlayer) {
            await activePlayer.cleanup();
            activePlayer = null;
        }
        sidebar.setActive('library');
        content.innerHTML = '';
        await mountLibrary(content, (gameId) => showPlayer(gameId));
    }

    async function showPlayer(gameId) {
        sidebar.setActive(null);
        content.innerHTML = '';
        activePlayer = await mountPlayer(content, gameId, showLibrary);
    }

    await showLibrary();
}
