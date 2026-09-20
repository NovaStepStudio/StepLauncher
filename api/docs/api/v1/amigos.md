# Amigos v1 — solicitudes, amistades y lista negra

Todo exige Bearer (`account`, salvo enviar solicitud y bloquear que usan
`sensitive`). La escritura social vive en funciones `SECURITY DEFINER` revocadas;
la privacidad se aplica **en SQL** (`resolve_friend_target`), nunca confiando en
el cliente. El email jamás se expone en este dominio.

## Solicitudes

### `POST /v1/friends/requests` — enviar (`sensitive`)

```json
{ "identifier": "amigo@ejemplo.com | Amigo_01 | uuid" }
```

Acepta email, usuario o UUID (de cuenta o de Minecraft). `201` →
`{ id, status: "pending", toUser: { userId, username, displayName, avatarUrl } }`.
Notifica al destinatario salvo que tenga `receive_notifications` en `false`.

| Error | Cuándo |
|---|---|
| `user_not_found` 404 | Inexistente, sin privacidad favorable **o te bloqueó** (genérico a propósito) |
| `cannot_add_self` 400 | A uno mismo |
| `user_blocked` 403 | Bloqueaste tú al destinatario (desbloquea primero) |
| `already_friends` 409 | Ya sois amigos |
| `already_pending` 409 | Ya hay pendiente entre ambos (en cualquier sentido) |

### `GET /v1/friends/requests?type=` — bandeja

`?type=incoming` (defecto) | `sent` | `all`. `200`:

```json
{
  "success": true,
  "data": { "requests": [{ "id": "uuid", "direction": "incoming", "status": "pending", "createdAt": "…", "respondedAt": null, "user": { "userId": "…", "username": "…", "displayName": "…", "avatarUrl": "…" } }] }
}
```

### `POST /v1/friends/requests/:id/accept` — aceptar (receptor)

Atómico en SQL (valida + crea la doble fila). `200` →
`{ id, status: "accepted", friend: {…} }` y notifica al emisor. Si no eres el
receptor o no está pendiente → `not_found` 404.

### `POST /v1/friends/requests/:id/decline` — rechazar (receptor)

`200` → `{ id, status: "declined" }`. Si no aplica → `not_found` 404.

### `POST /v1/friends/requests/:id/cancel` — cancelar enviada (emisor)

`200` → `{ id, status: "cancelled" }`. El receptor no puede cancelar (usa decline).

## Amistades

### `GET /v1/friends` — lista propia

`200`:

```json
{
  "success": true,
  "data": {
    "friends": [{ "userId": "…", "username": "…", "displayName": "…", "avatarUrl": "…", "mcUuid": "…", "friendsSince": "…", "isOnline": false }],
    "count": 2
  }
}
```

### `DELETE /v1/friends/:friendId` — romper (cualquier lado)

`:friendId` = UUID de cuenta del amigo. Borra ambas filas. `200` →
`{ removed: true }`; si no erais amigos → `not_found` 404.

## Lista negra

### `GET /v1/friends/blocks` — mi lista negra

`200` → `{ blocked: [{ user: { userId, username, displayName, avatarUrl }, blockedSince }] }`.

### `POST /v1/friends/blocks` — bloquear (`sensitive`)

```json
{ "identifier": "…" }
```

Mismo identificador que en solicitudes (**la privacidad no impide bloquear**).
Registra el bloqueo y **corta la relación**: pendientes cancelados + amistad
borrada. `201` → `{ blocked: true }`. A uno mismo o sin perfil →
`cannot_block` 400; inexistente → `user_not_found` 404.

### `DELETE /v1/friends/blocks/:userId` — desbloquear

`200` → `{ unblocked: true }`; si no estaba bloqueado → `not_found` 404.
Desbloquear no restaura la amistad: hay que solicitar de nuevo.
