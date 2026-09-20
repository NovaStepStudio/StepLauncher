# Errors/StepLauncher-2.3.1/StepLauncher-Error-2.md — Instaladores legacy de Forge (1.8.9 y anteriores) fallaban al instalar

- **Fecha**: 2026-08-10
- **Versión**: 2.3.1
- **Estado**: corregido
- **Fixed?**: Sí — corregido y mencionado en la release StepLauncher-2.3.1.

## Síntoma

Instalar Forge antiguo (p. ej. `forge-1.8.9-11.15.1.2318-1.8.9-installer.jar`) fallaba: el flujo buscaba un `version.json` dentro del instalador que no existe en los instaladores legacy, así que nunca se escribía el `version.json` del perfil y la versión quedaba incompleta.

## Causa raíz

- El instalador legacy de Forge (1.8.9 y anteriores) **no incluye `version.json`**: el perfil de versión vive embebido como `versionInfo` dentro de `install_profile.json`.
- `internal/Core/ModLoader/Installer/Executor.go` solo intentaba leer `version.json` del zip; al no encontrarlo dejaba `versionJsonData` vacío y la instalación del perfil nunca se completaba.
- Evidencia (inspección del jar): `forge-1.8.9-11.15.1.2318-1.8.9-installer.jar` trae `versionInfo` inline (`id: "1.8.9-forge1.8.9-11.15.1.2318-1.8.9"`, `inheritsFrom: "1.8.9"`, `mainClass: net.minecraft.launchwrapper.Launch`, `assets: "1.8"`, 19 librerías) y `processors` (1), pero ningún `version.json`.

## Solución aplicada

- `internal/Core/ModLoader/Installer/Executor.go`: nueva `InstallProfile.VersionInfo json.RawMessage`. Si no se encuentra `version.json`, se parsea `install_profile.json` y se usa su `versionInfo` como `versionJsonData` (fallback para instaladores legacy).

## Regla aprendida

Para los instaladores de Forge/NeoForge, la fuente del perfil de versión puede ser `version.json` (instaladores modernos) **o** `install_profile.versionInfo` (instaladores legacy); nunca asumir que existe `version.json`.

## Verificación

- `go build ./...` OK.
- `bun run build` (frontend) OK.
- Comportamiento: Forge 1.8.9 (`11.15.1.2318`) extrae el perfil desde `versionInfo` y queda registrado en `versions/`.