# StepLauncher-Error-1

- **Fecha**: agosto 2026
- **Estado**: corregido
- **Fixed?**: Sí — corregido y mencionado en la release `StepLauncher-2.3.0`.
- **Versión afectada**: 2.3.0 (productVersion en `wails.json`)

---

## 1. Síntoma

`wails build` y `wails dev` quedaban colgados de forma indefinida en el paso
**"Generating bindings"**, con la app a **CPU 0%** y aparentemente sin hacer nada.
El proceso nunca llegaba a generar los bindings ni a mostrar la ventana.

## 2. Causa raíz

**Self-deadlock de `sync.RWMutex` en `internal/Config/Config.go`.**

`Manager.Save()` hacía:

```go
m.mu.Lock()          // adquiere el write-lock
defer m.mu.Unlock()
...
m.logf(...)          // DENTRO del lock → logf hace m.mu.RLock()
```

Y `logf()`:

```go
func (m *Manager) logf(...) {
    m.mu.RLock()     // intenta adquirir read-lock del mismo mutex
    ...
}
```

`sync.RWMutex` **NO es reentrante**: cuando la misma goroutine ya posee el
`Lock()` (write), su `RLock()` bloquea en la semáfora del reader esperando a que
el writer libere... pero el writer es esa misma goroutine, parqueada a la espera
de liberar. Espera circular irresoluble → deadlock permanente.

El deadlock es **incondicional**: el `RLock()` bloquea antes de leer `logFn`, así
que ocurría incluso con `logFn == nil`.

## 3. Cadena que lo disparaba

```
main()  (main.go:16)
 → NewApp()  (app.go:29)
  → Handlers.NewApp(eng, cfgPath)  (app.go:39; internal/Handlers/App.go:27)
   → Config.NewManager(configPath)  (Config.go:181)
    → m.load()  (Config.go:183)
     → m.Save()  (Config.go:212: config válida en disco / Config.go:185: sin archivo)
      → m.mu.Lock()  (Config.go:453)
       → m.logf()    (Config.go:468)  → m.mu.RLock()  (Config.go:173)  ← DEADLOCK
```

Todo esto corre en `main()`, **antes** de `wails.Run()`, que es donde Wails v2
genera los bindings (env `WailsGenerateBindings=true`). Por eso el CLI se quedaba
en "Generating bindings": el proceso nunca llegaba a ejecutar el paso.

**Cualquier deadlock en la construcción de `NewApp()` (no solo en Config) tiene
este mismo síntoma.**

## 4. Por qué CPU 0%

La única goroutine activa (la `main`) queda parqueada en
`runtime_SemacquireMutex`. No queda ninguna goroutine ejecutable y nadie más
puede liberar el lock (quien lo posee está bloqueado). El proceso queda inactivo
sin trabajo: 0% CPU.

## 5. Diagnóstico y evidencia

- Se auditaron las llamadas dentro de cada `m.mu.Lock()` de `Config.go`.
- Se encontró que **solo `Save()`** llamaba a `m.logf()` con el lock retenido;
  el resto de métodos (`Update*`, `Set*`, `Reset`, etc.) llaman a `logf` después
  del `Unlock()`.
- Verificación de cadena: `main → NewApp → Handlers.NewApp → Config.NewManager
  → load → Save` confirmada por greps en `main.go`, `app.go` y
  `internal/Handlers/App.go`.
- Se reproduce también ejecutando el binario con `WailsGenerateBindings=true`.

## 6. Solución aplicada

Cambio mínimo en `Save()`: liberar el `Lock()` **antes** de `logf()`,
conservando el log solo en éxito mediante un flag.

```go
func (m *Manager) Save() error {
    m.mu.Lock()
    saved := false
    defer func() {
        m.mu.Unlock() // libera ANTES de logf (RLock ya seguro)
        if saved {
            m.logf("Configuracion guardada: %s", m.configPath)
        }
    }()
    if m.configPath == "" { return nil }
    // ... write (n daños)
    saved = true
    return nil
}
```

Comportamiento idéntico en todos los caminos: sin log en errores, con log solo
en éxito. Sin cambios de APIs, nombres, arquitectura ni lógica de negocio.

## 7. Regla aprendida

**Ninguna función que adquiera un lock puede llamar a otra función que adquiera
el MISMO lock** (aunque sea `RLock`, y aunque el unlock sea `defer`ed).

#### Auditarlo así

Por cada `m.mu.Lock()`:
1. Lista las llamadas del cuerpo (incluidas las ayudas/internas).
2. Si alguna vuelve a tocar `m.mu` (aunque sea `RLock`), romper el lock antes o
   diferir la llamada fuera (p. ej. `m.logf(...)` después del `Unlock()`, o
   capturar los datos/log dentro del lock y loguearlo fuera).

Los métodos de `Config` siguen patrón: mutar bajo `Lock()`, `Unlock()`, y
loguear **después**.

#### Patrón correcto

```go
m.mu.Lock()
// mutar m.cfg...
m.mu.Unlock()
m.logf("...") // logf internamente hace RLock: sin lock retenido es seguro
```

## 8. Verificación

- `go build ./...` → OK.
- `wails build` → **"Generating bindings: Done."** (el paso que colgaba), frontend
  compilado, `build/bin/StepLauncher.exe` generado correctamente.
- `wails dev` no se ejecutó (proceso interactivo largo), pero el paso crítico
  (generación de bindings) es idéntico en dev y en build y quedó validado.
