# Comunidad v1 — previsualización de perfiles

Exige Bearer + `account`. Resuelve por UUID de cuenta, nombre de usuario o UUID
de Minecraft y devuelve la preview con visibilidad por capas. **El email NUNCA
se expone aquí.**

## `GET /v1/community/profiles/:identifier` — previsualizar

```bash
curl http://localhost:8787/v1/community/profiles/Steve_01 -H 'Authorization: Bearer …'
```

### Capas de visibilidad (en este orden)

1. **Uno mismo**: siempre visible (`relationship.status = "self"`).
2. **Lista negra** (cualquier sentido): `200` con **solo el aviso**, sin datos:
   - Te bloqueó → `{ blocked: true, reason: "blocked_you", message: "…" }`.
   - Lo bloqueaste → `{ blocked: true, reason: "blocked_by_you", message: "…" }`.
3. **Amigos**: siempre visibles (con `friendsSince`).
4. **Resto**: exige `searchable` en `true`; si no → `profile_not_found` 404
   (igual que inexistente: genérico a propósito).

### Respuesta completa (`blocked: false`) — `200`

```json
{
  "success": true,
  "data": {
    "blocked": false,
    "profile": {
      "userId": "uuid", "username": "Steve_01", "displayName": "Steve",
      "avatarUrl": "…", "bannerUrl": "…", "bio": "…",
      "memberSince": "2026-…", "isOnline": false
    },
    "minecraft": { "uuid": "uuid-mc", "lastVersion": "1.21" },
    "equippedCosmetics": [{ "id": "uuid", "slug": "capa-fundador", "name": "Capa del Fundador", "kind": "cape", "imageUrl": null }],
    "relationship": { "status": "none", "requestId": null, "friendsSince": null }
  }
}
```

- `equippedCosmetics`: **solo equipados** (lo no equipado es privado).
- `relationship.status`: `self` | `none` | `friends` | `pending_sent` | `pending_received`.
  Con `pending_sent`, `requestId` permite retirar la enviada
  (`POST /v1/friends/requests/:id/cancel`) sin buscarla en la bandeja.
- Inexistente o no visible → `profile_not_found` 404.
