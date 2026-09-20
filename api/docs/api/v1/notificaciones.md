# Notificaciones v1 — bandeja propia

Las crea el servidor (p. ej. solicitudes de amistad, respetando
`receive_notifications`); el jugador solo las lista, marca como leídas o las
descarta. Todo con Bearer + RLS: un token solo ve las de **su** cuenta
(`account`, 30/min). Sin INSERT de cliente (ver `seguridad.md`).

## `GET /v1/notifications` — bandeja propia

`?limit=` (1–50, defecto 20) · `?offset=` (defecto 0) · `?unreadOnly=` (`true`
filtra no leídas, defecto `false`). `200`:

```json
{
  "success": true,
  "data": {
    "notifications": [{ "id": "uuid", "title": "…", "body": "…", "read": false, "createdAt": "…" }],
    "unreadCount": 3, "limit": 20, "offset": 0
  }
}
```

`unreadCount` es global (no respeta `unreadOnly`): sirve para el badge del launcher.

## `PATCH /v1/notifications/:id/read` — marcar una como leída

`:id` UUID. `200` devuelve la notificación marcada. Ajena/inexistente →
`not_found` 404.

## `POST /v1/notifications/read-all` — marcar todas como leídas

Sin cuerpo. `200` → `{ markedRead: <n> }` (nº de afectadas).

## `DELETE /v1/notifications/:id` — descartar una propia

`:id` UUID. `200` → `{ deleted: true }`; ajena/inexistente → `not_found` 404.
