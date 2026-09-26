package instance

import (
	"os"
	"path/filepath"
	"testing"
)

func nuevoManagerPrueba(t *testing.T) *InstanceManager {
	t.Helper()
	dir := t.TempDir()
	m := NewManager(filepath.Join(dir, "instances"), filepath.Join(dir, "shared"))
	return m
}

func TestProvisioningOcultaYBloquea(t *testing.T) {
	m := nuevoManagerPrueba(t)
	if err := m.BeginProvisioning("Pack", "Pack Lindo", "https://ejemplo/icono.png", "mp-1"); err != nil {
		t.Fatalf("begin: %v", err)
	}
	if !m.IsProvisioning("Pack") {
		t.Fatal("debería estar en creación")
	}
	if err := m.BeginProvisioning("Pack", "", "", "mp-2"); err == nil {
		t.Fatal("doble reserva debería fallar")
	}
	// Oculta de List y Get.
	if len(m.List()) != 0 {
		t.Fatal("List no debe mostrarla")
	}
	if _, _, err := m.Get("Pack"); err == nil {
		t.Fatal("Get no debe devolverla")
	}
	// Bloqueada: lanzar, borrar, descargar, metadatos.
	if err := m.assertUsable("Pack"); err == nil {
		t.Fatal("assertUsable debería rechazarla")
	}
	if err := m.Delete("Pack"); err == nil {
		t.Fatal("Delete debería rechazarla")
	}
	// El flujo interno sí puede trabajar con variantes System.
	if _, err := m.GameDirForSystem("Pack"); err == nil {
		t.Fatal("GameDirForSystem debería fallar sin metadata (aún no creada)")
	}
	// Visible tras liberar.
	m.EndProvisioning("Pack")
	if m.IsProvisioning("Pack") {
		t.Fatal("debería estar liberada")
	}
	if len(m.ListProvisioning()) != 0 {
		t.Fatal("sin marcas pendientes")
	}
}

func TestProvisioningListaYCancelacion(t *testing.T) {
	m := nuevoManagerPrueba(t)
	if _, err := os.Stat(filepath.Join("noexiste")); !os.IsNotExist(err) {
		t.Fatal("sanity")
	}
	if err := m.BeginProvisioning("A", "Pack A", "", "s1"); err != nil {
		t.Fatal(err)
	}
	if err := m.BeginProvisioning("B", "Pack B", "", "s2"); err != nil {
		t.Fatal(err)
	}
	lista := m.ListProvisioning()
	if len(lista) != 2 {
		t.Fatalf("marcas: %d", len(lista))
	}
	// Cancelación: borra el directorio a medias y libera.
	instPath := filepath.Join(m.instancesDir, "A")
	if err := os.MkdirAll(filepath.Join(instPath, "game", "mods"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := m.DeleteProvisioned("A"); err != nil {
		t.Fatalf("limpieza: %v", err)
	}
	if _, err := os.Stat(instPath); !os.IsNotExist(err) {
		t.Fatal("el directorio debería estar borrado")
	}
	if m.IsProvisioning("A") || !m.IsProvisioning("B") {
		t.Fatal("solo B debe seguir marcada")
	}
}
