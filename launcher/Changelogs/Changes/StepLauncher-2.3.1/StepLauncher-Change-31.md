# Changes/StepLauncher-2.3.1/StepLauncher-Change-31.md

- **Fecha**: 2026-08-14
- **Versión**: 2.3.1
- **Release**: StepLauncher-2.3.1 — en este release se menciona que fue añadido.
- **Estado**: implementado y verificado (bun run type-check OK).

## Qué cambió

`frontend/web/src/Modals/WelcomeModal.vue` — el onboarding ahora obliga a configurar la carpeta antes de poder continuar:

### 1. "Bienvenida > Configurar carpeta": la carpeta va primero

- Nuevo orden de pasos cuando la carpeta aún **no** está configurada: `welcome → directory → customize → account` (antes: `welcome → customize → directory → account`).
- El paso `directory` ya no es saltable: en la bienvenida, cuando `dirConfigured` es `false`, se muestra una única tarjeta **"Configurar carpeta" (Obligatoria)** que lleva al paso de carpeta; desaparecen las tarjetas "Personalizar" y "Configurar mi cuenta" hasta que la carpeta quede configurada. Los indicadores de pasos usan un nuevo `stepOrder` computed (4 pasos sin carpeta configurada, 3 con ella).
- Al confirmar la carpeta sin cambios ("Continuar"), se avanza a `customize` (antes iba directo a `account`); `nextFromCustomize` siempre continúa a `account`.
- Si la carpeta ya está configurada (p. ej. tras "Guardar y reiniciar"), el flujo vuelve a ser `welcome → customize → account` y el usuario puede seguir.
- `dirConfigured` arranca en `false` en lugar de `true`: si `GetDirectorySettings` fallara, el launcher no asume que la carpeta está configurada y el paso sigue siendo obligatorio.

### 2. Detección de `.minecraft` en el paso de carpeta

- Se mantiene el aviso existente: si existe la carpeta `.minecraft` oficial (`dirInfo.minecraftExists`), el paso `directory` muestra "Detectamos una instalación de Minecraft" con el botón "Usar" para adoptar el modo Minecraft sin tocar la selección manual.

## Por qué

El usuario pidió que en el primer inicio la configuración de la carpeta fuera lo primero tras la bienvenida y **obligatoria**: solo después de configurarla (o tras reiniciar con la carpeta ya guardada) se puede seguir con personalización y cuenta. Antes, el paso de carpeta podía saltarse por completo.

## API afectada

- Sin cambios en bindings de Wails ni en el backend: reutiliza `GetDirectorySettings`, `SetDirectoryMode`, `PickDirectory`, `RestartApp` y `SetFirstLaunchDone` existentes.

## Comportamiento anterior/nuevo

- **Antes**: la bienvenida ofrecía "Personalizar" y "Configurar mi cuenta" directamente; la carpeta solo se pedía si `configured` era `false` y podía omitirse (los pasos eran `welcome → customize → directory → account`, y `welcome` podía ir directo a `account`).
- **Ahora**: si la carpeta no está configurada, la única salida de la bienvenida es "Configurar carpeta" (obligatoria); sin configurarla no se llega a personalizar ni a crear cuenta; al reiniciar con la carpeta guardada el usuario puede seguir con el resto del onboarding.

## Cómo verificar

- `bun run type-check` en `frontend/`: OK.
- Manual: borrar `%APPDATA%\StepLauncher\directory.json` (y `launcher_config.json` con `firstLaunch: true`) → iniciar: la bienvenida solo permite "Configurar carpeta"; elegir modo (si hay `.minecraft`, aparece el aviso "Detectamos una instalación de Minecraft" con "Usar"); "Guardar y reiniciar" → al volver, la carpeta aparece configurada y se puede seguir a personalizar/cuenta; si solo se pulsa "Continuar" sin cambiar nada, se pasa a personalizar.
