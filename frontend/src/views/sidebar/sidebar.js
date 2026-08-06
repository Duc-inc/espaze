import './sidebar.css';

const NAV_ITEMS = [
    {id: 'library', label: 'Bibliothèque'},
];

/**
 * Renders the fixed left navigation rail into container and returns a
 * controller that lets the app switch which item is marked active.
 * @param {HTMLElement} container
 * @param {(viewId:string)=>void} onNavigate
 */
export function mountSidebar(container, onNavigate) {
    const root = document.createElement('div');
    root.className = 'sidebar';

    const brand = document.createElement('div');
    brand.className = 'sidebar__brand';
    brand.textContent = 'Espaze';

    const nav = document.createElement('div');
    nav.className = 'sidebar__nav';

    const buttons = new Map();
    for (const item of NAV_ITEMS) {
        const btn = document.createElement('button');
        btn.className = 'sidebar__item';
        btn.textContent = item.label;
        btn.addEventListener('click', () => onNavigate(item.id));
        buttons.set(item.id, btn);
        nav.appendChild(btn);
    }

    root.append(brand, nav);
    container.appendChild(root);

    return {
        setActive(viewId) {
            for (const [id, btn] of buttons) {
                btn.classList.toggle('active', id === viewId);
            }
        },
    };
}
