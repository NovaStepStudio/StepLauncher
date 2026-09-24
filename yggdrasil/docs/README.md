# Documentación de StepLauncher Yggdrasil

Esta carpeta es **obligatoria**: si algo no está aquí, para el resto del equipo
**no existe**. Un dev nuevo debe entender qué es Yggdrasil, de dónde salen los
usuarios y cómo lo usa el launcher solo leyendo estos archivos.

## Mapa

```text
docs/
  README.md              # Este índice
  base-yggdrasil.md      # Qué es Yggdrasil al completo (protocolo, modelos, firmas)
  integracion-api.md     # Cómo usa la API real / Supabase como fuente de usuarios
  uso-launcher.md        # Cómo el launcher debe consumir este servicio
  servidores.md          # Cómo un servidor online acepta cuentas StepLauncher
  contrato.md            # Rutas, cuerpos, respuestas y errores exactos (MVP)
  decisiones.md          # ADRs: servicio separado, tokens, sobre, skins
```

## Reglas

- Toda comunicación y comentarios en **español**.
- El protocolo manda: los ejemplos copian los campos exactos de Mojang /
  authlib-injector (nada de sobre `ok()`/`fail()`).
- Decisión (tabla, clave, límite, flujo) → ADR en `decisiones.md`.
- Cambio de SQL → evoluciona `db/` (compatible) o crea `db/v2/`-equivalente.
