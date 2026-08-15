# Changes/StepLauncher-2.3.1/StepLauncher-Change-12.md

- **Fecha**: 2026-08-09
- **Versión**: 2.3.1
- **Release**: StepLauncher-2.3.1 — en este release se menciona que fue añadido.
- **Estado**: implementado y verificado (mapeos contrastados contra datos oficiales reales de NeoForge).

## Qué cambió

`NeoForgeVersionToMcVersion` (en `internal/Core/ModLoader/Provider/Neoforge_resolver.go`) reescrito para las versiones de MC de la serie 26.x/2026: ya no antepone el prefijo `1.` y soporta las versiones de NeoForge de **4 partes** (`X.Y.Z.BUILD`). Al seleccionar una MC moderna, NeoForge vuelve a listar sus versiones reales ("No disponible" ya no se muestra).

### 1. Mapeo dual

- `SplitN(..., 4)`:
  - **4 partes** (serie 26.x, p. ej. `26.2.0.57`): MC = `X.Y`, añadiendo `.Z` solo si `Z != "0"` (`26.2.0.57` → `26.2`; `26.1.2.94` → `26.1.2`). Sin prefijo `1.`.
  - **3 partes** (series antiguas, p. ej. `21.11.45`): comportamiento anterior (`1.X[.Y]`).
- Los sufijos `-beta`/`-alpha` quedan en la última parte y no afectan al mapeo.

## Por qué

- Desde 2026 Mojang dejó de versionar con el prefijo `1.` (el manifest lista `26.2`, `26.1.2`, `26.1.1`, `26.1`), pero el resolver anteponía siempre `"1."`: mapeaba `26.2.0.57` → `"1.26.2"`, que nunca coincide con el id de MC de la UI (`"26.2"`) → NeoForge aparecía "No disponible".
- El mapeo oficial (campo `minecraft` del `install_profile.json` real): `26.2.0.57` → `"26.2"` (el tercer campo `0` se omite); `26.1.2.75` → `"26.1.2"`.
- **Forge no necesitaba cambio** (auditado): su `maven-metadata.json` ya usa las claves del manifest (`26.2`, ...) y empareja exacto con el `mcVersion` de la UI. Fabric/Quilt/LegacyFabric usan el meta propio con los ids del manifest, sin resolver local.

## API afectada

- Backend Go: solo `Neoforge_resolver.go` (`NeoForgeVersionToMcVersion`). Sin cambios en bindings, frontend ni config.

## Comportamiento anterior/nuevo

- Anterior: NeoForge mostraba "No disponible para 26.2" pese a existir `26.2.0.0-beta` … `26.2.0.57` en `maven.neoforged.net`.
- Nuevo: seleccionando MC `26.2` se listan las 52 versiones de NeoForge 26.2.x (verificado): `26.2` → 52, `26.1.2` → 91, `26.1.1` → 14, `26.1` → 33 (incluye alphas); series antiguas intactas (`1.21.11` → 45, `1.21.1` → 241, `1.20.2` → 79).

## Cómo verificar

- `go build ./...` sin errores.
- Mini-programa de reproducción con el parser exacto + el `maven-metadata.xml` real de NeoForge (evidencia en `C:\Users\STEPNI~1\AppData\Local\Temp\opencode\neoparse\`): mapeos correctos para todas las series (ver "Comportamiento anterior/nuevo").
- Verificado en runtime (2026-08-09): en el modal de instalación, MC `26.2` lista las versiones de NeoForge y su instalación completa (Change-11) es funcional.

## Regla aprendida

- El identificador de MC en el launcher SIEMPRE es el id del manifest (`26.2`); cualquier provider que derive la MC desde la versión del modloader debe contrastarse contra datos oficiales (campo `minecraft` del install_profile/pom del artefacto), no contra una convención asumida: el versionado de Mojang cambió (sin prefijo `1.` desde 2026) y el de NeoForge pasó a 4 partes.