# StepLauncher-Error-8: Crash al cancelar una descarga (`sync: unlock of unlocked mutex`)

- **Fecha**: 2026-08-04
- **Estado**: corregido
- **Fixed?**: Sí — corregido y mencionado en la release `StepLauncher-2.3.0`.
- **Versión afectada**: 2.3.0 (productVersion de `wails.json`)

---

## 1. Síntoma

Al cancelar una descarga activa (botón Cancelar del modal de instalación), la
aplicación moría con un fatal error del runtime de Go:

```
fatal error: sync: unlock of unlocked mutex

goroutine 393 [running]:
internal/sync.(*Mutex).unlockSlow(...)
sync.(*Mutex).Unlock(...)
StepLauncher/internal/Core/Downloader.(*Manager).runDownload(...)
        .../internal/Core/Downloader/Manager.go:511 +0x30b8
StepLauncher/internal/Core/Downloader.(*Manager).Start.func1()
        .../internal/Core/Downloader/Manager.go:187
StepLauncher/internal/Core/Downloader.(*Queue).Add.func1()
        .../internal/Core/Downloader/Queue.go:26
```

## 2. Causa raíz

En el bloque final de `runDownload()` (`internal/Core/Downloader/Manager.go`),
la lectura del estado final hacía un **doble `Unlock()` del mismo mutex** cuando
la descarga se cancelaba o pausaba:

```go
dl.mu.Lock()
shouldComplete := dl.State == StateDownloading || dl.State == StateVerifying
hasFailed := dl.failed
if !shouldComplete {
    hasError := dl.State == StateError
    dl.mu.Unlock()          // (1) primer unlock (estado cancelado/pausado)
    if hasError {
        return
    }
}
dl.mu.Unlock()              // (2) SEGUNDO unlock del mismo mutex -> panic
```

`sync.Mutex` no permite desbloquear un mutex que ya está desbloqueado: el
runtime lo detecta y aborta todo el proceso (no es recuperable con `recover`).

La cadena de cancelación que lo disparaba: `Cancel()` → `dl.cancel()` (contexto
cancelado) → los workers de `downloadAll` salen por `checkState()` → `runDownload`
llega al bloque final con `State == StateCancelled` (no es `downloading`/`verifying`
ni `error`) → primer `Unlock()` dentro del `if`, luego cae al segundo `Unlock()`
de la línea 511 → panic.

## 3. Solución aplicada

Se reescribe el bloque para que el estado se lea bajo **un único `Lock()`**, se
libere una sola vez y luego se ramifique:

```go
dl.mu.Lock()
shouldComplete := dl.State == StateDownloading || dl.State == StateVerifying
hasFailed := dl.failed
hasError := dl.State == StateError
dl.mu.Unlock()

if !shouldComplete {
    if hasError {
        return
    }
    return
}
if hasFailed > 0 {
    m.setError(dl, fmt.Errorf("%d files failed to download", hasFailed))
} else {
    m.setCompleted(dl)
}
```

## 4. Verificación

- `go build ./...` → OK.
- Traza mental del flujo de cancelación: `Cancel()` → `dl.cancel()` → ctx
  cancelado → `checkState()` devuelve false → `downloadAll` cierra `jobs` y
  retorna → `runDownload` llega al bloque final con `State == StateCancelled` →
  `shouldComplete == false` → `return` limpio, sin segundo unlock.
- `bun run build` (dentro de `frontend/`) → OK (el rediseño de la ronda que
  acompañaba al crash también compila).

## 5. Regla aprendida

Cada `Lock()` debe tener exactamente un `Unlock()` en todos los caminos. El
patrón «unlock dentro del `if` y otro unlock al final del bloque» es propenso a
doble unlock cuando hay ramas intermedias; es más seguro leer todo el estado
bajo el lock, liberar una sola vez y ramificar después. Auditar siempre los
bloques con `if` internos que liberen mutex.
