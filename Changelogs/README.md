# Changelogs

Historial oficial y **auditoría interna** del launcher **StepLauncher**. Esta carpeta NO es contenido generado "porque sí": cada entrada documenta un **error** (bug, incidente, deadlock) o un **cambio** (funcionalidad, mejora, refactor) que **le pasó de verdad al owner del repositorio durante el desarrollo** — por qué se hizo, qué se rompió, cómo se arregló y en qué release salió. Es la trazabilidad completa del proyecto: quien trabaje aquí después —desarrollador o agente de IA— debe consultarla ANTES de diagnosticar o modificar cualquier cosa.

## Los changelogs representan hechos reales

Las entradas de `Errors/`, `Changes/` y `Releases/` documentan cosas que **realmente ocurrieron** en el desarrollo del proyecto. No se generan entradas ficticias, inventadas o especulativas: un error se documenta cuando ocurrió, un cambio cuando se implementó, una release cuando se publicó. El historial no es un registro aspiracional.

Un agente de IA puede ayudar a redactar, investigar o estructurar una entrada, pero la información debe provenir siempre de una fuente real:

- código;
- logs y trazas;
- cambios efectivamente realizados;
- comportamiento reproducible;
- documentación existente;
- evidencia del repositorio;
- o información proporcionada por el desarrollador.

Si durante la investigación de una tarea aparece un error o un cambio nuevo, se documenta con la misma evidencia; lo que no puede documentarse son "casos hipotéticos" o problemas que solo la IA especuló.

## Esta carpeta también es la memoria de los agentes

StepLauncher es un proyecto grande y sus componentes están fuertemente conectados. Un agente de IA (o un desarrollador nuevo) que entre al repositorio después no tiene el contexto del owner: `Changelogs/` es lo que le permite reconstruirlo y trabajar sin empezar de cero.

Consultando `Changelogs/` se puede descubrir:

- bugs que ya ocurrieron;
- causas raíz conocidas;
- soluciones aplicadas;
- decisiones arquitectónicas;
- reglas aprendidas;
- cambios relevantes;
- posibles regresiones;
- y el comportamiento histórico de sistemas complejos.

Esto permite que una IA posterior **no vuelva a investigar desde cero problemas que ya fueron diagnosticados**: si una causa raíz y su solución ya están documentadas, se usan como punto de partida, no se re-descubren.

## Roles de la documentación

| Documento   | Rol                                                                                                        |
|-------------|------------------------------------------------------------------------------------------------------------|
| `AGENTS.md` | **Reglas actuales**: lo que los agentes y desarrolladores deben cumplir hoy al trabajar sobre el proyecto. |
| `Changelogs/` | **Memoria histórica**: lo que ocurrió en el desarrollo (errores, cambios, releases).                     |
| `Errors/`   | Auditoría técnica de errores reales. Puede contener: síntoma, reproducción, causa raíz, evidencia, flujo afectado, solución, verificación y regla aprendida. |
| `Changes/`  | Registro técnico de modificaciones reales. Puede contener: qué cambió, archivos/componentes afectados, motivo, APIs afectadas, comportamiento anterior/nuevo y cómo verificarlo. |
| `Releases/` | Historial orientado a la versión publicada y al usuario; no es documentación interna de build.             |

La relación conceptual es la siguiente:

- `AGENTS.md` = **reglas actuales**;
- `Changelogs/` = **memoria histórica**.

Un agente debe utilizar ambos: las reglas para saber qué hacer hoy y el historial para no repetir errores del pasado.

## Estructura de carpetas

| Carpeta       | Propósito                                                                                     |
|---------------|-----------------------------------------------------------------------------------------------|
| `Errors/`     | Bugs e incidentes reales que pasaron durante el desarrollo, en una **subcarpeta por versión**. Cada entrada incluye su `Fixed?`. |
| `Changes/`    | Funcionalidades, mejoras y modificaciones reales implementadas, en una **subcarpeta por versión**. Cada entrada indica su `Release`. |
| `Releases/`   | Una carpeta por versión publicada (`StepLauncher-X.Y.Z/`) con su changelog completo + noticia (JSON). |
| `index.json`  | **Punto de entrada** de todas las entradas (uno por carpeta: `Errors/index.json`, `Changes/index.json`, `Releases/index.json`): lista la versión más reciente y las rutas relativas a cada archivo. Fuente para el centro de noticias del launcher. |
| `README.md`   | Este archivo: estructura, convenciones, plantillas y el flujo para armar una release.          |

