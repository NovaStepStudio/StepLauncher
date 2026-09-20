# Archivos v1 — skins y capas (privados, con quota)

Buckets `skins`/`capes` privados; la descarga es por URL firmada de 1h.
Quota: **10 skins+capas por día y usuario** (ventana deslizante 24h contando
`file_uploads`; aproximada bajo carreras). Todo exige Bearer + `account`;
las subidas además pasan por `uploadQuota()` (`upload_quota_exceeded` 429).
Avatar y banner viven en `cuentas.md` (no cuentan para esta quota).

## `POST /v1/files/skin` · `POST /v1/files/cape` — subir (`multipart`)

Campo `file`: **PNG real** (firma `89 50 4E 47 0D 0A 1A 0A`, no la extensión),
**≤1 MB**. Se guarda en `<bucket>/<user_id>/<kind>-<timestamp>.png` (sin upsert)
y se anota en el libro; si el libro falla, se borra el huérfano best-effort.

```bash
curl -X POST http://localhost:8787/v1/files/skin \
  -H 'Authorization: Bearer …' -F 'file=@skin.png'
```

`201`:

```json
{ "success": true, "data": { "id": "uuid", "kind": "skin", "sizeBytes": 1234, "mime": "image/png", "createdAt": "…" } }
```

Errores: `upload_quota_exceeded` 429 · `file_too_large` 400 · `invalid_file_type`
400 (`Solo PNG válido.`) · `upload_failed` 500.

## `GET /v1/files` — listar subidas propias

`?limit=` (1–50, defecto 20) y `?offset=` (defecto 0). `200`:

```json
{
  "success": true,
  "data": {
    "uploads": [{ "id": "uuid", "kind": "skin", "sizeBytes": 1, "mime": "image/png", "createdAt": "…" }],
    "limit": 20, "offset": 0,
    "dailyQuota": { "limit": 10, "window": "24h", "kinds": ["skin", "cape"] }
  }
}
```

## `GET /v1/files/:id/url` — URL firmada (1h) para descargar

Solo archivos propios; ajeno/inexistente → `not_found` 404 (genérico).
`200` → `{ url, expiresIn: 3600 }`. El launcher descarga el PNG desde `url`
antes de que caduque (no la almacenes: pide una nueva cada vez).

## `DELETE /v1/files/:id` — borrar archivo propio

Borra de Storage **y** del libro (libera quota del día). `200` → `{ deleted: true }`;
ajeno/inexistente → `not_found` 404.
