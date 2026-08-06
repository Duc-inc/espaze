import {libraryFolders, pickAndAddLibraryFolder, removeLibraryFolder} from '../../api/games.js';
import {t} from '../../i18n/i18n.js';
import {loadAppLocale} from '../../i18n/storage.js';

/**
 * Renders the list of tracked ROM folders, with add/remove controls.
 * @param {HTMLElement} container
 */
export async function mountLibrarySection(container) {
    const locale = loadAppLocale();

    const addBtn = document.createElement('button');
    addBtn.className = 'primary settings__add-folder';
    addBtn.innerHTML = `<i class="fa-solid fa-folder-plus"></i> ${t(locale, 'librarySectionAddFolder')}`;

    const list = document.createElement('div');
    list.className = 'settings__list';

    container.append(addBtn, list);

    async function refresh() {
        const folders = await libraryFolders();
        list.innerHTML = '';

        if (folders.length === 0) {
            const empty = document.createElement('div');
            empty.className = 'settings__empty';
            empty.textContent = t(locale, 'librarySectionEmpty');
            list.appendChild(empty);
            return;
        }

        for (const folder of folders) {
            const row = document.createElement('div');
            row.className = 'settings__row';

            const path = document.createElement('div');
            path.className = 'settings__row-label settings__row-path';
            path.textContent = folder;
            path.title = folder;

            const removeBtn = document.createElement('button');
            removeBtn.className = 'settings__row-remove';
            removeBtn.title = t(locale, 'librarySectionRemove');
            removeBtn.innerHTML = '<i class="fa-solid fa-trash"></i>';
            removeBtn.addEventListener('click', async () => {
                removeBtn.disabled = true;
                await removeLibraryFolder(folder);
                await refresh();
            });

            row.append(path, removeBtn);
            list.appendChild(row);
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

    await refresh();
}
