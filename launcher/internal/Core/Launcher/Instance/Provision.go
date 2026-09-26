package instance

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// Provision marca una instancia en creación por un modpack: mientras exista,
// la instancia NO aparece en List/Get (se muestra solo al 100%) y assertUsable
// bloquea lanzarla, modificarla o borrarla. El flujo interno del motor usa las
// variantes System de AddVersion/InstallModLoader/Update para trabajar con
// ella; al terminar (o al fallar/cancelar) se cierra con EndProvisioning (y
// DeleteProvisioned limpia el directorio a medias).
type Provision struct {
	Name      string
	Title     string
	IconURL   string
	SessionID string
	StartedAt string
}

// ProvisionInfo es la vista pública para las tarjetas "Creando…" del panel.
type ProvisionInfo struct {
	Name      string `json:"name"`
	Title     string `json:"title"`
	IconURL   string `json:"iconUrl"`
	SessionID string `json:"sessionId"`
	StartedAt string `json:"startedAt"`
}

// BeginProvisioning reserva el nombre para una instancia en creación.
func (m *InstanceManager) BeginProvisioning(name, title, iconURL, sessionID string) error {
	if err := sanitizeInstanceName(name); err != nil {
		return err
	}
	// La comprobación de existencia va sin lock (I/O fuera del mutex).
	if _, err := os.Stat(filepath.Join(m.instancesDir, name)); err == nil {
		return fmt.Errorf("la instancia %s ya existe", name)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.provisioning[name] != nil {
		return fmt.Errorf("la instancia %s ya se está creando", name)
	}
	if title == "" {
		title = name
	}
	m.provisioning[name] = &Provision{
		Name: name, Title: title, IconURL: iconURL, SessionID: sessionID,
		StartedAt: time.Now().Format(time.RFC3339),
	}
	return nil
}

// EndProvisioning libera la marca: la instancia pasa a ser visible y usable.
func (m *InstanceManager) EndProvisioning(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.provisioning, name)
}

// IsProvisioning indica si la instancia sigue en creación.
func (m *InstanceManager) IsProvisioning(name string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.provisioning[name] != nil
}

// ListProvisioning devuelve las creaciones en curso, ordenadas por inicio.
func (m *InstanceManager) ListProvisioning() []*ProvisionInfo {
	m.mu.RLock()
	list := make([]*ProvisionInfo, 0, len(m.provisioning))
	for _, p := range m.provisioning {
		list = append(list, &ProvisionInfo{
			Name: p.Name, Title: p.Title, IconURL: p.IconURL,
			SessionID: p.SessionID, StartedAt: p.StartedAt,
		})
	}
	m.mu.RUnlock()
	sort.Slice(list, func(i, j int) bool { return list[i].StartedAt < list[j].StartedAt })
	return list
}

// DeleteProvisioned borra sin condiciones una instancia en creación (fallo o
// cancelación del modpack): elimina su directorio y libera la marca.
func (m *InstanceManager) DeleteProvisioned(name string) error {
	if err := sanitizeInstanceName(name); err != nil {
		return err
	}
	// I/O fuera del mutex: primero se borra el directorio...
	instPath := filepath.Join(m.instancesDir, name)
	if _, err := os.Stat(instPath); err == nil {
		if err := os.RemoveAll(instPath); err != nil {
			return fmt.Errorf("limpiar %s: %w", name, err)
		}
	}
	// ...y luego se libera la marca.
	m.EndProvisioning(name)
	return nil
}
