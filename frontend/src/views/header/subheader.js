import './subheader.css';

const NAV_ITEMS = [
    {id: 'library', label: 'BIBLIOTHÈQUE'},
];

/**
 * Renders the dense nav strip: history chevrons, then section tabs, then
 * settings on the far right.
 * @param {HTMLElement} container
 * @param {(viewId:string)=>void} onNavigate
 * @param {()=>void} onSettings
 */
export function mountSubheader(container, onNavigate, onSettings) {
    const el = document.createElement('div');
    el.className = 'subheader';

    const history = document.createElement('div');
    history.className = 'subheader__history';
    for (const icon of ['fa-chevron-left', 'fa-chevron-right']) {
        const btn = document.createElement('button');
        btn.innerHTML = `<i class="fa-solid ${icon}"></i>`;
        btn.disabled = true;
        history.appendChild(btn);
    }

    const nav = document.createElement('nav');
    nav.className = 'subheader__nav';
    const buttons = new Map();
    for (const item of NAV_ITEMS) {
        const btn = document.createElement('button');
        btn.className = 'subheader__item';
        btn.textContent = item.label;
        btn.addEventListener('click', () => onNavigate(item.id));
        buttons.set(item.id, btn);
        nav.appendChild(btn);
    }

    const right = document.createElement('div');
    right.className = 'subheader__right';
    const settingsBtn = document.createElement('button');
    settingsBtn.className = 'subheader__icon-btn';
    settingsBtn.title = 'Paramètres';
    settingsBtn.innerHTML = '<i class="fa-solid fa-gear"></i>';
    settingsBtn.addEventListener('click', onSettings);
    right.appendChild(settingsBtn);

    el.append(history, nav, right);
    container.appendChild(el);

    return {
        setActive(viewId) {
            for (const [id, btn] of buttons) {
                btn.classList.toggle('active', id === viewId);
            }
        },
    };
}
