# vendor/ — librerías del mapa de nodos, servidas por el propio explorador

Nada se carga desde CDNs ni terceros: quien abre el explorador solo habla con el explorador.

| Archivo | Origen | Versión | Licencia |
|---|---|---|---|
| `d3-7.9.0.min.js` | https://github.com/d3/d3 | 7.9.0 | ISC (`LICENSE-d3`) |
| `topojson-client-3.1.0.min.js` | https://github.com/topojson/topojson-client | 3.1.0 | ISC (`LICENSE-topojson-client`) |
| `countries-110m.json` | https://github.com/topojson/world-atlas (datos: Natural Earth, dominio público) | 2.0.2 | ISC (`LICENSE-world-atlas`) |

Bajados con `npm pack` el 14-oct-2026, sin modificar. Para actualizar: `npm pack d3@X topojson-client@Y world-atlas@Z`, copiar y anotar aquí.
