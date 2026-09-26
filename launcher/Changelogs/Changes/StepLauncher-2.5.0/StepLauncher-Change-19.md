# Cambios de StepLauncher 2.5.0 (Step-19) — Textos de Configuración más amigables y técnicos

- **Fecha**:   2026-09-25
- **Versión**: 2.5.0
- **Estado**:  implementado
- **Release**: StepLauncher-2.5.0 — en esta release se menciona que fue añadido.

## Qué cambió

Reescritura de todos los textos visibles de Configuración (etiquetas, descripciones, tips, placeholders, botones y mensajes): tono amigable pero con precisión técnica (nombres reales de conceptos: −Xmx, hash, JVM, Yggdrasil, SMTC, SOCKS5…), rangos y unidades donde ayudan (50%–200%, 1–8, 3–300 s, 20–100) y consecuencias de cada opción. Solo strings: sin cambios de lógica, valores, bindings ni estilos.

### 1. General (`Settings/Sections/General.vue`)

- "Tamaño de todo" → "Escala de la interfaz" (menciona rango y atajo Ctrl +/−); Rich Presence con nombre técnico; "Jugar automáticamente al terminar de instalar" con caso de uso; efectos con costo honesto (blur usa GPU); "En tu ausencia" con aviso de que las descargas siguen; "Zona de peligro" con alcance explícito (mundos e instancias a salvo); actualizaciones mencionando GitHub.

### 2. Minecraft (`Settings/Sections/Minecraft.vue`)

- GPU: "Ajustes de GPU personalizados", "Aceleración por hardware", "GPU preferida" (dedicada/integrada con ejemplos NVIDIA/AMD, Intel/APU), "Perfil de la GPU" y tip de crash/pantalla negra.
- RAM con flag −Xmx y regla de la mitad; Java con modos renombrados ("Instalado en mi PC", "El oficial de Minecraft"); ventana con F11; JVM/juego como "Argumentos" con advertencia real; verificación por hash; offline/compatibilidad/logs con lenguaje técnico claro. Grupo "Más opciones" → "Comportamiento".

### 3. Personalización (`Settings/Sections/Personalization.vue`)

- Fondo con requisitos y porqués (1920×1080, 20 MB/1080p, PCs modestos); música de fondo aclarando que no suena dentro del juego; fuentes como "Fuente de títulos/textos" con defaults; colores con ejemplos concretos (Fabric, release, poco espacio); "Probar en vivo". Se suavizó la mención a rutas internas (carpeta interna de audio).

### 4. Música (`Settings/Sections/MusicPanel.vue`)

- Escaneo incremental explicado (primera vez tarda, después solo lo nuevo); carátula gigante sin jerga de unidades (1rem); ambiente/SMTC en criollo técnico; stats ("tamaño del índice", "carátulas guardadas"); mensajes de éxito/error completos (qué pasó y dónde).

### 5. Red, Descargas, Integridad, Almacenamiento, Cuentas

- Red: "Proxy solo para Minecraft", "Servidor proxy", "Credenciales del proxy" y tip de "malformed HTTP status" orientado a solución.
- Descargas: "Descargas en paralelo" (1–8, redes flojas) y "Tope de velocidad" (Mbps, 0 = todo el ancho de banda).
- Integridad: fases con vocabulario real (inventario, hashes), alcance "Juego base/Instancias" (antes "Juego/Mundos", impreciso) y distinción prevenir vs. curar.
- Almacenamiento: modos con explicación inline ("Normal (recomendada)", "Portable (USB)"), aviso de respaldo antes de migrar, "Carpeta «game» separada" y mensajes de error accionables (permisos, ruta).
- Cuentas: renovación de "tokens Yggdrasil caducados" al abrir.

## Por qué

A pedido del owner: que cualquier persona entienda qué hace cada ajuste, sin perder precisión técnica.

## API afectada

Sin cambios en el backend Go ni bindings. Solo strings del frontend.

## Comportamiento anterior/nuevo

- Antes: etiquetas cortas y vagas ("Qué gráfica usar", "Elige uno", "Limpiar", "Borrar").
- Ahora: cada opción dice qué es, qué valores admite y qué consecuencia tiene.

## Cómo verificar

- Recorrer Ajustes sección por sección y leer etiquetas, descripciones, tips y mensajes.
- `bun run build` en `launcher/frontend` y `go build ./...` en `launcher/` — pendientes: el owner pidió no compilar en esta tanda.
