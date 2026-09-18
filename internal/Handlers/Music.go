package Handlers

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	assets "StepLauncher/internal/Core/Assets"
)

// ReadMusicFile lee el archivo de audio `src` (ruta absoluta elegida con el
// diálogo) y devuelve sus bytes, para que el frontend extraiga metadatos y
// duración antes de importarlo. Solo acepta extensiones de audio <= 15 MB.
func (a *App) ReadMusicFile(src string) ([]byte, error) {
	if strings.TrimSpace(src) == "" {
		return nil, fmt.Errorf("no se seleccionó ningún archivo")
	}
	ext := strings.ToLower(filepath.Ext(src))
	if !assets.IsAudioExt(ext) {
		return nil, fmt.Errorf("formato de audio no soportado: %s (usa .mp3, .wav, .ogg o .m4a)", ext)
	}
	st, err := os.Stat(src)
	if err != nil {
		return nil, fmt.Errorf("no se pudo leer el archivo: %v", err)
	}
	if st.Size() > assets.MaxAudioBytes {
		return nil, fmt.Errorf("el audio no debe pesar más de 15 MB")
	}
	data, err := os.ReadFile(src)
	if err != nil {
		return nil, fmt.Errorf("no se pudo leer el archivo: %v", err)
	}
	return data, nil
}

// PickMusicFile abre el diálogo para elegir un archivo de audio de fondo
// (MP3/WAV/OGG/M4A) y devuelve la ruta absoluta seleccionada.
func (a *App) PickMusicFile() (string, error) {
	if a.runtime == nil {
		return "", fmt.Errorf("runtime no disponible")
	}
	return a.runtime.OpenFileDialog("Seleccionar música de fondo", []FileFilter{
		{DisplayName: "Audio (*.mp3, *.wav, *.ogg, *.m4a)", Pattern: "*.mp3;*.wav;*.ogg;*.m4a"},
	})
}

// ListMusic devuelve las pistas de música de fondo registradas en
// launcher_assets.json.
func (a *App) ListMusic() []assets.MusicSlot {
	if a.assets == nil {
		return []assets.MusicSlot{}
	}
	list, err := a.assets.ListMusic()
	if err != nil {
		a.logf("[Music] WARN: no se pudo listar la música de fondo: %v", err)
		return []assets.MusicSlot{}
	}
	return list
}

// AddMusic copia el archivo `src` (validado: extensión de audio, <= 15 MB,
// máx. 5 pistas) a cache/audio/ y lo registra en launcher_assets.json.
// Devuelve la ruta relativa registrada.
func (a *App) AddMusic(name, src string) (string, error) {
	if a.assets == nil {
		return "", fmt.Errorf("assets no disponible")
	}
	if strings.TrimSpace(src) == "" {
		return "", fmt.Errorf("no se seleccionó ningún archivo")
	}
	if filepath.IsAbs(src) == false {
		return "", fmt.Errorf("ruta de archivo no válida")
	}
	rel, err := a.assets.AddMusic(name, src)
	if err != nil {
		return "", err
	}
	a.logf("[Music] Audio de fondo añadido: %s -> %s", name, rel)
	return rel, nil
}

// RemoveMusic elimina la pista `name` de la música de fondo (JSON + archivo).
func (a *App) RemoveMusic(name string) error {
	if a.assets == nil {
		return fmt.Errorf("assets no disponible")
	}
	if err := a.assets.RemoveMusic(name); err != nil {
		return err
	}
	a.logf("[Music] Audio de fondo eliminado: %s", name)
	return nil
}

// PickCustomCoverFile abre el diálogo para elegir una imagen de carátula custom (PNG/JPG/WEBP)
func (a *App) PickCustomCoverFile() (string, error) {
	if a.runtime == nil {
		return "", fmt.Errorf("runtime no disponible")
	}
	return a.runtime.OpenFileDialog("Seleccionar carátula", []FileFilter{
		{DisplayName: "Imágenes (*.png, *.jpg, *.jpeg, *.webp)", Pattern: "*.png;*.jpg;*.jpeg;*.webp"},
	})
}