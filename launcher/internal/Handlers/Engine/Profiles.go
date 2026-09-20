package engine

import (
	lprofile "StepLauncher/internal/Core/Launcher/Profile"
)

type Profile = lprofile.Profile

func (e *Engine) ListProfiles() map[string]*Profile {
	return e.profiles.List()
}

func (e *Engine) GetProfile(name string) (*Profile, error) {
	return e.profiles.Get(name)
}

func (e *Engine) CreateProfile(p *Profile) error {
	return e.profiles.Create(p)
}

func (e *Engine) UpdateProfile(name string, p *Profile) error {
	return e.profiles.Update(name, p)
}

func (e *Engine) DeleteProfile(name string) error {
	return e.profiles.Delete(name)
}

func (e *Engine) GetSelectedProfile() string {
	return e.profiles.Selected()
}

func (e *Engine) SetSelectedProfile(name string) error {
	return e.profiles.SetSelected(name)
}

func (e *Engine) GetSelectedVersion() string {
	return e.profiles.SelectedVersion()
}

func (e *Engine) SetSelectedVersion(version string) error {
	return e.profiles.SetSelectedVersion(version)
}
