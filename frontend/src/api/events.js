// Subscribes to the Wails runtime events emitted by internal/app/events
// (see FrameEvent/AudioEvent) and exposes a small callback API so views
// don't need to know about the underlying runtime bus.
import {EventsOn} from '../../wailsjs/runtime/runtime';

const FRAME_EVENT = 'emulation:frame';
const AUDIO_EVENT = 'emulation:audio';

export function onFrame(callback) {
    return EventsOn(FRAME_EVENT, callback);
}

export function onAudio(callback) {
    return EventsOn(AUDIO_EVENT, callback);
}
