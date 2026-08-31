# Errores de StepLauncher 2.5.0 (Error-4) — Verificación antes de lanzar ignorada y fondos de galería duplicados por `UnixNano`

- **Fecha**: 2026-08-31
- **Versión**: 2.5.0 (en desarrollo)
- **Estado**: corregido
- **Release**: en desarrollo — aún no mencionado en ninguna release.

## Síntoma

Dos regresiones detectadas al auditar `internal/Config/Config.go` y `internal/Core/Assets/Assets.go` entre la base 2.3.1 (Wails v2) y el desarrollo 2.5.0:

1. **Verificación antes de lanzar no respetaba la configuración**: `LauncherConfig` no tenía campo `VerifyBeforeLaunch`, por lo que `internal/Core/Launcher/Manager.go:50` siempre verificaba o nunca verificaba según el flag global `VerifyIntegrity`, sin distinguir “verificar antes de lanzar” (sector configurable en `Settings/Sections/Integrity.vue`). Instalaciones antiguas sin campo mantenían comportamiento indefinido.

2. **Fondos de galería duplicados por URL**: `internal/Handlers/Gallery.go:58` generaba `fmt.Sprintf("%s_%d%s", baseName, time.Now().UnixNano(), ext)` y `RegisterGalleryBackground` solo deduplicaba por `Path`, nunca por `Url`. El reporte de `Changelogs/Bugs/StepLauncher-2.5.0/StepLauncher-Bug-1.md:15` mostraba 22 entradas en `launcher_assets.json.gallery` con misma `Url` (`35b1b4eb...`, `52b5e927b7...`) pero distinto `UnixNano` en `Path`, ocupando disco y JSON huérfano tras `cleanupBackgrounds`.

## Causa raíz

- `Config.go:61` carecía de `VerifyBeforeLaunch *bool` y helper `VerifyBeforeLaunchEnabled()` (default `true` si `nil`). `Manager.go` leía `VerifyIntegrity` genérico, no el flag específico.
- `Assets.go:413` `RegisterGalleryBackground` hacía `if exists.Path == new.Path { return }` sin comparar `Url`; `Gallery.go:58` no verificaba existencia por `Url` antes de `os.Create`. `referencedBackgrounds()` en `App.go:792` solo guardaba `Base`, no `clean` completo, por lo que `PruneOrphanGallery` no podía comparar.

## Diagnóstico y evidencia

- Inspección de `internal/Config/Config.go:61` → `VerifyBeforeLaunch *bool json:"verifyBeforeLaunch"` y `VerifyBeforeLaunchEnabled() bool`.
- Inspección de `internal/Core/Assets/Assets.go:413` → `if exists.Url == new.Url { removeWithRetry(oldPath) }`.
- Evidencia de usuario en `Bugs/2.5.0/Bug-1.md:12` con 22 entradas `gallery`, solo 1 archivo vivo `Solas Shader_Septonious_1787986530182608100.png`.

## Solución aplicada

1. **`internal/Config/Config.go:61-72`**: añadido `VerifyBeforeLaunch *bool json:"verifyBeforeLaunch"` + `VerifyBeforeLaunchEnabled() bool` (nil → true) + `floatPtr` helper. `Default()` y `sanitize()` lo registran.
2. **`internal/Core/Assets/Assets.go:413-554`**: `RegisterGalleryBackground` ahora deduplica por `Url` (si `exists.Url == urlNorm` borra viejo con `removeWithRetry` y elimina entrada), `PruneOrphanGallery` y `RemoveGalleryBackground` nuevos (ver `Changes/2.5.0/Change-10.md:21-30`).
3. **`internal/Handlers/App.go:792-830`**: `referencedBackgrounds()` ampliado a `clean` + `basename` + clave original; `updatePersonalizationInternal` + `pruneOrphanGallery()` en `Startup` y `ClearAllCache`.
4. **`internal/Handlers/Gallery.go:24-52`**: dedup por `Url` antes de descargar (si `assets.Load()` tiene `Url` y `os.Stat(full)` existe, reutiliza `Path` sin `os.Create`), log `[Gallery] Fondo reutilizado (deduplicado por URL)`.

## Regla aprendida

- **Todo flag de verificación debe ser `*bool` con helper `Enabled()` que defaultea a `true`** para compatibilidad con instalaciones antiguas sin campo (nil → true, no false). No usar `bool` plano que rompe `nil`.
- **Deduplicar galería por `Url`, no solo por `Path`**: `UnixNano` garantiza colisión si la misma URL se aplica dos veces; comparar `Url` evita `gallery` infinito y disco huérfano. `cleanupBackgrounds` debe tener su par `PruneOrphanGallery` para JSON.

## Verificación

- `go vet ./internal/Config ./internal/Core/Assets ./internal/Handlers` → 0 errores.
- Aplicar dos veces “Aplicar fondo” misma URL → `gallery` con 1 entrada, archivo reutilizado.
- `VerifyBeforeLaunch` nil → `Enabled()==true`, `false` → false.
