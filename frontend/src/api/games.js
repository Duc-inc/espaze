// Thin wrapper around the generated Wails bindings for library/game
// concerns. Views never import wailsjs directly - they go through here.
import {
    AddLibraryFolder,
    AvailableSystems,
    BrowseForLibraryFolder,
    LibraryFolders,
    ListGames,
    RemoveGame,
    RemoveLibraryFolder,
    RescanLibrary,
} from '../../wailsjs/go/app/App';

export async function listGames() {
    return (await ListGames()) ?? [];
}

export async function libraryFolders() {
    return (await LibraryFolders()) ?? [];
}

export async function availableSystems() {
    return (await AvailableSystems()) ?? [];
}

export async function pickAndAddLibraryFolder() {
    const folder = await BrowseForLibraryFolder();
    if (!folder) {
        return null;
    }
    const added = await AddLibraryFolder(folder);
    return {folder, added};
}

export async function rescanLibrary() {
    return await RescanLibrary();
}

export async function removeGame(id) {
    return await RemoveGame(id);
}

export async function removeLibraryFolder(path) {
    return await RemoveLibraryFolder(path);
}
