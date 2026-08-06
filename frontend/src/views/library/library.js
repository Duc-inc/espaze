import './library.css';
import {createGameCard} from '../../components/game-card/game-card.js';
import {listGames, pickAndAddLibraryFolder, rescanLibrary} from '../../api/games.js';

/**
 * Renders the Steam-like game grid into container.
 * @param {HTMLElement} container
 * @param {(gameId:string)=>void} onLaunch
 */
export async function mountLibrary(container, onLaunch) {
    const root = document.createElement('div');
    root.className = 'library';

    const header = document.createElement('div');
    header.className = 'library__header';

    const title = document.createElement('div');
    title.className = 'library__title';
    title.textContent = 'Bibliothèque';

    const actions = document.createElement('div');
    actions.className = 'library__actions';

    const addBtn = document.createElement('button');
    addBtn.className = 'primary';
    addBtn.textContent = '+ Ajouter un dossier';

    const rescanBtn = document.createElement('button');
    rescanBtn.textContent = 'Rescanner';

    actions.append(addBtn, rescanBtn);
    header.append(title, actions);

    const grid = document.createElement('div');
    grid.className = 'library__grid';

    root.append(header, grid);
    container.appendChild(root);

    async function refresh() {
        const games = await listGames();
        grid.innerHTML = '';

        if (games.length === 0) {
            const empty = document.createElement('div');
            empty.className = 'library__empty';
            empty.textContent = 'Aucun jeu pour le moment. Ajoute un dossier contenant des ROMs.';
            grid.appendChild(empty);
            return;
        }

        for (const game of games) {
            grid.appendChild(createGameCard(game, onLaunch));
        }
    }

    addBtn.addEventListener('click', async () => {
        addBtn.disabled = true;
        try {
            const result = await pickAndAddLibraryFolder();
            if (result) await refresh();
        } finally {
            addBtn.disabled = false;
        }
    });

    rescanBtn.addEventListener('click', async () => {
        rescanBtn.disabled = true;
        try {
            await rescanLibrary();
            await refresh();
        } finally {
            rescanBtn.disabled = false;
        }
    });

    await refresh();

    return {refresh};
}
