import './sidebar.css';
import {systemColor} from '../../systems/colors.js';

/**
 * Renders the left rail: a "back to library" strip, a content-type
 * selector with "add folder" beside it, search with "rescan" beside it,
 * and a flat alphabetical game list colored by system.
 * @param {HTMLElement} container
 * @param {{onLaunch:(id:string)=>void, onHome:()=>void, onGridView:()=>void, onAddFolder:()=>Promise<void>, onRescan:()=>Promise<void>}} handlers
 */
export function mountSidebar(container, {onLaunch, onHome, onGridView, onAddFolder, onRescan}) {
    const el = document.createElement('div');
    el.className = 'sidebar';

    const head = document.createElement('div');
    head.className = 'sidebar__head';

    const homeBtn = document.createElement('button');
    homeBtn.className = 'sidebar__head-home';
    homeBtn.textContent = 'Accueil';
    homeBtn.addEventListener('click', onHome);

    const gridBtn = document.createElement('button');
    gridBtn.className = 'sidebar__head-grid';
    gridBtn.title = 'Tous les jeux, rangés par console';
    gridBtn.innerHTML = '<i class="fa-solid fa-table-cells"></i>';
    gridBtn.addEventListener('click', onGridView);

    head.append(homeBtn, gridBtn);

    const filter = document.createElement('div');
    filter.className = 'sidebar__filter';
    filter.innerHTML = '<select><option>Tous les jeux/catégories</option></select>';
    const addFolderBtn = document.createElement('button');
    addFolderBtn.className = 'sidebar__icon-btn';
    addFolderBtn.title = 'Ajouter un dossier';
    addFolderBtn.innerHTML = '<i class="fa-solid fa-folder-plus"></i>';
    filter.appendChild(addFolderBtn);

    const tools = document.createElement('div');
    tools.className = 'sidebar__tools';
    const search = document.createElement('div');
    search.className = 'sidebar__search';
    search.innerHTML = '<i class="fa-solid fa-magnifying-glass"></i>';
    const input = document.createElement('input');
    input.type = 'text';
    input.placeholder = 'Rechercher un jeu';
    search.appendChild(input);
    const rescanBtn = document.createElement('button');
    rescanBtn.className = 'sidebar__icon-btn';
    rescanBtn.title = 'Rescanner';
    rescanBtn.innerHTML = '<i class="fa-solid fa-arrows-rotate"></i>';
    tools.append(search, rescanBtn);

    const groupTitle = document.createElement('div');
    groupTitle.className = 'sidebar__group-title';

    const list = document.createElement('div');
    list.className = 'sidebar__list';

    el.append(head, filter, tools, groupTitle, list);
    container.appendChild(el);

    let games = [];

    function render() {
        const query = input.value.trim().toLowerCase();
        const visible = query
            ? games.filter((g) => g.title.toLowerCase().includes(query))
            : games;

        groupTitle.textContent = `Tous les jeux (${visible.length}/${games.length})`;
        list.innerHTML = '';

        if (visible.length === 0) {
            const empty = document.createElement('div');
            empty.className = 'sidebar__empty';
            empty.textContent = games.length === 0 ? 'Aucun jeu.' : 'Aucun résultat.';
            list.appendChild(empty);
            return;
        }

        for (const game of [...visible].sort((a, b) => a.title.localeCompare(b.title))) {
            const item = document.createElement('div');
            item.className = 'sidebar__item';

            const dot = document.createElement('span');
            dot.className = 'sidebar__dot';
            dot.style.background = `linear-gradient(135deg, ${systemColor(game.system)}, var(--color-bg-card))`;

            const label = document.createElement('span');
            label.textContent = game.title;
            label.title = game.title;

            item.append(dot, label);
            item.addEventListener('click', () => onLaunch(game));
            list.appendChild(item);
        }
    }

    input.addEventListener('input', render);
    addFolderBtn.addEventListener('click', onAddFolder);

    const rescanIcon = rescanBtn.querySelector('i');
    rescanBtn.addEventListener('click', async () => {
        rescanBtn.disabled = true;
        rescanIcon.classList.add('fa-spin');
        try {
            await onRescan();
        } finally {
            rescanIcon.classList.remove('fa-spin');
            rescanBtn.disabled = false;
        }
    });

    render();

    return {
        setGames(next) {
            games = next;
            render();
        },
    };
}
