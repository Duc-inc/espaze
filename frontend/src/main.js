import '@fortawesome/fontawesome-free/css/all.min.css';
import './style/theme.css';
import './style/base.css';
import {createApp} from './app/app.js';
import {WindowFullscreen, WindowIsFullscreen, WindowUnfullscreen} from '../wailsjs/runtime/runtime';

createApp(document.getElementById('app'));

window.addEventListener('keydown', async (e) => {
    if (e.key !== 'F11') {
        return;
    }
    e.preventDefault();
    if (await WindowIsFullscreen()) {
        WindowUnfullscreen();
    } else {
        WindowFullscreen();
    }
});
