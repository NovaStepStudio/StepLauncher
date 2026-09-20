# Guía para Agentes de IA - StepLauncher Website

Sitio web público e informativo de StepLauncher (Vue 3 + TS + Vite, en `web/src`). Toda la comunicación y comentarios de código deben redactarse en **español**.

## 1. Área de trabajo (obligatorio)

- El área de trabajo es `website/`. No salir de ella: no leer, modificar ni crear archivos fuera (repo padre, `Changelogs/`, backend Go, etc.).
- Si una tarea parece requerir cambios fuera del área, avisar y pedir permiso explícito antes de tocar nada.
- No borrar archivos sin permiso explícito del usuario.

## 2. Estilo y hojas de estilo (SCSS obligatorio)

- La web debe mantener el estilo visual que ya tiene: reutilizar variables (`var(--color-*)`), componentes de `Common/` y patrones existentes.
- No inventar estilos ajenos al sistema (colores, fuentes o layouts que no correspondan a la identidad de la web).
- Toda la interfaz se escribe en **SASS/SCSS** (`<style scoped lang="scss">` en cada componente). Prohibido CSS plano suelto y estilos inline complejos.
- **Coherencia obligatoria**: hay una sola forma de escribir SCSS en el proyecto. Antes de crear o tocar una hoja, revisar cómo están escritas las existentes y seguir ese mismo patrón (anidado, nombres de clases, uso de variables, orden de propiedades). No cambiar de convención en cada archivo.
- El SCSS vive junto a su dominio/componente y se importa con `@use`.
- Imports TS con alias `@/...`, nunca rutas relativas largas.

## 3. Secrets y tokens

- Todo secret o token va en `.env` (desarrollo/localhost) y en `wrangler.jsonc` (variables de producción), usando la **misma variable** en ambos entornos.
- El código debe funcionar tanto en localhost como en producción leyendo esas variables (con fallback local solo si no expone nada sensible).
- Nunca hardcodear secrets, tokens ni sitekeys en el código fuente.

## 4. Código limpio (no exponer internals)

- No exponer al usuario endpoints internos (p. ej. `GET /me/@user`), tokens, URLs de API ni detalles de sesión en la UI, logs o mensajes de error.
- Los errores visibles deben ser mensajes genéricos en español; el detalle técnico queda en consola solo en desarrollo.

## 5. Render 3D de Minecraft

- Skins, capas y demás elementos de Minecraft se renderizan en **3D** utilizando una librería **ya existente** del ecosistema NPM.
- No reinventar un renderer propio ni usar imágenes planas 2D como sustituto.

## 6. Gestor de paquetes: Bun siempre

- Usar siempre **Bun** (`bun install`, `bun run build`, etc.). **Prohibido npm**: es más lento y supone un riesgo de seguridad (malware distribuido en paquetes que npm no filtra; Bun valida los paquetes al descargar).
- Verificación: `bun run build` en `website/` (type-check + build de Vite) antes de dar una tarea por terminada.

## 7. Responsive: móvil y tablet siempre que se pueda

- Todo lo nuevo o modificado debe funcionar en **móvil y tablet**, no solo en escritorio: layouts fluidos, sin scroll horizontal, textos e imágenes que no se rompan en pantallas angostas.
- Usar los breakpoints que ya usa el proyecto (revisarlos antes de inventar otros) y mantener el mismo criterio en cada componente.
- Pensar en táctil: botones y enlaces con tamaño suficiente, menús usables sin hover.
- Si algo realmente no puede ser responsive, avisar y justificarlo antes de entregarlo solo para escritorio.
