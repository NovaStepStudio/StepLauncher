# StepLauncher-Error-15: El mapeo de versiones de NeoForge a Minecraft era incorrecto (21.1.x → "1.21" en vez de "1.21.1") y el filtrado daba falsos positivos

- **Fecha**: 2026-08-06
- **Estado**: corregido
- **Fixed?**: Sí — corregido y mencionado en la release `StepLauncher-2.3.0`.
- **Versión afectada**: 2.3.0 (productVersion de `wails.json`)

---

## 1. Síntoma

Pedir versiones de NeoForge para una versión de MC concreta devolvía un conjunto erróneo:
- `NeoForgeVersionToMcVersion("21.1.x")` devolvía `"1.21"` (perdía el patch).
- El filtrado por prefijo (`strings.HasPrefix(v, "21")` para 1.21) incluía versiones de NeoForge de otras versiones de MC (p. ej. 21.5.x aparecía también al pedir 1.21.1).

La lista de loaders mostrada para una versión podía mezclar releases de NeoForge de Minecrafts distintos, y la etiqueta `MinecraftVersion` devuelta a la UI no siempre coincidía con la real.

## 2. Causa raíz

`internal/Core/ModLoader/Provider/Neoforge_resolver.go`:
- `NeoForgeVersionToMcVersion` usaba solo el primer segmento (`"1." + parts[0]`), ignorando el segundo, que es donde NeoForge codifica el patch de MC.
- `FilterVersionsForMinecraft` comparaba por prefijo del minor en vez de por el mapeo inverso.

Numeración real de NeoForge: `20.1.x → 1.20.1`, `20.4.x → 1.20.4`, `21.0.x → 1.21`, `21.1.x → 1.21.1`, `21.5.x → 1.21.5`. El segundo segmento es el patch de MC; un `0` significa MC sin patch (no existe "1.21.0").

## 3. Solución aplicada

`internal/Core/ModLoader/Provider/Neoforge_resolver.go`:
- Nueva función de paquete `NeoForgeVersionToMcVersion(neoVersion)`: `"1." + parts[0]` y, si el segundo segmento no es `"0"`, se añade `"." + parts[1]`. El método `(NeoForgeVersionResolver).NeoForgeVersionToMcVersion` delega en ella.
- `FilterVersionsForMinecraft` ahora filtra por comparación exacta `NeoForgeVersionToMcVersion(v) == mcVersion`, sin prefijos ambiguos.

## 4. Verificación

- `go build ./...` OK.
- Mapeos comprobados: `21.1.x → 1.21.1`, `21.0.x → 1.21`, `20.4.x → 1.20.4`, `20.1.x → 1.20.1`, `21.5.x → 1.21.5`.
- Pedir versiones para `1.21.1` solo devuelve releases `21.1.x`.

## 5. Regla aprendida

NeoForge (a diferencia de Forge) numera sus releases igual que MC pero sin el prefijo "1." y con el patch de MC como segundo segmento. No filtrar por prefijo de minor: usar el mapeo inverso exacto versión→MC y recordar que un segundo segmento `0` significa MC sin patch.