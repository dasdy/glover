function hexToRgb(hex) {
  var result = /^#?([a-f\d]{2})([a-f\d]{2})([a-f\d]{2})$/i.exec(hex);
  return result
    ? {
        r: parseInt(result[1], 16),
        g: parseInt(result[2], 16),
        b: parseInt(result[3], 16),
      }
    : null;
}

let viridis = [
  hexToRgb("#fde725"),
  hexToRgb("#7ad151"),
  hexToRgb("#22a884"),
  hexToRgb("#2a788e"),
  hexToRgb("#414487"),
  hexToRgb("#440154"),
];
let inferno = [
  hexToRgb("#fcffa4"),
  hexToRgb("#fca50a"),
  hexToRgb("#dd513a"),
  hexToRgb("#932667"),
  hexToRgb("#420a68"),
  hexToRgb("#000004"),
];
let magma = [
  hexToRgb("#fcfdbf"),
  hexToRgb("#fe9f6d"),
  hexToRgb("#de4968"),
  hexToRgb("#8c2981"),
  hexToRgb("#3b0f70"),
  hexToRgb("#000004"),
];

let defaultGradient = [
  hexToRgb("#0000ff"),
  hexToRgb("#00ff00"),
  hexToRgb("#ff0000"),
];

function normalize(n, min, max) {
  if (min === max) {
    return 0;
  }
  return (n - min) / (max - min);
}
function expand(n, min, max) {
  return (max - min) * n + min;
}
function translate(n, min1, max1, min2, max2) {
  return expand(normalize(n, min1, max1), min2, max2);
}

function interpolateColor(value, schema) {
  if (schema === undefined) {
    schema = defaultGradient;
    // schema = viridis.slice().reverse();
    // schema = magma.slice().reverse();
    // schema = inferno.slice().reverse();
  }

  let r, g, b;
  if (value >= 1) {
    let c = schema[schema.length - 1];
    r = c.r;
    g = c.g;
    b = c.b;
  } else if (value <= 0) {
    let c = schema[0];
    r = c.r;
    g = c.g;
    b = c.b;
  } else {
    const N = schema.length - 1, // number of buckets
      s = 1 / N, // size of each bucket
      n = Math.floor(value / s), // bucket number
      minVal = s * n, // bucket min value
      maxVal = s * (n + 1), // bucket max value
      A = schema[n], // bucket min color
      B = schema[n + 1]; //bucket max color

    if (B === undefined) {
      console.log("heyo");
    }

    r = translate(value, minVal, maxVal, A.r, B.r);
    g = translate(value, minVal, maxVal, A.g, B.g);
    b = translate(value, minVal, maxVal, A.b, B.b);
  }

  let blendFactor = 0.5;
  // // Blend with white (255, 255, 255) to reduce intensity
  r = Math.round(r + (255 - r) * blendFactor);
  g = Math.round(g + (255 - g) * blendFactor);
  b = Math.round(b + (255 - b) * blendFactor);

  return `rgba(${r}, ${g}, ${b})`;
}

function colorize(maxVal) {
  // Apply the gradient-based color to all .grid-item elements
  document.querySelectorAll(".keys-pressed").forEach(function (item) {
    let value = parseFloat(item.textContent.trim()); // Get the number from 0 to 1
    if (!isNaN(value)) {
      const color = interpolateColor(value / maxVal); // Calculate the color based on the value
      let parentItem = item.parentNode;
      parentItem.style.backgroundColor = color; // Set the background color
    }
  });
}
