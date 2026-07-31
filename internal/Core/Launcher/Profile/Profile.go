package profile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Profile struct {
	Name             string            `json:"name"`
	Version          string            `json:"version,omitempty"`
	GameDir          string            `json:"gameDir,omitempty"`
	JavaExec         string            `json:"javaExec,omitempty"`
	JavaArgs         string            `json:"javaArgs,omitempty"`
	ResWidth         int               `json:"resWidth,omitempty"`
	ResHeight        int               `json:"resHeight,omitempty"`
	Fullscreen       bool              `json:"fullscreen"`
	ModLoader        string            `json:"modLoader,omitempty"`
	ModLoaderVer     string            `json:"modLoaderVersion,omitempty"`
	Icon             string            `json:"icon,omitempty"`
	CreatedAt        string            `json:"createdAt"`
	LastUsed         string            `json:"lastUsed,omitempty"`
	CustomProperties map[string]string `json:"customProperties,omitempty"`
}

type ProfilesFile struct {
	Profiles        map[string]*Profile `json:"profiles"`
	SelectedProfile string              `json:"selectedProfile,omitempty"`
}

type Manager struct {
	mu       sync.RWMutex
	filePath string
	data     ProfilesFile
}

func NewManager(workDir string) *Manager {
	return &Manager{
		filePath: filepath.Join(workDir, "launcher_profiles.json"),
		data: ProfilesFile{
			Profiles: make(map[string]*Profile),
		},
	}
}

func (m *Manager) Load() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, err := os.Stat(m.filePath); os.IsNotExist(err) {
		return nil
	}
	data, err := os.ReadFile(m.filePath)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, &m.data); err != nil {
		return fmt.Errorf("parse profiles: %w", err)
	}
	if m.data.Profiles == nil {
		m.data.Profiles = make(map[string]*Profile)
	}
	return nil
}

func (m *Manager) save() error {
	data, err := json.MarshalIndent(m.data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(m.filePath, data, 0644)
}

func (m *Manager) List() map[string]*Profile {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make(map[string]*Profile, len(m.data.Profiles))
	for k, v := range m.data.Profiles {
		cp := *v
		result[k] = &cp
	}
	return result
}

func (m *Manager) Get(name string) (*Profile, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.data.Profiles[name]
	if !ok {
		return nil, fmt.Errorf("profile %q not found", name)
	}
	cp := *p
	return &cp, nil
}

func (m *Manager) Create(p *Profile) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.data.Profiles[p.Name]; exists {
		return fmt.Errorf("profile %q already exists", p.Name)
	}
	if p.CreatedAt == "" {
		p.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	}
	m.data.Profiles[p.Name] = p
	return m.save()
}

func (m *Manager) Update(name string, p *Profile) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.data.Profiles[name]; !exists {
		return fmt.Errorf("profile %q not found", name)
	}
	p.Name = name
	m.data.Profiles[name] = p
	return m.save()
}

func (m *Manager) Delete(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.data.Profiles[name]; !exists {
		return fmt.Errorf("profile %q not found", name)
	}
	delete(m.data.Profiles, name)
	if m.data.SelectedProfile == name {
		m.data.SelectedProfile = ""
	}
	return m.save()
}

func (m *Manager) Selected() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.data.SelectedProfile
}

func (m *Manager) SetSelected(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if name != "" {
		if _, exists := m.data.Profiles[name]; !exists {
			return fmt.Errorf("profile %q not found", name)
		}
	}
	m.data.SelectedProfile = name
	return m.save()
}
