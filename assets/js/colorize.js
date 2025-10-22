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
let chatgpt1 = [
  hexToRgb("#f0f9ff"),
  hexToRgb("#bae6fd"),
  hexToRgb("#7dd3fc"),
  hexToRgb("#38bdf8"),
  hexToRgb("#0ea5e9"),
  hexToRgb("#1e3a8a"),
];

let chatgpt2 = [
  hexToRgb("#f0f9ff"),
  hexToRgb("#dbeafe"),
  hexToRgb("#a5b4fc"),
  hexToRgb("#818cf8"),
  hexToRgb("#4f46e5"),
  hexToRgb("#312e81"),
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
    // schema = defaultGradient;
    // schema = viridis.slice().reverse();
    // schema = magma.slice().reverse();
    // schema = inferno.slice().reverse();
    schema = chatgpt1.slice().reverse();
    // schema = chatgpt2.slice().reverse();
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
  // Blend with white (255, 255, 255) to reduce intensity
  r = Math.round(r + (255 - r) * blendFactor);
  g = Math.round(g + (255 - g) * blendFactor);
  b = Math.round(b + (255 - b) * blendFactor);

  return `rgba(${r}, ${g}, ${b})`;
}

function colorize(maxVal) {
  // Apply the gradient-based color to all SVG key elements
  document.querySelectorAll(".key-rect").forEach(function (rect) {
    let presses = rect.getAttribute("data-presses");
    let value = parseFloat(presses);

    if (!isNaN(value)) {
      const color = interpolateColor(value / maxVal); // Calculate the color based on the value
      rect.setAttribute("fill", color); // Set the fill color for SVG rect
    }
  });
}

function addConnectionPath(fromId, toId, strength) {
  const fromBox = document.getElementById(`key-box-${fromId}`);
  const toBox = document.getElementById(`key-box-${toId}`);
  const svgBox = document.getElementById(`keysgrid`);

  const pathsGroup = document.querySelector(".connection-paths");

  const path = document.createElementNS("http://www.w3.org/2000/svg", "path");
  path.setAttribute("fill", "none");
  path.setAttribute("stroke", "#6366f1");
  path.setAttribute("stroke-width", strength || "2");
  path.setAttribute("stroke-opacity", "0.7");
  path.setAttribute("stroke-linecap", "round");

  // Get the transformed center points
  const fromBounds = fromBox.getBBox();
  const toBounds = toBox.getBBox();

  const point1 = keyCenter(fromBox, svgBox);
  const point2 = keyCenter(toBox, svgBox);

  // Create curved path
  const midX = (point1.x + point2.x) / 2;
  const midY = (point1.y + point2.y) / 2 - 40; // Curve control point
  path.setAttribute(
    "d",
    `M ${point1.x} ${point1.y} Q ${midX} ${midY} ${point2.x} ${point2.y}`,
  );

  pathsGroup.appendChild(path);
}

function addCircleAtElem(boxId) {
  const box = document.getElementById(`key-box-${boxId}`);
  const svgBox = document.getElementById(`keysgrid`);
  const pathsGroup = document.querySelector(".connection-paths");

  const circle = document.createElementNS(
    "http://www.w3.org/2000/svg",
    "circle",
  );

  const coords = keyCenter(box, svgBox);

  circle.setAttribute("cx", coords.x);
  circle.setAttribute("cy", coords.y);
  circle.setAttribute("r", "5");
  circle.setAttribute("fill", "red");
  circle.setAttribute("stroke", "black");
  circle.setAttribute("stroke-width", "2");
  pathsGroup.appendChild(circle);
}

function keyCenter(box, svgBox) {
  const fromBounds = box.getBBox();
  const point1 = box.getBoundingClientRect();
  const realViewBox = svgBox.getBoundingClientRect();
  const viewBox = svgBox.viewBox.baseVal;

  const xScale = viewBox.width / realViewBox.width;
  const yScale = viewBox.height / realViewBox.height;

  const x = (point1.x - realViewBox.x + point1.width / 2) * xScale;
  const y = (point1.y - realViewBox.y + point1.height / 2) * yScale;

  return { x: x, y: y };
}
