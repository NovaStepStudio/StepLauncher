# Cambios de StepLauncher 2.5.0 (Step-20) — Cambiar de sección en Configuración vuelve arriba del todo

- **Fecha**:   2026-09-25
- **Versión**: 2.5.0
- **Estado**:  implementado
- **Release**: StepLauncher-2.5.0 — en esta release se menciona que fue añadido.

## Qué cambió

- **`Settings/Settings.vue`**: el contenedor con scroll (`.SettingsModal_Content`) vuelve arriba (`scrollTo({ top: 0 })` tras `nextTick`) cada vez que se cambia de sección en la barra lateral y al abrir el modal. Antes quedabas a la altura donde habías dejado la sección anterior.

## Por qué

A pedido del owner: al cambiar de sección en la config, el scroll quedaba a mitad de camino y parecía que la sección estaba vacía o cortada.

## API afectada

Sin cambios en el backend Go ni bindings. Solo frontend (una ref + un watcher en el modal de ajustes).

## Comportamiento anterior/nuevo

- Antes: el scroll se conservaba entre secciones.
- Ahora: cada sección abre desde arriba.

## Cómo verificar

- Abrir Configuración, bajar en una sección larga (ej. Personalización), cambiar a otra sección y comprobar que arranca desde arriba.
- `bun run build` en `launcher/frontend` y `go build ./...` en `launcher/` — pendientes: el owner pidió no compilar en esta tanda.
