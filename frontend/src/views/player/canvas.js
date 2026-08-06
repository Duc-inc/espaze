// Turns raw frame payloads ({width, height, pixelsBase64}) coming from the
// Go engine into pixels on screen, scaling the native low-res picture up
// with nearest-neighbor so it stays crisp instead of blurry.
export function createCanvasRenderer(canvas) {
    const ctx = canvas.getContext('2d');
    ctx.imageSmoothingEnabled = false;

    let offscreen = null;
    let offCtx = null;

    function ensureOffscreen(width, height) {
        if (offscreen && offscreen.width === width && offscreen.height === height) {
            return;
        }
        offscreen = document.createElement('canvas');
        offscreen.width = width;
        offscreen.height = height;
        offCtx = offscreen.getContext('2d');
    }

    function decodePixels(pixelsBase64) {
        const binary = atob(pixelsBase64);
        const bytes = new Uint8ClampedArray(binary.length);
        for (let i = 0; i < binary.length; i++) {
            bytes[i] = binary.charCodeAt(i);
        }
        return bytes;
    }

    function drawFrame({width, height, pixelsBase64}) {
        ensureOffscreen(width, height);
        const imageData = new ImageData(decodePixels(pixelsBase64), width, height);
        offCtx.putImageData(imageData, 0, 0);

        ctx.clearRect(0, 0, canvas.width, canvas.height);
        ctx.drawImage(offscreen, 0, 0, width, height, 0, 0, canvas.width, canvas.height);
    }

    return {drawFrame};
}
