package mods

import (
	"archive/zip"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestContentSubdir(t *testing.T) {
	casos := map[string]string{
		"mod": "mods", "resourcepack": "resourcepacks",
		"shader": "shaderpacks", "modpack": "modpacks",
	}
	for tipo, esperado := range casos {
		sub, err := ContentSubdir(tipo)
		if err != nil {
			t.Fatalf("tipo %s: %v", tipo, err)
		}
		if sub != esperado {
			t.Fatalf("tipo %s: esperado %s, obtenido %s", tipo, esperado, sub)
		}
	}
	if _, err := ContentSubdir("plugin"); err == nil {
		t.Fatal("se esperaba error para tipo desconocido")
	}
}

func TestResolveGameDir(t *testing.T) {
	if got := ResolveGameDir("/base", true); got != filepath.Join("/base", "game") {
		t.Fatalf("separada: %s", got)
	}
	if got := ResolveGameDir("/base", false); got != "/base" {
		t.Fatalf("base: %s", got)
	}
}

func TestSanitizeFileName(t *testing.T) {
	if _, err := SanitizeFileName("../fuera.jar"); err == nil {
		t.Fatal("se esperaba error para path traversal")
	}
	if _, err := SanitizeFileName(""); err == nil {
		t.Fatal("se esperaba error para nombre vacío")
	}
	if got, err := SanitizeFileName("sodium-1.0.jar"); err != nil || got != "sodium-1.0.jar" {
		t.Fatalf("nombre válido: %q %v", got, err)
	}
}

func TestSanitizeContentPath(t *testing.T) {
	if _, err := SanitizeContentPath("../../mods/fuera.jar"); err == nil {
		t.Fatal("se esperaba error para escape del gameDir")
	}
	if got, err := SanitizeContentPath("mods/sodium.jar"); err != nil || got != filepath.Join("mods", "sodium.jar") {
		t.Fatalf("path válido: %q %v", got, err)
	}
}

func TestDownloadToFileConHash(t *testing.T) {
	contenido := []byte("contenido-de-mod-de-prueba")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(contenido)
	}))
	defer srv.Close()

	dir := t.TempDir()
	dest := filepath.Join(dir, "mods", "prueba.jar")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	// Sin hash: descarga directa.
	if err := DownloadToFile(ctx, srv.Client(), srv.URL+"/prueba.jar", dest, "", "", 0); err != nil {
		t.Fatalf("descarga: %v", err)
	}
	// Segunda llamada con el mismo tamaño: se omite sin red.
	if err := DownloadToFile(ctx, srv.Client(), srv.URL+"/prueba.jar", dest, "", "", int64(len(contenido))); err != nil {
		t.Fatalf("omisión: %v", err)
	}
	// Hash incorrecto: debe fallar.
	if err := DownloadToFile(ctx, srv.Client(), srv.URL+"/prueba.jar", filepath.Join(dir, "otro.jar"), "00", "", 0); err == nil {
		t.Fatal("se esperaba error de hash")
	}
}

