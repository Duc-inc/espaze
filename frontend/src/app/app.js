import {mountHeader} from '../views/header/header.js';
import {mountSubheader} from '../views/header/subheader.js';
import {mountFooter} from '../views/footer/footer.js';
import {mountSidebar} from '../views/sidebar/sidebar.js';
import {mountLibraryMain, mountGroupedGrid} from '../views/library/library.js';
import {mountGameDetail} from '../views/game-detail/game-detail.js';
import {mountSettings} from '../views/settings/settings.js';
import {mountPlayer} from '../views/player/player.js';
import {listGames, pickAndAddLibraryFolder, rescanLibrary} from '../api/games.js';

/**
 * Boots the whole UI into root: a brand bar, a nav sub-bar, and a bottom
 * status bar around a content area. The library grid and a game's detail
 * page share one persistent sidebar (the "shell"); only the in-app
 * player takes over the full content area, since it needs all the room
 * it can get.
 * @param {HTMLElement} root
 */
export async function createApp(root) {
    root.innerHTML = '';

    const headerContainer = document.createElement('div');
    const subheaderContainer = document.createElement('div');
    const content = document.createElement('div');
    content.className = 'app__content';
    const footerContainer = document.createElement('div');

    root.append(headerContainer, subheaderContainer, content, footerContainer);

    mountHeader(headerContainer);
    const subheader = mountSubheader(subheaderContainer, () => showLibrary(), () => showSettings());
    mountFooter(footerContainer);

    let activePlayer = null;
    let games = [];
    let shellSidebar = null;
    let shellStage = null;

    async function ensureShell() {
        subheaderContainer.style.display = '';
        footerContainer.style.display = '';

        if (shellStage) {
            return;
        }
        content.innerHTML = '';
        const shell = document.createElement('div');
        shell.className = 'library';

        const sidebarContainer = document.createElement('div');
        shellStage = document.createElement('div');
        shellStage.className = 'library__stage';
        shell.append(sidebarContainer, shellStage);
        content.appendChild(shell);

        shellSidebar = mountSidebar(sidebarContainer, {
            onLaunch: (game) => showDetail(game),
            onHome: () => showLibrary(),
            onGridView: () => showGridView(),
            onAddFolder: async () => {
                const result = await pickAndAddLibraryFolder();
                if (result) await refreshGames();
            },
            onRescan: async () => {
                await rescanLibrary();
                await refreshGames();
            },
        });
    }

    async function refreshGames() {
        games = await listGames();
        shellSidebar.setGames(games);
    }

    async function showLibrary() {
        if (activePlayer) {
            await activePlayer.cleanup();
            activePlayer = null;
        }
        subheader.setActive('library');
        await ensureShell();
        await refreshGames();
        shellStage.innerHTML = '';
        mountLibraryMain(shellStage, games, (game) => showDetail(game));
    }

    async function showGridView() {
        if (activePlayer) {
            await activePlayer.cleanup();
            activePlayer = null;
        }
        subheader.setActive('library');
        await ensureShell();
        await refreshGames();
        shellStage.innerHTML = '';
        mountGroupedGrid(shellStage, games, (game) => showDetail(game));
    }

    async function showDetail(game) {
        subheader.setActive('library');
        await ensureShell();
        shellStage.innerHTML = '';
        mountGameDetail(shellStage, game, () => showPlayer(game), showLibrary);
    }

    async function showSettings() {
        if (activePlayer) {
            await activePlayer.cleanup();
            activePlayer = null;
        }
        subheader.setActive(null);
        subheaderContainer.style.display = '';
        footerContainer.style.display = '';
        shellStage = null;
        content.innerHTML = '';
        mountSettings(content, showLibrary);
    }

    async function showPlayer(game) {
        subheader.setActive(null);
        subheaderContainer.style.display = 'none';
        footerContainer.style.display = 'none';
        shellStage = null;
        content.innerHTML = '';
        activePlayer = await mountPlayer(content, game, showLibrary);
    }

    await showLibrary();
}
