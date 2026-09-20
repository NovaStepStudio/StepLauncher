# Salud v1 — sonda pública

Sin auth, sin secretos, sin Supabase. Para uptime, despliegues y verificar el
versionado. Rate-limit `public` (120/min).

## `GET /v1/health`

```bash
curl http://localhost:8787/v1/health
```

`200`:

```json
{
  "success": true,
  "data": { "status": "ok", "version": "v1", "env": "local", "time": "2026-09-20T06:00:00.000Z" }
}
```

- `version` = `"v1"` siempre en esta versión (el launcher detecta capacidades).
- `env` = `API_ENV` (`local` | `preview` | `production`).
- `time` = hora del edge en ISO (no usar como reloj autoritativo).

## `GET /` (índice, fuera de versión)

```bash
curl http://localhost:8787/
```

`200`:

```json
{
  "success": true,
  "data": {
    "service": "steplauncher-api",
    "versions": ["v1"],
    "docs": "ver carpeta docs/ del repo",
    "health": "/v1/health"
  }
}
```

Único endpoint fuera de `/vN` junto a la sonda. Sin datos sensibles.