func TestExtractMrpackYOverrides(t *testing.T) {
	dir := t.TempDir()
	mrpack := filepath.Join(dir, "pack.mrpack")
	zw, err := os.Create(mrpack)
	if err != nil {
		t.Fatal(err)
	}
	w := zip.NewWriter(zw)
	agregar := func(nombre, contenido string) {
		fw, err := w.Create(nombre)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := fw.Write([]byte(contenido)); err != nil {
			t.Fatal(err)
		}
	}
	agregar("modrinth.index.json", `{"formatVersion":1,"game":"minecraft","versionId":"1.0","name":"Pack","files":[{"path":"mods/a.jar","hashes":{},"downloads":["https://ejemplo/a.jar"],"fileSize":3}],"dependencies":{"minecraft":"1.21","fabric-loader":"0.16.0"}}`)
	agregar("overrides/config/opciones.txt", "opciones")
	agregar("overrides/resourcepacks/pack.zip", "zip")
	agregar("../fuera.txt", "malicioso")
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	zw.Close()

	extraido := filepath.Join(dir, "extraido")
	if err := ExtractMrpack(mrpack, extraido); err != nil {
		t.Fatalf("extraer: %v", err)
	}
	if _, err := os.Stat(filepath.Join(extraido, "fuera.txt")); !os.IsNotExist(err) {
		t.Fatal("el zip con escape no debió escribir fuera")
	}
	index, err := ParseMrpackIndex(extraido)
	if err != nil {
		t.Fatalf("manifiesto: %v", err)
	}
	if index.MinecraftVersion() != "1.21" {
		t.Fatalf("mc: %s", index.MinecraftVersion())
	}
	loader, version := index.RequiredLoader()
	if loader != "fabric" || version != "0.16.0" {
		t.Fatalf("loader: %s %s", loader, version)
	}
	gameDir := filepath.Join(dir, "game")
	n, err := CopyOverrides(extraido, gameDir)
	if err != nil {
		t.Fatalf("overrides: %v", err)
	}
	if n != 2 {
		t.Fatalf("overrides copiados: %d", n)
	}
	if _, err := os.Stat(filepath.Join(gameDir, "config", "opciones.txt")); err != nil {
		t.Fatalf("override config: %v", err)
	}
}

func TestEnsureContentDirs(t *testing.T) {
	gameDir := filepath.Join(t.TempDir(), "game")
	if err := EnsureContentDirs(gameDir); err != nil {
		t.Fatalf("crear dirs: %v", err)
	}
	for _, sub := range []string{"", "mods", "resourcepacks", "shaderpacks"} {
		if info, err := os.Stat(filepath.Join(gameDir, sub)); err != nil || !info.IsDir() {
			t.Fatalf("falta %q: %v", sub, err)
		}
	}
}

func TestGestionInstalados(t *testing.T) {
	gameDir := t.TempDir()
	if err := EnsureContentDirs(gameDir); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(gameDir, "mods", "sodio.jar"), []byte("jar"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(gameDir, "shaderpacks", "apagado.zip.disabled"), []byte("zip"), 0644); err != nil {
		t.Fatal(err)
	}
	files, err := ListInstalled(gameDir)
	if err != nil {
		t.Fatalf("listar: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("instalados: %d", len(files))
	}
	if !files[0].Enabled || files[0].Name != "sodio.jar" {
		t.Fatalf("mod activo: %+v", files[0])
	}
	if files[1].Enabled || files[1].Name != "apagado.zip" {
		t.Fatalf("shader apagado: %+v", files[1])
	}
	if err := SetInstalledEnabled(gameDir, "mod", "sodio.jar", false); err != nil {
		t.Fatalf("deshabilitar: %v", err)
	}
	if _, err := os.Stat(filepath.Join(gameDir, "mods", "sodio.jar.disabled")); err != nil {
		t.Fatalf("rename disabled: %v", err)
	}
	if err := SetInstalledEnabled(gameDir, "mod", "sodio.jar", true); err != nil {
		t.Fatalf("habilitar: %v", err)
	}
	if err := DeleteInstalledContent(gameDir, "shader", "apagado.zip"); err != nil {
		t.Fatalf("borrar: %v", err)
	}
	if err := DeleteInstalledContent(gameDir, "mod", "inexistente.jar"); err == nil {
		t.Fatal("se esperaba error al borrar lo inexistente")
	}
}

func TestDownloadIcon(t *testing.T) {
	// PNG mínimo de 1x1 para que DetectContentType lo reconozca.
	png := []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A, 0, 0, 0, 0}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		w.Write(png)
	}))
	defer srv.Close()
	final, err := DownloadIcon(srv.Client(), srv.URL+"/icono", filepath.Join(t.TempDir(), "icon"))
	if err != nil {
		t.Fatalf("icono: %v", err)
	}
	if filepath.Ext(final) != ".png" {
		t.Fatalf("extensión: %s", final)
	}
	if _, err := os.Stat(final); err != nil {
		t.Fatalf("archivo icono: %v", err)
	}
	if _, err := DownloadIcon(srv.Client(), "ftp://invalido/icono", filepath.Join(t.TempDir(), "x")); err == nil {
		t.Fatal("se esperaba error con esquema inválido")
	}
}

