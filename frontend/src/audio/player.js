import {loadMuted, loadVolume} from './storage.js';

/**
 * Streams the PCM chunks emitted by the running core out through Web
 * Audio. Chunks arrive once per emulated frame rather than continuously,
 * so each one is scheduled to start right after the previous one ends
 * instead of immediately, keeping playback gapless.
 */
export function createAudioPlayer() {
    const ctx = new (window.AudioContext || window.webkitAudioContext)();
    const gain = ctx.createGain();
    gain.connect(ctx.destination);
    gain.gain.value = loadMuted() ? 0 : loadVolume();

    let nextStartTime = 0;

    function queueSamples(sampleRate, samples) {
        if (!samples || samples.length === 0) return;

        const buffer = ctx.createBuffer(1, samples.length, sampleRate);
        const channel = buffer.getChannelData(0);
        for (let i = 0; i < samples.length; i++) {
            channel[i] = samples[i] / 32768;
        }

        const source = ctx.createBufferSource();
        source.buffer = buffer;
        source.connect(gain);

        // If playback has fallen behind (e.g. after a pause), resync to
        // now instead of queuing up a growing backlog of stale audio.
        const startAt = Math.max(ctx.currentTime, nextStartTime);
        source.start(startAt);
        nextStartTime = startAt + buffer.duration;
    }

    function setVolume(volume) {
        gain.gain.value = loadMuted() ? 0 : volume;
    }

    function setMuted(muted) {
        gain.gain.value = muted ? 0 : loadVolume();
    }

    function stop() {
        return ctx.close();
    }

    return {queueSamples, setVolume, setMuted, stop};
}
