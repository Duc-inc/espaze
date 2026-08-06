import './savestates.css';
import {listSaveSlots, loadStateFromSlot, saveStateToSlot} from '../../api/emulation.js';

/**
 * Builds the collapsible save-slot panel shown under the player toolbar:
 * a fixed row of slots, each with its own Save/Load buttons.
 */
export function createSaveSlotsPanel() {
    const el = document.createElement('div');
    el.className = 'savestates';
    el.hidden = true;

    async function refresh() {
        const slots = await listSaveSlots();
        el.innerHTML = '';
        for (const slot of slots) {
            el.appendChild(buildSlotCard(slot, refresh));
        }
    }

    async function toggle() {
        el.hidden = !el.hidden;
        if (!el.hidden) {
            await refresh();
        }
    }

    return {el, toggle, refresh};
}

function buildSlotCard(slot, refresh) {
    const card = document.createElement('div');
    card.className = 'savestates__slot';

    const label = document.createElement('div');
    label.className = 'savestates__label';
    label.textContent = `Emplacement ${slot.slot + 1}`;

    const status = document.createElement('div');
    status.className = 'savestates__status';
    status.textContent = slot.savedAt ? formatDate(slot.savedAt) : 'Vide';

    const actions = document.createElement('div');
    actions.className = 'savestates__actions';

    const saveBtn = document.createElement('button');
    saveBtn.textContent = 'Sauvegarder';
    saveBtn.addEventListener('click', async () => {
        saveBtn.disabled = true;
        try {
            await saveStateToSlot(slot.slot);
            await refresh();
        } finally {
            saveBtn.disabled = false;
        }
    });

    const loadBtn = document.createElement('button');
    loadBtn.className = 'primary';
    loadBtn.textContent = 'Charger';
    loadBtn.disabled = !slot.savedAt;
    loadBtn.addEventListener('click', async () => {
        loadBtn.disabled = true;
        try {
            await loadStateFromSlot(slot.slot);
        } finally {
            loadBtn.disabled = false;
        }
    });

    actions.append(saveBtn, loadBtn);
    card.append(label, status, actions);
    return card;
}

function formatDate(isoString) {
    return new Date(isoString).toLocaleString('fr-FR', {
        day: '2-digit',
        month: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
    });
}