func escribirZip(t *testing.T, path string, archivos map[string]string) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	w := zip.NewWriter(f)
	for nombre, contenido := range archivos {
		fw, err := w.Create(nombre)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := fw.Write([]byte(contenido)); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	f.Close()
}

func TestIconoPackPNG(t *testing.T) {
	dir := t.TempDir()
	gameDir := filepath.Join(dir, "game")
	if err := EnsureContentDirs(gameDir); err != nil {
		t.Fatal(err)
	}
	escribirZip(t, filepath.Join(gameDir, "resourcepacks", "pack.zip"), map[string]string{
		"pack.png":   "PNG-DATOS",
		"pack.mcmeta": "{}",
	})
	rel, err := ExtractContentIcon(gameDir, "resourcepack", "pack.zip", filepath.Join(dir, "cache"))
	if err != nil {
		t.Fatalf("icono: %v", err)
	}
	if rel == "" || filepath.Ext(rel) != ".png" {
		t.Fatalf("rel: %q", rel)
	}
}

func TestIconoFabricMod(t *testing.T) {
	dir := t.TempDir()
	gameDir := filepath.Join(dir, "game")
	if err := EnsureContentDirs(gameDir); err != nil {
		t.Fatal(err)
	}
	escribirZip(t, filepath.Join(gameDir, "mods", "mod.jar"), map[string]string{
		"fabric.mod.json":      `{"id":"mod","icon":"assets/mod/icon.png"}`,
		"assets/mod/icon.png": "PNG-DATOS",
	})
	rel, err := ExtractContentIcon(gameDir, "mod", "mod.jar", filepath.Join(dir, "cache"))
	if err != nil || rel == "" {
		t.Fatalf("icono fabric: %q %v", rel, err)
	}
}

func TestIconoForgeMod(t *testing.T) {
	dir := t.TempDir()
	gameDir := filepath.Join(dir, "game")
	if err := EnsureContentDirs(gameDir); err != nil {
		t.Fatal(err)
	}
	escribirZip(t, filepath.Join(gameDir, "mods", "forge.jar"), map[string]string{
		"META-INF/mods.toml": "logoFile=\"logo.png\"\n",
		"logo.png":           "PNG-DATOS",
	})
	rel, err := ExtractContentIcon(gameDir, "mod", "forge.jar", filepath.Join(dir, "cache"))
	if err != nil || rel == "" {
		t.Fatalf("icono forge: %q %v", rel, err)
	}
}

func TestIconoAusente(t *testing.T) {
	dir := t.TempDir()
	gameDir := filepath.Join(dir, "game")
	if err := EnsureContentDirs(gameDir); err != nil {
		t.Fatal(err)
	}
	escribirZip(t, filepath.Join(gameDir, "mods", "plano.jar"), map[string]string{
		"fabric.mod.json": `{"id":"plano"}`,
	})
	rel, err := ExtractContentIcon(gameDir, "mod", "plano.jar", filepath.Join(dir, "cache"))
	if err != nil || rel != "" {
		t.Fatalf("sin icono: %q %v", rel, err)
	}
	if _, err := ExtractContentIcon(gameDir, "mod", "noexiste.jar", filepath.Join(dir, "cache")); err != nil {
		t.Fatalf("inexistente no debe fallar: %v", err)
	}
}

