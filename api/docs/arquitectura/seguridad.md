# Seguridad mínima en endpoints de cuenta (obligatoria, no negociable)

Todo endpoint que lea o modifique datos del usuario debe cumplir, **por orden**:

1. `rateLimit(preset)` — `account` por defecto, `sensitive` en login/registro/
   recuperación/refresh, `public` solo en sondas sin datos.
2. `requireAuth()` — el dueño sale de `c.get("user").id` (Bearer validado en
   Supabase con `auth.getUser()`). **Prohibido** aceptar `userId`/`email` del
   cliente como identidad. El 401 es genérico: no distingue ausente de expirado
   de inválido.
3. `validateJson` / `validateQuery` / `validateParam` con Zod (`.strict()` donde
   aplique). El 400 trae `details` por campo, nunca datos sensibles.
4. Acceso a datos con `supabaseForUser()` (RLS activo: `auth.uid() = user_id` en
   `USING` + `WITH CHECK`). `supabaseAdmin()` solo si es imprescindible
   (alta de cuenta, resoluciones `SECURITY DEFINER`, borrado de otro), documentando
   el porqué, y jamás devolviendo su resultado crudo con campos sensibles.
   Funciones `SECURITY DEFINER` con `search_path` fijo y `REVOKE ALL … FROM
   public, anon, authenticated`.
5. Errores genéricos (`unauthorized`, `not_found`, `profile_not_found`,
   `rate_limited`) sin stack/SQL/secretos, con `requestId`. Las contraseñas solo
   viajan en tránsito HTTPS y jamás se registran en logs.
6. Tablas de usuario con RLS `auth.uid() = user_id` y prueba cruzada A-vs-B antes
   de mergear (A no lee/escribe lo de B).

## Estado v1 (verificado contra el código)

| Capa | Estado |
|---|---|
| `requireAuth()` en todos los endpoints de cuenta | Sí (auth/register-login-refresh son públicos por necesidad) |
| RLS propia en `profiles`, `privacy_settings`, `file_uploads`, `notifications` | Sí |
| Lectura social solo vía funciones revocadas | Sí (`resolve_*`, `search_users`, `my_friends`, `block_*`, `equipped_cosmetics`) |
| Buckets sin listado global | Sí (`avatars`/`banners` públicos sin SELECT amplio; `skins`/`capes` privados) |
| CORS sin wildcard, cabeceras duras, `X-Request-Id` | Sí |
| Rate-limit distribuido (KV) | **No**: en memoria por instancia (límite conocido, ver ADR-006) |
| `DELETE /v1/accounts/me` (borrado de cuenta) | **No**: pendiente con ADR (ver `docs/cuentas/README.md`) |

## Límites conocidos (no son bugs, son decisiones)

- El rate-limit se multiplica por nº de instancias (sin KV no hay conteo global).
- La quota 10/día es aproximada bajo subidas simultáneas (conteo sin bloqueo).
- Sin transacción entre Storage y `file_uploads`: el handler limpia huérfanos
  best-effort y el error se registra con `requestId`.

Si un punto de la checklist falla, no se mergea.
