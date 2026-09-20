# Versionado (v1, v2, v3…)

El launcher instalado no se actualiza a la vez: **una versión publicada nunca se rompe**.
Evolucionar = añadir versión nueva, no editar la anterior para cambiar su contrato.

## Reglas

1. Toda ruta pública cuelga de `/vN` (`/v1/health`, `/v1/accounts/me`). Solo `/` (índice)
   puede vivir fuera de versión.
2. Una versión nueva es una carpeta nueva: `src/routes/v2/`, `src/schemas/v2/`,
   `docs/api/v2/`, `db/v2/`. Se crea **copiando** la anterior y evolucionando la copia.
3. Cambios **compatibles** (añadir campo opcional, endpoint nuevo, índice SQL nuevo)
   pueden entrar en la misma versión (en SQL: regenerar `db/v1/install.sql`).
4. Cambios **incompatibles** (quitar/renombrar campo, cambiar semántica de un 200,
   exigir auth donde no había, `DROP COLUMN`) exigen versión nueva.
5. Deprecación: marcar en `docs/api/vX/README.md` con `Deprecated: usar /vY/...`,
   mantener la versión vieja al menos 2 releases del launcher y monitorizar uso
   antes de apagar (apagar = decisión con ADR + aviso en changelog del launcher).

## Ejemplo: de v1 a v2

```text
src/routes/v1/accounts.ts   # intacto, sigue sirviendo al launcher viejo
src/routes/v2/accounts.ts   # copia evolucionada (p. ej. incluye `profile`)
src/schemas/v2/account.ts   # copia evolucionada de los contratos
docs/api/v2/cuentas.md      # contrato nuevo documentado aparte
db/v2/install.sql           # paquete SQL nuevo (migra desde el estado v1)
src/index.ts                # app.route("/v1", v1); app.route("/v2", v2);
```

Nunca se comparten esquemas entre versiones: v2 **copia** de v1, nunca reutiliza
(ver `api/AGENTS.md` §1).

## Respuesta de versión

`GET /v1/health` devuelve `{ "version": "v1" }` y `/` lista `versions: ["v1"]`
para que el launcher detecte capacidades sin adivinar.
