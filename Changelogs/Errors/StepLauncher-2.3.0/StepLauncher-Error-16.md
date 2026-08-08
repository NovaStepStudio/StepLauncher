# StepLauncher-Error-16: Forge/NeoForge modernos fallaban al descargar el jar "client" (URL vacía) → crash `Could not find .forge_patched_minecraft`

- **Fecha**: 2026-08-06
- **Estado**: corregido
- **Fixed?**: Sí — corregido y mencionado en la release `StepLauncher-2.3.0`.
- **Versión afectada**: 2.3.0 (productVersion de `wails.json`)

---

## 1. Síntoma

Al lanzar Forge 26.2 (MC 1.21.x) el juego crasheaba al instante con `exit 1`:

```
java.lang.IllegalStateException: Could not find .forge_patched_minecraft in classloader SecureModuleClassLoader[SECURE-BOOTSTRAP]
```

El log de preparación mostraba:

```
Downloading [2/37] net.minecraftforge:forge:26.2-65.1.0:client (75.5 MB)
download failed after 3 retries: HTTP 404
WARN: 1/37 libraries could not be downloaded, continuing anyway
```

## 2. Causa raíz

En `%APPDATA%\.StepLauncher\versions\26.2-forge-65.1.0\26.2-forge-65.1.0.json`, la librería patcheada `net.minecraftforge:forge:26.2-65.1.0:client` declaraba en `downloads.artifact` el `path`, `sha1` y `size` correctos pero con `"url": ""`. El instalador de Forge no publica URL para ese artifact (vive solo en su maven).

El fallback de URL del launcher apuntaba SIEMPRE a `https://libraries.minecraft.net`, que no sirve artifacts `net/minecraftforge/...` → HTTP 404. El jar patcheado (79 MB) nunca se descargaba, faltaba en el classpath y Forge no podía encontrar la versión parcheada de Minecraft.

## 3. Solución aplicada

Se centralizó en `internal/Core/Downloader/Helpers.go`:
- `repoForgeMaven = "https://maven.minecraftforge.net"` y `repoMinecraftLibraries = "https://libraries.minecraft.net"`.
- `IsForgeGroup(group)`: grupos del ecosistema Forge/NeoForge (`net.minecraftforge`, `cpw.mods`, `org.modlauncher`, `org.spongepowered`, `io.github.llamalad7`, `org.jline`, `org.ow2.asm`).
- `LibraryRepositoryBase(lib)` → maven de Forge para esos grupos, librería oficial para el resto.
- `HasRepositoryFallback(lib)`: solo se cae al repositorio cuando hay URL propia o el grupo es Forge (no se inventan tareas para librerías sin repo).

Se aplicó en los dos puntos que resuelven descargas:
- `internal/Core/Downloader/Tasks.go` (`addLibraryTasks`): ahora acepta artifact con `url==""` pero `path!=""` (resuelve `base + path` conservando SHA1/size) y el fallback por grupo.
- `internal/Core/Launcher/Helpers/Classpath.go` (`ResolveLibraryDownload`): artefacto con URL vacía → `LibraryRepositoryBase(name) + "/" + a.Path` manteniendo `SHA1`/`Size` para la verificación.

## 4. Verificación

- `go build ./...` OK.
- Con el fix, el `client` jar del Forge 26.2 se resuelve a `https://maven.minecraftforge.net/net/minecraftforge/forge/26.2-65.1.0/forge-26.2-65.1.0-client.jar` (con SHA1), se descarga y no se pierde `.forge_patched_minecraft`.
- Las librerías Mojang/LWJGL sin URL propia siguen cayendo a `libraries.minecraft.net` como antes.

## 5. Regla aprendida

Los version.json de Forge/NeoForge modernos dejan algún artifact SIN URL (`url: ""`) pero con `sha1`/`size`/`path`; hay que caer al maven correcto según el grupo Maven, nunca asumir `libraries.minecraft.net`, y para distinguir entre librerías que residen en un repo vs. las que no tienen repo (los niveles) hay fallback solo para los grupos con repositorio propio.