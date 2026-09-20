# StepLauncher-Error-5: El panel de Configuración no cargaba (ReferenceError por Temporal Dead Zone en ColorField)

- **Fecha**: 2026-08-04
- **Estado**: corregido
- **Fixed?**: Sí — corregido y mencionado en la release `StepLauncher-2.3.0`.
- **Versión afectada**: 2.3.0 (productVersion de `wails.json`)

---

## 1. Síntoma

Al abrir Configuración, el panel no cargaba (contenido en blanco) y después
la app quedaba rota (no cargaba nada más). El build (`bun run build` con
type-check) pasaba sin errores.

## 2. Causa raíz

En `ColorField.vue`, Change-11 introdujo el guard `lastEmitted` para ignorar
el feedback del v-model durante el arrastre:

```ts
watch(() => props.modelValue, (v) => { ... if (s === lastEmitted) return; ... }, { immediate: true });  // ~línea 289

// ...más abajo...
let lastEmitted = '';  // ~línea 567
```

El watcher con `immediate: true` ejecuta su callback **síncronamente durante
el setup**, antes de que la línea del `let lastEmitted` se ejecute. El binding
existe (hoisting) pero está en **Temporal Dead Zone** → accederlo lanza
`ReferenceError: Cannot access 'lastEmitted' before initialization`. Como
`PersonalizationSettings` monta 7 ColorField a la vez, el setup de cada uno
reventaba, la sección no se renderizaba y el error no controlado rompía el
árbol de Vue (de ahí el "y luego no carga nada").

El type-check no lo detecta: TypeScript no analiza el orden de ejecución
síncrona de los callbacks respecto a la declaración de variables `let`.

## 3. Solución aplicada

Mover la declaración `let lastEmitted = ''` ANTES del watcher (junto a las
demás variables de estado del setup), y quitar la declaración duplicada junto
a `commit()`. Con eso, cuando el callback `immediate` corre durante el setup,
`lastEmitted` ya está inicializado.

## 4. Verificación

- `bun run build` (dentro de `frontend/`) → OK (type-check incluido).
- Abrir Configuración → Personalización: el panel carga con sus 7 selectores
  de color; arrastrar un color guarda en cada movimiento sin tirones y el
  resto de la app sigue funcionando.

## 5. Regla aprendida

En `<script setup>`, cualquier variable `let`/`const` que se lea dentro de un
callback con `immediate` (watchers) o de cualquier código que corra durante el
setup debe declararse ANTES de ese registro. Auditoría: por cada
`watch(..., { immediate: true })`, verificar que todo lo que lee ya está
declarado arriba; los `ReferenceError` por TDZ no los pilla el compilador.