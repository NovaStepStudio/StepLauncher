# StepLauncher-Error-2: accessToken guardado sin los guiones del UUID

- **Fecha**: 2026-08-04
- **Estado**: corregido
- **Fixed?**: Sí — corregido y mencionado en la release `StepLauncher-2.3.0`.
- **Versión afectada**: 2.3.0 (productVersion de `wails.json`)

---

## 1. Síntoma

Al iniciar sesión en una cuenta AuthLib, el `accessToken` quedaba guardado en
`launcher_accounts.json` como `e9bfc7b1c7c444dc9dc43de03e3c5109` (32 hex sin
separadores), mientras el mismo token se ve/recupera como
`e9bfc7b1-c7c4-44dc-9dc4-3de03e3c5109` (UUID con guiones). El usuario percibía
que el launcher no guardaba "lo que el servidor retorna".

## 2. Causa raíz

Los accessTokens Yggdrasil son **UUIDs**, y cada servidor de autenticación los
devuelve en el JSON con un criterio distinto de separadores: algunos los mandan
con los guiones del UUID canónico (8-4-4-4-12) y otros sin ellos. El manager de
cuentas (`internal/Core/Accounts/Manager.go`) guardaba el token **tal cual lo
devolvía el servidor**, sin normalizar, de modo que el almacén podía quedar con
una representación y la vista del servidor con otra del mismo UUID.

No había ninguna transformación que quitara los guiones: el problema era la
ausencia de un formato fijo de guardado. Además, los tokens ya persistidos con
el formato "sin guiones" no se reparaban nunca.

## 3. Solución aplicada

Nuevo helper `normalizeAccessToken` en `internal/Core/Accounts/Auth.go`:

```go
func normalizeAccessToken(token string) string {
	token = strings.TrimSpace(token)
	if len(token) != 32 {
		return token
	}
	if _, err := hex.DecodeString(token); err != nil {
		return token
	}
	return fmt.Sprintf("%s-%s-%s-%s-%s",
		token[0:8], token[8:12], token[12:16], token[16:20], token[20:32])
}
```

- Si el token es un UUID de 32 hexes (con o sin guiones) se re-formatea
  siempre al UUID canónico con guiones: es el mismo UUID para el servidor
  (parser los ignora), así que la sesión no cambia.
- Si el token no parece un UUID (tokens opacos de servidores no estándar), se
  deja intacto.

Se aplica en todos los puntos de escritura de `internal/Core/Accounts/Manager.go`:
- `apply` (creación/actualización manual).
- `LoginAuthLib` (login Yggdrasil).
- `ResolveForLaunch` y `RefreshAuthLib` (refresh automático y manual).

Y en `sanitizeMap` + `Load()` (reparación en carga): los tokens ya guardados
sin guiones se normalizan y el archivo se re-escribe en disco la primera vez
que se carga, migrando los datos viejos sin intervención del usuario.

## 4. Verificación

- `go build ./...` → OK.
- `bun run build` en `frontend/` (type-check + build) → OK.
- Con una cuenta AuthLib existente, al abrir el launcher el accessToken del
  `launcher_accounts.json` pasa al formato `8-4-4-4-12` automáticamente.
- Los logins nuevos guardan el token ya normalizado.

## 5. Regla aprendida

Los datos que dependen de un servidor externo (tokens, UUIDs, fechas) no deben
guardarse "tal cual llegan": hay que fijar un formato canónico de persistencia
y normalizarlo en **todos** los puntos de escritura (y migrar lo antiguo en la
carga) para que el almacen sea estable aunque los servidores varíen el formato.