Tanto errores como cambios y releases se guardan **por versión**:

| Tipo    | Formato                                            | Ejemplo                                                             |
|---------|----------------------------------------------------|---------------------------------------------------------------------|
| Error   | `Errors/StepLauncher-X.Y.Z/StepLauncher-Error-N.md`         | `Errors/StepLauncher-2.3.0/StepLauncher-Error-1.md`        |
| Cambio  | `Changes/StepLauncher-X.Y.Z/StepLauncher-Change-N.md`       | `Changes/StepLauncher-2.3.0/StepLauncher-Change-1.md`      |
| Release | `Releases/StepLauncher-X.Y.Z/` (carpeta)                    | `Releases/StepLauncher-2.3.0/`                              |

- **`N` es secuencial por tipo y por versión**: cada versión nueva crea su carpeta `StepLauncher-X.Y.Z/` y la numeración empieza **de nuevo desde 1**. Dentro de una misma versión `N` no se reutiliza ni se re-numera; el número más alto de esa carpeta es la entrada más reciente de esa versión.
- `X.Y.Z` corresponde al `productVersion` de `wails.json`.
- Nombres de carpetas y archivos en inglés; **contenido en español**.

## Ciclo de vida de un Error

Un error documentado en `Errors/` pasa por estos estados (indicarlo en la entrada):

1. **Encontrado** — el problema existe y está identificado.
2. **En corrección** — hay un arreglo en curso.
3. **Corregido** — verificado y cerrado, con la solución y la regla aprendida.

Además, cada error tiene el campo **`Fixed?`**, que responde si ya se solucionó **y en qué release se menciona**:

- `- **Fixed?**: Sí — corregido y mencionado en la release StepLauncher-2.3.0.`
- `- **Fixed?**: No — sigue pendiente, aún no aparece en ninguna release.`

Siempre se documenta el error **aunque el arreglo siga en curso**; el historial advierte a quien trabaje después de lo que no debe repetirse.

## Ciclo de vida de un Cambio

Cada cambio registrado en `Changes/` indica la **`Release`** donde se menciona que fue añadido por primera vez:

- Formato clásico: `- **Release**: StepLauncher-2.3.0 — en este release se menciona que fue añadido.`
- Formato con sección (archivos que no declaran campos en la cabecera):

  ```markdown
  ## Release
  StepLauncher-2.3.0 — se mencionó por primera vez en esta release.
  ```

Si el cambio todavía no está publicado: `StepLauncher-2.3.0 (en desarrollo)`.

## Una tarea, un solo MD (consolidación de entradas)

Cada `StepLauncher-Error-N.md` o `StepLauncher-Change-N.md` representa **una tarea completa**, no un archivo tocado ni un subcambio aislado.

Si una misma petición o sesión de trabajo implica **más de un cambio** (varias funcionalidades, varios módulos, varios archivos) o **más de un error relacionado** (misma investigación, misma causa raíz, misma tanda de trabajo), se documenta todo en **UN SOLO MD** con subsecciones (p. ej. `### 1.`, `### 2.`...). No se genera un MD por cada cambio o error suelto: eso fragmenta el historial y dispara la numeración sin aportar trazabilidad.

Criterios:

- **Tarea = 1 entrada**: una petición que toca varios sitios (backend, frontend, config, bindings...) agrupa todo en un `StepLauncher-Change-N.md` que enumera los bloques.
- **Causa raíz = 1 error**: varios síntomas del mismo origen comparten el `StepLauncher-Error-N.md`. Solo se abre un error nuevo cuando la causa raíz es independiente (otro módulo, otro comportamiento, otra sesión).
- **Ampliar antes que duplicar**: si la tarea ya tiene una entrada en la versión en curso, se extiende esa entrada; no se crea una nueva para lo mismo.
- **Numeración baja = historial legible**: un MD con varias secciones es preferible a varias entradas de una sola línea.

La regla práctica: **cuando una petición tenga más de un cambio o error, se juntan en un único MD**.

---

## Cómo se arma una release (flujo obligatorio)

