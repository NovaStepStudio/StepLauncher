# StepLauncher-Change-40: Centro de Noticias rediseñado (estilo, navegación, rendimiento y responsive)

## Fecha
Agosto 2026

## Release
StepLauncher-2.3.0

## Contexto
Se rediseñó el apartado de Noticias (`frontend/web/src/Modals/NewsModal.vue` + `frontend/web/src/Stores/News.ts`): mejor estilo visual, adaptación a ventanas pequeñas (escritorio, no móviles), navegación coherente (el botón de regresar no volvía al panel principal), carga más rápida desde el renderer (no backend) y contenido mejor alineado (los títulos del markdown tenían margen distinto al texto y los links usaban el color del fondo del modal).

## Qué se hizo

### Navegación (Store `News.ts`)
- `readerBack()` ahora, al estar en la raíz del lector (primer documento), cierra el lector y **vuelve al panel principal de noticias** (antes el botón atrás quedaba deshabilitado y parecía no funcionar). Lo mismo aplica para la flecha `←` del teclado (nuevo `goBack()` en el componente).
- Nuevo helper `reloadDetail(version)`: reintento individual de una card que falló (botón "Reintentar" dentro de la propia card).

### Rendimiento del renderer
- `preloadDetails()` ya no espera cada petición con `await` en serie: **lanza todas las cargas de `news.json` en paralelo** (Go ya corre en goroutines), las cards se rellenan de forma progresiva según llega cada evento.
- La grilla ya no muestra una pantalla gigante de "Buscando novedades…": mientras carga muestra **skeletons con shimmer** (fantasma animado) y cada card pasa a esqueleto hasta que llega su detalle.
- Caché del HTML renderizado por `marked`: `markdownHtml` se cachea por URL de documento (máx. 60 entradas), así volver atrás/adelante en el historial del lector es instantáneo (no se re-parsea el markdown).

### Estilo y coherencia
- Cabecera común en ambas vistas (icono, título, subtítulo dinámico en el lector, botón refrescar con spinner, botón cerrar).
- Barra del lector nueva: botón "Noticias/Volver" prominente con flecha, breadcrumb limpio ("Noticias / … / documento actual") y botón adelante solo cuando hay historial.
- Transición animada entre grilla y lector (`NewsView` con `mode="out-in"`).
- Cards: texto del body **completo sin `line-clamp`** (los bodies son cortos, ≤250 caracteres), footer anclado abajo con `margin-top: auto`, tag "Última" para la versión `latest`, chip de tipo con etiqueta corta.
- Sombra de cards y `pre` ahora usan la variable `--shadow-settings-normal` (se desactiva con "sombras off" del usuario), nunca sombras hardcodeadas.
- **Links del lector**: ya no usan `--background-button-primary` (el color del fondo del modal, invisible por defecto); ahora usan `--color-tag` mezclado con texto y subrayado.
- Tipografía del markdown con ritmo coherente: párrafos y listas con margen inferior uniforme, títulos sin indentación izquierda (el `h2` tenía `padding-left` + `border-left` que desalineaban el texto), `strong` a color de texto, marcadores de lista/`th`/checkbox/blockquote con colores neutros visibles con cualquier personalización.

### Responsive (ventanas pequeñas de escritorio, no móviles)
- Breakpoints a 1280/1120/980/860/680 px + `max-height: 720px`: se compactan paddings de cabecera/toolbar/cuerpo, la grilla reduce columnas (hasta 1 columna bajo 680 px), el botón "Volver" queda solo icono, breadcrumb truncado y la columna de lectura se expande.

## Resultado del build verificable
- `bun run build` en `frontend/` con type-check OK (`vue-tsc --build` sin errores).
- Sin cambios en backend Go (solo se tocó frontend).

## Notas
- Los estilos siguen usando las variables de personalización (`--color-*`, `--control-*`, `--text-*`, `--font-*`, `--background-button-primary` para acentos de botones, `--shadow-settings-normal` para sombras).
- La numeración de cards ("novedades") se mantiene singular/plural correcto.
