package server

import "html/template"

var tpl *template.Template = template.Must(template.New("something").Parse(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta http-equiv="refresh" content="5">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Grid Layout</title>
    <style>
        .grid-container {
            display: grid;
            grid-template-columns: repeat(12, 1fr);
            gap: 10px;
        }

        .grid-item {
            background-color: #ddd;
            border: 1px solid #ccc;
            padding: 20px;
            text-align: center;
        }

        .hidden {
            visibility: hidden
        }
    </style>
</head>
<body>
    <div class="grid-container">
        {{- range .Items }}
            {{ if .Visible }}
                <div class="grid-item">{{ .Label }}</div>
            {{ else }}
                <div class="hidden"></div>
            {{ end }}
        {{- end }}
    </div>
    <script>
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

            return ` + "`rgba(${r}, ${g}, ${b}, 30)`" + `;
        }

        // Apply the gradient-based color to all .grid-item elements
        document.querySelectorAll('.grid-item').forEach(function(item) {
            let value = parseFloat(item.textContent.trim()); // Get the number from 0 to 1
            if (!isNaN(value)) {
                const color = interpolateColor(value/ {{.MaxVal}});  // Calculate the color based on the value
                item.style.backgroundColor = color;     // Set the background color
            }
        });
    </script>
</body>
`))