Un changelog de release **no es un solo archivo suelto**: cada versión vive en sus propias carpetas (errores, cambios y release). Así se hace, en orden:

### Paso 1 — Crear las carpetas de la versión

Al empezar a trabajar sobre una versión nueva (según el `productVersion` de `wails.json`), crear:

- `Changelogs/Errors/StepLauncher-X.Y.Z/` — donde se registran los errores de ESTA versión (numeración desde 1).
- `Changelogs/Changes/StepLauncher-X.Y.Z/` — donde se registran los cambios de ESTA versión (numeración desde 1).
- `Changelogs/Releases/StepLauncher-X.Y.Z/` — donde se empaqueta la release al publicar (changelog + noticia).

Ejemplo para la 2.3.0: `Errors/StepLauncher-2.3.0/`, `Changes/StepLauncher-2.3.0/`, `Releases/StepLauncher-2.3.0/`.

Además, tener en cuenta cómo se actualizan los `index.json` (ver Paso 4): solo se vuelven a generar **al publicar la release**, no al crear cada error o cambio.

### Paso 2 — Changelog completo (MD)

Dentro de la carpeta, crear `StepLauncher-Release-X.Y.Z.md` (p. ej. `StepLauncher-Release-2.3.0.md`): el registro completo de **todo lo que trajo la actualización** (funcionalidades nuevas, mejoras, correcciones), redactado para que un usuario entienda qué pasó y qué se arregló.

Reglas de este MD:

- Debe mencionar los **errores que se solucionaron en esta versión** como enlaces directos a su MD de auditoría:

  ```markdown
  - [StepLauncher-Error-8: Crash al cancelar una descarga](../../Errors/StepLauncher-2.3.0/StepLauncher-Error-8.md) fue corregido en esta versión.
  - [StepLauncher-Error-9: Stack overflow al lanzar Minecraft](../../Errors/StepLauncher-2.3.0/StepLauncher-Error-9.md) fue corregido en esta versión.
  ```

- **NO incluir instrucciones de compilación ni de build** (nada de "compilar con wails", "bun run build", etc.): eso pertenece al centro de noticias, no al changelog. Este MD se centra en el contenido de la actualización para el usuario.
- Si hay notas de instalación para el usuario (p. ej. actualizador automático), se pueden incluir brevemente.

### Paso 3 — Noticia para el centro de noticias (JSON)

Dentro de la misma carpeta, crear `news.json`: la entrada que verán los usuarios en el **apartado de noticias** del launcher. Esto NO es solo "registrar cambios o errores y mencionar la release": es la noticia pública de la versión.

| Campo        | Tipo     | Regla                                                                                      |
|--------------|----------|--------------------------------------------------------------------------------------------|
| `title`      | string   | Título de la versión indicando el tipo de contenido (cambios, mejoras, bug fixes, actualización importante, cambios internos). |
| `body`       | string   | Pequeño resumen de la actualización. **NO más de 512 caracteres.**                         |
| `type`       | string   | Tipo de noticia. **Seguir el patrón fijo — NO colocar cualquier cosa** (ver abajo).        |
| `date`       | string   | Fecha de creación de la noticia, formato `YYYY-MM-DD`.                                     |
| `changelog`  | string   | Ubicación del MD completo de todos los cambios (ruta relativa del repo, o URL si el centro de noticias la sirve). |

**Patrón de `type` (único permitido):**

| Valor         | Significado                                            |
|---------------|--------------------------------------------------------|
| `changes`     | Cambios generales en el launcher.                      |
| `improvements`| Mejoras de funcionalidad o experiencia.                |
| `bugfix`      | Corrección de errores.                                 |
| `major`       | Actualización importante (grandes novedades).          |
| `internal`    | Cambios internos (backend, builds, refactors).         |

Ejemplo real:

```json
{
  "title": "StepLauncher 2.3.0 — Actualización importante",
  "type": "major",
  "body": "La actualización más grande hasta la fecha: sistema de cuentas (offline y AuthLib), presencia en Discord, actualizador automático, instalador de versiones y modloaders renovado y decenas de errores corregidos.",
  "date": "2026-08-07",
  "changelog": "Changelogs/Releases/StepLauncher-2.3.0/StepLauncher-Release-2.3.0.md"
}
```

### Paso 4 — Punto de entrada: los `index.json`

