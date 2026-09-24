# Servidores multijugador — cómo aceptar cuentas StepLauncher

Cualquier servidor Minecraft Java en `online-mode` acepta jugadores con
cuenta StepLauncher si verifica contra este servicio con authlib-injector.
El servidor **no** habla con Mojang/Microsoft: todo el login pasa por aquí.

## Requisitos del servidor

- `online-mode=true` en `server.properties` (si está en `false`, no hay
  verificación y entra cualquiera con cualquier nombre: no usar).
- `enforce-secure-profile=false` en `server.properties` (1.19+): este
  servicio aún no emite certificados de chat (`player/certificates`), así
  que con la firma de chat obligatoria los jugadores no podrían chatear.
- El Worker reachable por HTTPS desde el host del servidor
  (producción: `https://steplauncher-yggdrasil.stepnicka012.workers.dev/`).
- `authlib-injector.jar` (descargable desde
  `https://authlib-injector.yushi.moe/`).

## Arranque

```bash
java -javaagent:authlib-injector.jar=https://steplauncher-yggdrasil.stepnicka012.workers.dev/ -jar server.jar nogui
```

La URL es la **raíz** del servicio (con `/` final). El injector descubre el
resto solo (`sessionserver/...`, metadatos, whitelist).

## Cómo entra un jugador StepLauncher

1. Entra al launcher con su cuenta AuthLib (la misma de la web) y lanza el juego.
2. Al conectarse, su cliente anuncia la sesión (`join`) y el servidor la
   verifica (`hasJoined?username=&serverId=`) contra este Worker.
3. Si el `join` tiene menos de 30 s y el token sigue vivo → `200` con perfil
   + texturas firmadas: entra con su skin. Si no → `204` y el servidor lo
   expulsa (`Failed to verify username` / sesión inválida).

## Quién puede entrar (y quién no)

- ✅ Cuentas StepLauncher con correo confirmado y perfil (`selectedProfile`).
- ❌ Cuentas premium de Mojang/Microsoft: un servidor verifica contra UN solo
  backend de sesiones. Mezclar ambas comunidades exige un proxy con fallback
  entre backends (fuera del MVP; documentar cuando exista).
- ❌ Nombres de StepLauncher fuera de la regla Minecraft (3–16,
  `[A-Za-z0-9_]`): el juego/servidor los rechaza aunque el login pase.

## Proxies (avanzado)

- **Conexión directa o Velocity** con soporte de injector: el flujo estándar
  de arriba vale sin más.
- **BungeeCord con `online-mode=false` detrás**: requiere reenvío de IP y de
  identidad (BungeeGuard o similar) + el backend en online; si el `ip` no
  coincide o el handshake no se reenvía, el `hasJoined` falla. Probar en
  staging antes de producción.

## Sin acceso a los argumentos JVM (hostings compartidos)

Si el hosting no deja agregar el `-javaagent`, la alternativa es reemplazar
la URL de verificación dentro del `.jar` (como documenta Ely.by para
BungeeCord): en `InitialHandler.class` (proxy) o en el authlib del servidor,
cambiar `https://sessionserver.mojang.com/session/minecraft/hasJoined?username=`
por `https://steplauncher-yggdrasil.stepnicka012.workers.dev/sessionserver/session/minecraft/hasJoined?username=`
y usar `online_mode=true`. Es cirugía manual: solo si no hay otra opción.

## Versiones soportadas

- **1.7.2+**: vía authlib-injector (cubre login, sesión y skins legacy con su
  polyfill local). Las skins pre-1.8 salen del polyfill o de
  `GET /skins/MinecraftSkins/:usuario.png` de este servicio.
- **Menores a 1.7.2**: no usan Yggdrasil (login legacy) y el injector no las
  soporta; requieren parcheo manual clase por clase (ver la doc de Ely.by de
  instalación en versiones viejas). Fuera del alcance de este servicio.

## Problemas típicos

| Síntoma en el servidor | Causa probable |
|---|---|
| `Failed to verify username!` / entra y lo echa | `join` expirado (>30 s entre join y hasJoined), token revocado (`signout`), o `username` distinto al del perfil |
| Entra pero todos en Steve | El jugador no tiene skin subida (ver `file_uploads`), o el servidor no usa injector (online contra Mojang) |
| No puede chatear (1.19+) | `enforce-secure-profile=true`: ponerlo en `false` hasta tener certificados |
| `Invalid session` al instante | El servidor no tiene el `-javaagent` o apunta a otra URL que el cliente |
