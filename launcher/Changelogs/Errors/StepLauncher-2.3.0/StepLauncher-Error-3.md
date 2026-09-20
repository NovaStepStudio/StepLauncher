# StepLauncher-Error-3: `launcher_accounts.json` reportado como "corrupto" al cargar

- **Fecha**: 2026-08-04
- **Estado**: corregido
- **Fixed?**: Sí — corregido y mencionado en la release `StepLauncher-2.3.0`.
- **Versión afectada**: 2.3.0 (productVersion de `wails.json`)

---

## 1. Síntoma

Al arrancar la app, `Load()` del gestor de cuentas fallaba con:

```
WARN: archivo de cuentas corrupto, arrancando vacio: json: cannot unmarshal
string into Go struct field AccountsFile.accounts of type accounts.Account
```

El archivo `%APPDATA%\.StepLauncher\launcher_accounts.json` **no estaba
corrupto** y contenía las cuentas (visible a simple vista), pero la app las
ignoraba todas y arrancaba con el almacén vacío.

## 2. Causa raíz

**Incoherencia entre el formato que escribe `persist()` y el que espera
`Load()`** en `internal/Core/Accounts/Manager.go`:

- `persist()` escribía el almacén (AccountsFile) **envuelto en una sección**
  `"accounts"` (formato pensado para compartir fichero con otros gestores):

  ```json
  { "accounts": { "accounts": { "acc-1": {...} }, "selectedAccount": "...", "clientToken": "..." } }
  ```

- `Load()` deserializaba la **raíz** directamente a `AccountsFile`
  (`json.Unmarshal(raw, &m.data)`), que espera:

  ```json
  { "accounts": { "acc-1": {...} }, "selectedAccount": "...", "clientToken": "..." }
  ```

Con el formato escrito, la llave raíz `"accounts"` se intenta interpretar como
`map[string]*Account` y el valor `"selectedAccount": "acc-..."` (string) rompe
el parseo → error. Además, **el archivo es propio** de cuentas
(`launcher_accounts.json`), así que el diseño de "secciones compartidas" era
innecesario y erróneo.

Consecuencia colateral: el parseo fallido podía dejar `m.data` **parcialmente
poblado** (Go hace merge sobre mapas existentes al deserializar), generando
cuentas fantasma (`accounts`, `selectedAccount`, `clientToken` como cuentas
vacías) si se reintentaba el parseo sobre el mismo objeto.

## 3. Solución aplicada

- `persist()` ahora escribe el `AccountsFile` **directamente en la raíz** del
  archivo (sin sección envolvente), que es el formato que `Load()` espera.
- `Load()` gana compatibilidad con el formato legacy: si el parseo directo
  falla, intenta cargar desde la sección `"accounts"` (el formato de los
  primeros builds) **en una variable fresca** (nunca sobre `m.data`, para no
  arrastrar residuos del intento fallido) y lo asigna al final.
- El `persist()` que ya se hacía tras `sanitizeLoaded()` migra el archivo
  legacy al formato nuevo automáticamente en la primera carga, sin
  intervención del usuario.

## 4. Verificación

- `go build ./...` → OK.
- Replicado el flujo exacto de `Load()` en un programa de prueba con el
  `launcher_accounts.json` real del usuario: con el formato legacy se migran
  las 5 cuentas del almacén real (2 AuthLib + datos), se conserva
  `selectedAccount` y `clientToken`, y el re-persist genera el formato de raíz
  correcto. Sin la variable fresca, aparecían las cuentas fantasma; con ella,
  2 cuentas exactas.
- Los tokens JWT (Ely.by) se conservan intactos (no son UUIDs de 32 hex, el
  `normalizeAccessToken` no los toca).

## 5. Regla aprendida

El formato de escritura y el de lectura de un JSON persistido deben definirse
**juntos** y verificarse con el mismo test. Al añadir compatibilidad hacia
atrás, deserializar **siempre en una variable fresca** y asignarla después:
`encoding/json` hace merge sobre mapas/structs ya poblados y un intento
fallido deja residuos que corrompen el intento siguiente.