Cada carpeta tiene su `index.json` que sirve como **punto de entrada de TODAS sus entradas** — así el launcher puede montar el centro de noticias sin escanear carpetas. **Apuntan SIEMPRE al archivo** (nunca a la carpeta ni a un JSON intermedio):

- `Releases/index.json` → apunta a cada `news.json` (la noticia de cada release).
- `Errors/index.json` → apunta a cada `StepLauncher-Error-N.md` (el archivo, sin resumir su contenido).
- `Changes/index.json` → apunta a cada `StepLauncher-Change-N.md` (el archivo, sin resumir su contenido).

> **⚠️ Los `index.json` se regeneran SOLO al publicar una release** — NO al crear cada error o cambio (esos MD se crean y listo, sin tocar índices). Cuando toque crear el MD de release (`StepLauncher-Release-X.Y.Z.md`), ejecutar el generador `generate_indexes.ps1` (en la raíz de `Changelogs/`): escanea las carpetas de versión ya escritas, agrupa por versión (la más nueva primero), aplica rutas relativas y formatea el JSON a 4 espacios. Ejecutar: `powershell -NoProfile -ExecutionPolicy Bypass -File Changelogs\generate_indexes.ps1`. No toca los MD: solo lee los nombres existentes. Regenerarlo además siempre que se cree una carpeta de versión nueva.

Estructura de `Releases/index.json`:

```json
{
  "latest": "2.3.0",
  "content": [
    {
      "version": "2.3.0",
      "path": "./StepLauncher-2.3.0/news.json"
    }
  ]
}
```

Estructura de `Errors/index.json` y `Changes/index.json` (agrupados por versión; cada bloque lista los archivos MD de esa versión):

```json
{
  "versions": [
    {
      "version": "2.3.0",
      "errors": [
        "./StepLauncher-2.3.0/StepLauncher-Error-1.md",
        "./StepLauncher-2.3.0/StepLauncher-Error-2.md"
      ]
    }
  ]
}
```

`Changes/index.json` usa la clave `changes` en cada bloque (mismo formato).

**TODAS las rutas de los JSON son RELATIVAS a la ubicación del propio `index.json`** (p. ej. dentro de `Releases/index.json` → `./StepLauncher-2.3.0/news.json`). **NUNCA** rutas absolutas, nunca `Changelogs/...` ni `./Releases/...` desde la base: el launcher resuelve cada ruta partiendo del directorio donde está el índice que la contiene.

### Paso 5 — Cerrar la trazabilidad

Actualizar los campos de auditoría de las entradas que entran en esta versión:

- Errores corregidos: `- **Fixed?**: Sí — corregido y mencionado en la release StepLauncher-X.Y.Z.`
- Cambios añadidos: `- **Release**: StepLauncher-X.Y.Z — en este release se menciona que fue añadido.`

Así la auditoría queda enlazada de punta a punta: error → fix → release → noticia.

---

## Plantilla de un Error

```markdown
# Errors/StepLauncher-X.Y.Z/StepLauncher-Error-N.md — <título breve>

- **Fecha**:     <YYYY-MM-DD>
- **Versión**:   <X.Y.Z o "en desarrollo">
- **Estado**:    encontrado | en corrección | corregido
- **Fixed?**:    Sí — corregido y mencionado en la release StepLauncher-X.Y.Z. | No — sigue pendiente.

## Síntoma
Qué se veía/rompía y qué comandos afectaba.

## Causa raíz
Archivo(s), línea(s), flujo de llamadas y por qué ocurría.

## Diagnóstico y evidencia
Cómo se encontró (logs, reproducción, trazas) y qué se verificó.

## Solución aplicada
Cambio mínimo, sin tocar APIs ni arquitectura innecesariamente.

## Regla aprendida
Qué evitar en el futuro para que no vuelva a pasar.

## Verificación
Comandos/validaciones que confirmaron el arreglo
(`go build ./...`, `wails build`, `wails dev`...).
```

## Plantilla de un Cambio

```markdown
# Changes/StepLauncher-X.Y.Z/StepLauncher-Change-N.md

- **Fecha**: <YYYY-MM-DD>
- **Versión**: <X.Y.Z>
- **Release**: StepLauncher-X.Y.Z — en este release se menciona que fue añadido.

## Qué cambió
Archivos/componentes tocados (backend Go, frontend Vue, config...).
Si la tarea agrupa varios cambios, enumerarlos con subsecciones (`### 1.`, `### 2.`, ...).