func TestIconoNeoforgeToml(t *testing.T) {
	dir := t.TempDir()
	gameDir := filepath.Join(dir, "game")
	if err := EnsureContentDirs(gameDir); err != nil {
		t.Fatal(err)
	}
	// Nombre nuevo de NeoForge + logo + displayName (caso JEI, pero con logo).
	escribirZip(t, filepath.Join(gameDir, "mods", "neo.jar"), map[string]string{
		"META-INF/neoforge.mods.toml": "modId=\"neo\"\ndisplayName=\"Neo Lindo\"\nlogoFile=\"logo.png\"\n",
		"logo.png":                    "PNG-DATOS",
	})
	rel, err := ExtractContentIcon(gameDir, "mod", "neo.jar", filepath.Join(dir, "cache"))
	if err != nil || rel == "" {
		t.Fatalf("icono neoforge: %q %v", rel, err)
	}
	if title := ContentDisplayName(gameDir, "mod", "neo.jar"); title != "Neo Lindo" {
		t.Fatalf("título: %q", title)
	}
	if title := ContentDisplayName(gameDir, "mod", "plano.jar"); title != "" {
		t.Fatalf("sin título: %q", title)
	}
}

func TestNormalizaVersionForge(t *testing.T) {
	// Manifiesto corto (formato Modrinth) -> completa de Maven.
	index := &MrpackIndex{Dependencies: map[string]string{"minecraft": "1.16.5", "forge": "36.2.34"}}
	loader, version := index.RequiredLoader()
	if loader != "forge" || version != "1.16.5-36.2.34" {
		t.Fatalf("normalizado: %s %s", loader, version)
	}
	// Ya completa: no se toca.
	index = &MrpackIndex{Dependencies: map[string]string{"minecraft": "1.20.1", "forge": "1.20.1-47.2.0"}}
	_, version = index.RequiredLoader()
	if version != "1.20.1-47.2.0" {
		t.Fatalf("completa: %s", version)
	}
	// Otros loaders no se tocan.
	if v := NormalizeLoaderVersion("fabric-loader", "0.16.0", "1.21"); v != "0.16.0" {
		t.Fatalf("fabric: %s", v)
	}
	if v := NormalizeLoaderVersion("neoforge", "21.1.209", "1.21.1"); v != "21.1.209" {
		t.Fatalf("neoforge: %s", v)
	}
}

func TestIconoPorNombreCache(t *testing.T) {
	dir := t.TempDir()
	gameDir := filepath.Join(dir, "game")
	cacheDir := filepath.Join(dir, "cache")
	if err := EnsureContentDirs(gameDir); err != nil {
		t.Fatal(err)
	}
	// Jar sin metadatos (caso Essential): solo la caché por nombre lo viste.
	escribirZip(t, filepath.Join(gameDir, "mods", "esencial.jar"), map[string]string{
		"META-INF/MANIFEST.MF": "Manifest-Version: 1.0\n",
	})
	if rel, _ := ExtractContentIcon(gameDir, "mod", "esencial.jar", cacheDir); rel != "" {
		t.Fatalf("sin caché aún: %q", rel)
	}
	png := []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A, 0, 0, 0, 0}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		w.Write(png)
	}))
	defer srv.Close()
	CacheProjectIcon(cacheDir, "esencial.jar", srv.URL+"/icono")
	rel, err := ExtractContentIcon(gameDir, "mod", "esencial.jar", cacheDir)
	if err != nil || rel == "" {
		t.Fatalf("icono por nombre: %q %v", rel, err)
	}
}

func TestMoverInstalado(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "origen")
	dst := filepath.Join(dir, "destino")
	if err := EnsureContentDirs(src); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "mods", "mover.jar"), []byte("jar"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := MoveInstalledContent(src, dst, "mod", "mover.jar"); err != nil {
		t.Fatalf("mover: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dst, "mods", "mover.jar")); err != nil {
		t.Fatalf("destino: %v", err)
	}
	if _, err := os.Stat(filepath.Join(src, "mods", "mover.jar")); !os.IsNotExist(err) {
		t.Fatal("el origen debería estar vacío")
	}
	if err := MoveInstalledContent(src, dst, "mod", "mover.jar"); err == nil {
		t.Fatal("mover lo inexistente debería fallar")
	}
	if path, err := InstalledPath(dst, "mod", "mover.jar"); err != nil || path == "" {
		t.Fatalf("ruta: %q %v", path, err)
	}
}
