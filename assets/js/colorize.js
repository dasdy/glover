
function interpolateColor(value) {
    let r, g, b;

    if (value <= 0.5) {
        // Interpolate between blue (0, 0, 255) and green (0, 255, 0)
        const ratio = value / 0.5;
        r = 0;
        g = Math.round(255 * ratio);
        b = Math.round(255 * (1 - ratio));
    } else {
        // Interpolate between green (0, 255, 0) and red (255, 0, 0)
        const ratio = (value - 0.5) / 0.5;
        r = Math.round(255 * ratio);
        g = Math.round(255 * (1 - ratio));
        b = 0;
    }

    let blendFactor = 0.5
    // Blend with white (255, 255, 255) to reduce intensity
    r = Math.round(r + (255 - r) * blendFactor);
    g = Math.round(g + (255 - g) * blendFactor);
    b = Math.round(b + (255 - b) * blendFactor);

    return `rgba(${r}, ${g}, ${b})`;
}