## Por qué
Motivación del cambio o mejora.

## API afectada
Bindings, funciones públicas, modelos o eventos —si hay.

## Comportamiento anterior/nuevo
Qué hacía antes y qué hace ahora (solo si aplica).

## Cómo verificar
Comandos de build/dev (`bun run build`, `go build ./...`, `wails dev`...).
```

## Plantilla de un Release

Carpeta `Releases/StepLauncher-X.Y.Z/` con:

```markdown
# Releases/StepLauncher-X.Y.Z/StepLauncher-Release-X.Y.Z.md

- **Fecha**: <YYYY-MM-DD>
- **Versión**: <X.Y.Z>

## Funcionalidades nuevas
- ...

## Correcciones (con enlaces a los Errors/ de la auditoría)
- [StepLauncher-Error-N](../../Errors/StepLauncher-X.Y.Z/StepLauncher-Error-N.md) fue corregido...
- ...

## Notas para el usuario
Actualizador automático, migraciones de config, etc. (sin instrucciones de build).
```

Y el `news.json` descrito arriba.

## Auditoría de bugs complejos (reglas para agentes)

Cuando un agente investiga un bug complejo, el flujo obligatorio es:

1. **Consultar primero `Changelogs/`** (y `AGENTS.md`) antes de diagnosticar o modificar.
2. **Identificar entradas relacionadas**: mismo módulo, misma versión, misma causa raíz, mismas reglas aprendidas.
3. **Inspeccionar el código actual**: el historial documenta, no sustituye al código.
4. **Rastrear dependencias relevantes** (callers/callees, bindings, sistemas compartidos).
5. **Determinar si el comportamiento actual coincide con la documentación histórica**: si el bug ya está documentado, usar la entrada como punto de partida y no re-investigarlo desde cero.
6. **No reutilizar automáticamente una solución antigua si la arquitectura cambió**: verificar que el fix documentado siga siendo válido contra el código actual.
7. **Verificar la causa raíz antes de aplicar un fix**: una hipótesis no es una causa raíz.
8. **Documentar el problema en `Changelogs/` si es nuevo** (ver "Ciclo de vida de un Error").
9. **Actualizar `AGENTS.md` solamente** cuando se descubra una regla permanente que deba aplicarse en futuras tareas.
10. **No convertir una hipótesis de la IA en una "causa raíz" sin evidencia**: distinguir entre hechos observados, hipótesis y conclusiones verificadas.

## Reglas

- **SIEMPRE consulta esta carpeta antes de diagnosticar o modificar el proyecto**: lo que hiciste pudo haber pasado antes, y la solución ya está documentada.
- **SIEMPRE crea la entrada correspondiente en la carpeta de la versión en curso** al terminar de corregir un error o implementar un cambio relevante. No se da una tarea por terminada sin su registro en `Changelogs/`.
- **Una tarea, un solo MD**: si una petición o sesión de trabajo implica varios cambios o errores relacionados, se documentan juntos en una única entrada (con subsecciones); no se fragmenta la documentación en varias entradas (ver "Una tarea, un solo MD").
- **Cada versión nueva crea sus carpetas** `Errors/StepLauncher-X.Y.Z/`, `Changes/StepLauncher-X.Y.Z/` y `Releases/StepLauncher-X.Y.Z/`, y la numeración `N` de errores y cambios **empieza de nuevo desde 1** en cada una.
- **Los `index.json` se regeneran SOLO al publicar la release** (con `generate_indexes.ps1`), nunca al crear un error o cambio: esos se registran en su carpeta de versión y listos; el generador los recogerá cuando toque la release.
- **Toda release publicada necesita su carpeta `StepLauncher-X.Y.Z/`** con el changelog completo y su `news.json` para el centro de noticias — no basta con mencionar la release suelta.
- Contenido en **español** (los nombres de carpetas y archivos, en inglés).
- No borrar entradas antiguas: el historial es la auditoría del proyecto.
- Cuando el arreglo introduzca una **regla de arquitectura nueva** (p. ej. concurrencia, mutex, bindings), además de `Changelogs/` reflejala también en `AGENTS.md` si la regla debe aplicarse siempre.
