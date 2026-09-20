# Usuarios v1 — búsqueda pública de jugadores

Solo jugadores con `searchable` en `true` aparecen (se aplica en SQL con
`search_users`). Exige Bearer + `account` (no es anónima: evita scraping).

## `GET /v1/users/search?q=…&limit=…` — buscar visibles

- `q`: 1–40 caracteres (se escapan `%`, `_`, `\`; busca en usuario y nombre).
- `limit`: 1–20, defecto 10.

```bash
curl 'http://localhost:8787/v1/users/search?q=ste&limit=10' -H 'Authorization: Bearer …'
```

`200`:

```json
{
  "success": true,
  "data": { "users": [{ "username": "Steve_01", "displayName": "Steve", "avatarUrl": "…", "mcUuid": "…", "isOnline": false }] }
}
```

Notas:

- **Sin `userId`**: la búsqueda no expone UUIDs de cuenta (para actuar, usar la
  preview de `comunidad.md`, que sí identifica con contexto de privacidad/bloqueo).
- Si desactivaste `searchable`, **tú tampoco** apareces aquí (pero tus amigos te
  siguen viendo en su lista y en tu preview si son amigos).
- `q` vacío o `limit` fuera de rango → `validation_error` 400.
