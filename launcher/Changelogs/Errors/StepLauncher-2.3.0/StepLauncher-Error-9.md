# StepLauncher-Error-9: Stack overflow al lanzar Minecraft (recursión infinita en `SubstituteVars`)

- **Fecha**: 2026-08-05
- **Estado**: corregido
- **Fixed?**: Sí — corregido y mencionado en la release `StepLauncher-2.3.0`.
- **Versión afectada**: 2.3.0 (productVersion de `wails.json`)

---

## 1. Síntoma

Al pulsar **Jugar** en una versión/perfil, el launcher moría con un crash del
runtime de Go (volcado en `WAILS_ERROR.txt`): `goroutine stack exceeds
1000000000-byte limit` con cientos de frames repetidos del mismo método
`SubstituteVars`, es decir, **stack overflow por recursión sin límite**.

## 2. Causa raíz

En `internal/Core/Launcher/Helpers/Args.go`, la antigua implementación de
`SubstituteVars` se **llamaba a sí misma de forma recursiva incondicional**
cuando el valor resuelto de una variable volvía a contener `${...}`:

```go
// Antes (aproximación): la resolución de un valor que contenía otra
// variable volvía a entrar en SubstituteVars sin ninguna condición de salida.
func SubstituteVars(template string, vars map[string]string) string {
    ...
    val, ok := vars[key]
    if ok {
        if strings.Contains(val, "${") {
            val = SubstituteVars(val, vars) // recursión SIN límite
            ...
        }
    }
}
```

Los argumentos de lanzamiento reales (jvm/game args) contienen cadenas de
variables anidadas y/o con referencias cruzadas (un valor guarda la expresión
`${...}` de otro), por lo que la recursión nunca encontraba una base y
desbordaba la pila en **cada lanzamiento**.

## 3. Solución aplicada

Reescritura **iterativa** con un número limitado de pasadas: se resuelven las
`${clave}` de izquierda a derecha por pasada y se repite solo mientras haya
reemplazos reales, con un tope de 8 pasadas para no entrar en bucles infinitos.
Las variables sin valor en el mapa se dejan **tal cual** (no se re-procesan), y
las llaves `{` sin cierre tampoco se tocan:

```go
func SubstituteVars(template string, vars map[string]string) string {
    if template == "" || !strings.Contains(template, "${") {
        return template
    }
    current := template
    for pass := 0; pass < 8; pass++ {
        if !strings.Contains(current, "${") {
            break
        }
        // ... barre ${key}, sustituye val si existe, o conserva la
        // secuencia ${...} intacta; sale si no hubo reemplazos.
    }
    return current
}
```

El comentario del método deja constancia explícita: **no puede volver a llamarse
a sí misma sin condición** porque antes provocaba stack overflow en cada
lanzamiento.

## 4. Verificación

- `go build ./...` → OK.
- `bun run build` (dentro de `frontend/`) → OK (`vue-tsc --build` + `vite build`).
- Traza mental de la cadena de lanzamiento: `Launch` mergea la config del
  launcher y del perfil → `BuildJVMArgs`/`BuildGameArgs` → `SubstituteVars` →
  con la nueva versión se itera a lo sumo 8 pasadas y siempre termina, incluso
  con variables anidadas o sin resolver.

## 5. Regla aprendida

Las funciones de sustitución de templates con variables **nunca** deben
re-llamarse a sí mismas de forma incondicional: si un valor contiene `${...}`,
procesarlo iterativamente (pasadas limitadas) o plantear una condición de
salida real. Todo código que resuelva cadenas contra un mapa de variables debe
garantizar la terminación (límite de pasadas) para escenarios de variables
anidadas o autocontenidas.