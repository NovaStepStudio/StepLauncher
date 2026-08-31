# Bugs de StepLauncher 2.4.1 (Bug-2) — El widget de descargas se rompe con textos extremadamente largos

- **Fecha**: 2026-08-16
- **Versión**: 2.4.1 (en desarrollo)
- **Estado**: corregido y verificado
- **Release**: en desarrollo — aún no mencionado en ninguna release.

## El bug en cuestión

El widget de descargas (`frontend/web/src/Downloads/Widget.vue`) muestra en la línea secundaria el nombre de la versión y la fase en curso (descargando, verificando, instalando…). Cuando ese texto era extremadamente largo (p. ej. una versión de un servidor de terceros con un slug muy extenso), la línea no estaba protegida contra desbordamiento: la barra de progreso y el resto de la tarjeta se empujaban fuera de su contenedor, rompiendo la maquetación del widget completo.

## Qué afectaba y qué hacía

- La línea secundaria (`title`/`sub` generados en `Widget.vue`) sin `max-width` ni `overflow`, y el contenedor `.DownloadWidget` sin `overflow: hidden`, hacían que el texto largo estirara la tarjeta horizontalmente (o se solapara con la barra de progreso y el botón de cancelar).
- El widget de descargas activas se volvía ilegible y deforme con nombres de versión largos.

## Solución final

En `frontend/web/src/Downloads/Styles/Widget.scss`:

- `.DownloadWidget_Sub`: ahora tiene `max-width: 100%`, `white-space: nowrap`, `overflow: hidden` y `text-overflow: ellipsis`, de modo que el texto largo se trunca con "…" en una sola línea en lugar de desbordar.
- `.DownloadWidget`: ahora tiene `overflow: hidden` como defensa final, para que ningún hijo pueda escapar del contenedor.

## Verificación

- `bun run build` (frontend): type-check + build de producción correctos.
- Revisión visual: el widget mantiene su maquetación con versiones de nombre muy largo (el texto se corta con puntos suspensivos).