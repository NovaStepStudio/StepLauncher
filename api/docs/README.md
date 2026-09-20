# Documentación de StepLauncher API

Esta carpeta es **esencial y obligatoria**: si un endpoint o una decisión no está aquí,
para el resto del equipo **no existe**. Un dev nuevo debe entender qué se decidió,
por qué y cómo añadir endpoints sin romper los anteriores solo leyendo `docs/`.

## Mapa

```text
docs/
  README.md                  # Este índice: cómo leer la docs
  arquitectura/              # Decisiones transversales (el "por qué")
    decisiones.md            # ADRs: runtime, Supabase, versionado, SQL, docs...
    estructura.md            # Dónde va cada archivo y cómo añadir un endpoint
    versionado.md            # Reglas v1, v2, v3: nunca romper, cómo deprecar
    seguridad.md             # Checklist mínimo para endpoints de cuenta
    secretos-y-entorno.md    # Vars vs secrets, wrangler, revisión pre-commit
  oauth/                     # Sesiones y Auth: modelo actual + camino PKCE futuro
    README.md
  cuentas/                   # Modelo de cuenta, perfil y vida social (dominio maestro)
    README.md
  api/
    v1/                      # Contrato público v1 (lo que consume el launcher)
      README.md              # Base URL, sobre de respuesta, auth, códigos, límites
      health.md              # Sonda de salud y versión
      auth.md                # Registro, login, renovación y cierre de sesión
      cuentas.md             # Perfil, presencia, email, contraseña, avatar, banner, cosméticos, privacidad
      archivos.md            # Skins y capas: subida, listado, URL firmada, borrado
      notificaciones.md      # Bandeja propia: listar, leer, descartar
      amigos.md              # Solicitudes, amistades y lista negra
      usuarios.md            # Búsqueda pública de jugadores
      comunidad.md           # Previsualización de perfiles + relación
    # v2/ se creará copiando v1/ cuando toque evolucionar sin romper.
```

## Regla de dominios (dónde documentar cada cosa)

- ¿Contrato HTTP exacto (ruta, método, 200/400/401, ejemplos)? → `docs/api/vX/`.
- ¿Modelo de cuenta, perfil, privacidad, amistad, cosméticos? → `docs/cuentas/` + `docs/api/vX/`.
- ¿Sesiones, tokens, refresh, PKCE, launcher desktop? → `docs/oauth/` + `docs/api/vX/auth.md`.
- ¿Cómo versionar o dónde poner un archivo? → `docs/arquitectura/`.
- ¿Tablas, RLS, funciones SQL, paquetes `db/vN/`? → `db/vN/README.md` (la docs de
  SQL vive junto al SQL); aquí solo el porqué (ADRs).

## Cómo documentar un endpoint nuevo (obligatorio)

1. Crear/actualizar `docs/api/vX/<dominio>.md` con método, auth, límites, cuerpo,
   respuesta, errores y ejemplo `curl`.
2. Si toca cuentas, sesiones u OAuth, actualizar también `docs/cuentas/README.md`
   o `docs/oauth/README.md`.
3. Si se tomó una decisión (elegir tabla, flujo, librería, límite), añadir ADR en
   `docs/arquitectura/decisiones.md`.
4. Si cambia el SQL, evoluciona el paquete `db/vN/` (ver `db/v1/README.md`).

## Convenciones

- Toda comunicación y comentarios en **español**.
- Códigos de error estables en `snake_case` (`unauthorized`, `not_found`…): el
  launcher los gestiona por `code`, nunca por el mensaje.
- Los ejemplos usan `https://steplauncher.tudominio` como base; en local es
  `http://localhost:8787`.
